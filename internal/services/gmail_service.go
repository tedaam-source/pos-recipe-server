package services

import (
	"context"
	"encoding/json"
	"fmt"
	"log"

	"gagarin-soft/internal/auth"
	"gagarin-soft/internal/config"
	"gagarin-soft/internal/gmail"
	"gagarin-soft/internal/storage"
)

type GmailWatchService struct {
	Config      *config.Config
	AuthManager auth.TokenManager
	Repo        storage.HistoryRepository
}

func NewGmailWatchService(cfg *config.Config, authMgr auth.TokenManager, repo storage.HistoryRepository) *GmailWatchService {
	return &GmailWatchService{
		Config:      cfg,
		AuthManager: authMgr,
		Repo:        repo,
	}
}

func (s *GmailWatchService) Renew(ctx context.Context) ([]byte, error) {
	// 1. Determine Topic
	topicName := s.Config.GmailPubSubTopic
	if topicName == "" {
		topicName = fmt.Sprintf("projects/%s/topics/gmail-hook-topic", s.Config.ProjectID)
	}

	// 2. Get Refresh Token
	refreshToken, err := s.AuthManager.GetRefreshToken(ctx, "gmail-refresh-token")
	if err != nil {
		log.Printf("Error retrieving refresh token: %v", err)
		return nil, fmt.Errorf("internal server error")
	}

	// 3. Get Authenticated HTTP Client
	client := s.AuthManager.GetHTTPClient(ctx, refreshToken)

	// 4. Create Gmail Client
	gmailClient, err := gmail.NewClient(ctx, client)
	if err != nil {
		log.Printf("Error creating Gmail client: %v", err)
		return nil, fmt.Errorf("internal server error")
	}

	// 5. Call Renew Watch
	log.Printf("Renewing watch for topic: %s", topicName)
	resp, err := gmailClient.RenewWatch(topicName)
	if err != nil {
		log.Printf("Error renewing watch: %v", err)
		return nil, fmt.Errorf("failed to renew watch: %w", err)
	}

	// 6. Log & Save Results
	log.Printf("Successfully renewed watch. HistoryID: %d, Expiration: %d", resp.HistoryId, resp.Expiration)

	if err := s.Repo.SaveWatchStatus(ctx, resp.HistoryId, resp.Expiration); err != nil {
		log.Printf("Warning: Failed to save watch status: %v", err)
	}

	return json.Marshal(resp)
}

func (s *GmailWatchService) ProcessPushNotification(ctx context.Context, startHistoryID uint64) error {
	// 1. Get Authenticated Client
	refreshToken, err := s.AuthManager.GetRefreshToken(ctx, "gmail-refresh-token")
	if err != nil {
		return fmt.Errorf("failed to get refresh token: %w", err)
	}
	client := s.AuthManager.GetHTTPClient(ctx, refreshToken)
	gmailClient, err := gmail.NewClient(ctx, client)
	if err != nil {
		return fmt.Errorf("failed to create gmail client: %w", err)
	}

	// 2. List History
	msgIDs, err := gmailClient.ListMessageIDs(startHistoryID)
	if err != nil {
		return fmt.Errorf("failed to list history: %w", err)
	}

	log.Printf("Found %d messages in history", len(msgIDs))

	// 3. Process Messages
	targetLabel := s.Config.TargetGmailLabel

	statsReceived := len(msgIDs)
	statsOk := 0
	statsError := 0

	for _, msgID := range msgIDs {
		msg, err := gmailClient.GetMessage(msgID)
		if err != nil {
			log.Printf("Failed to get message %s: %v", msgID, err)
			statsError++
			_ = s.Repo.RecordEvent(ctx, storage.Event{
				MessageID: msgID,
				Status:    "error",
				Error:     fmt.Sprintf("Failed to get message: %v", err),
			})
			continue
		}

		matched := false
		if targetLabel != "" {
			for _, label := range msg.LabelIds {
				if label == targetLabel {
					matched = true
					break
				}
			}
		} else {
			matched = true
		}

		if matched {
			log.Printf("Message %s matched label %s. Saving...", msgID, targetLabel)

			// Save using old logic
			processed := storage.ProcessedEmail{
				MessageID: msg.Id,
				HistoryID: msg.HistoryId,
				LabelIDs:  fmt.Sprintf("%v", msg.LabelIds),
				Snippet:   msg.Snippet,
			}
			if err := s.Repo.SaveProcessedEmail(ctx, processed); err != nil {
				log.Printf("Failed to save processed email: %v", err)
				statsError++
				_ = s.Repo.RecordEvent(ctx, storage.Event{
					MessageID: msg.Id,
					Status:    "error",
					Error:     fmt.Sprintf("Failed to save to db: %v", err),
				})
			} else {
				statsOk++
				// Record Success Event for Admin Dashboard
				_ = s.Repo.RecordEvent(ctx, storage.Event{
					MessageID: msg.Id,
					Status:    "processed",
					// Assuming we can't get Filter ID easily unless we re-fetch filters.
					// For now, leave FilterID empty or infer from label name if needed.
				})
			}
		} else {
			// Message didn't match, maybe log as filtered?
			// Existing logic just ignores it.
			// Admin dashboard might want to know about ignored messages too?
			// For now, stick to processed ones to avoid spamming events.
		}
	}

	// Update Daily Stats
	if statsReceived > 0 || statsOk > 0 || statsError > 0 {
		if err := s.Repo.UpdateDailyStats(ctx, statsReceived, statsOk, statsError); err != nil {
			log.Printf("Failed to update daily stats: %v", err)
		}
	}

	return nil
}
