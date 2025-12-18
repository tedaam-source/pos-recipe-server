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
