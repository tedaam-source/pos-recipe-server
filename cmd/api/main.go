package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"time"

	"gagarin-soft/internal/auth"
	"gagarin-soft/internal/config"
	"gagarin-soft/internal/gmail"
	"gagarin-soft/internal/storage"
)

func main() {
	// 1. Load Config
	cfg := config.Load()

	// 2. Initialize Auth Manager (Secret Manager)
	ctx := context.Background()
	authManager, err := auth.NewManager(ctx, cfg.ProjectID, cfg.OAuthClientID, cfg.OAuthClientSecret)
	if err != nil {
		log.Fatalf("Failed to initialize auth manager: %v", err)
	}
	defer authManager.Close()

	// 3. Initialize Storage
	var repo storage.HistoryRepository
	if cfg.InstanceConnectionName != "" {
		log.Printf("Initializing Cloud SQL storage...")
		postgresRepo, cleanup, err := storage.NewPostgresRepository(
			ctx,
			cfg.InstanceConnectionName,
			cfg.DBUser,
			cfg.DBPass,
			cfg.DBName,
		)
		if err != nil {
			log.Fatalf("Failed to initialize Cloud SQL: %v", err)
		}
		defer func() {
			if err := cleanup(); err != nil {
				log.Printf("Failed to close Cloud SQL dialer: %v", err)
			}
		}()
		repo = postgresRepo
	} else {
		log.Println("Using No-Op storage (no DB configured)")
		repo = &storage.NoOpRepository{}
	}

	// 4. Define Handlers
	mux := http.NewServeMux()

	mux.HandleFunc("POST /health", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("OK"))
	})

	mux.HandleFunc("POST /renew-watch", func(w http.ResponseWriter, r *http.Request) {
		handleRenewWatch(w, r, cfg, authManager, repo)
	})

	// 5. Start Server
	log.Printf("Starting server on :%s", cfg.Port)
	server := &http.Server{
		Addr:         ":" + cfg.Port,
		Handler:      mux,
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 30 * time.Second,
	}

	if err := server.ListenAndServe(); err != nil {
		log.Fatalf("Server failed to start: %v", err)
	}
}

type RenewWatchRequest struct {
	TopicName string `json:"topicName"` // Optional override
}

func handleRenewWatch(w http.ResponseWriter, r *http.Request, cfg *config.Config, authMgr *auth.Manager, repo storage.HistoryRepository) {
	ctx, cancel := context.WithTimeout(r.Context(), 30*time.Second)
	defer cancel()

	// 1. Parse Request (Optional topic override)
	// Default topic: projects/<projectID>/topics/gmail-hook-topic
	// Note: In a real scenario, we might want to validate this strictly.
	topicName := fmt.Sprintf("projects/%s/topics/gmail-hook-topic", cfg.ProjectID)

	var reqBody RenewWatchRequest
	if r.Body != nil {
		_ = json.NewDecoder(r.Body).Decode(&reqBody)
		if reqBody.TopicName != "" {
			topicName = reqBody.TopicName
		}
	}

	// 2. Get Refresh Token
	// Secret Name: gmail-refresh-token
	refreshToken, err := authMgr.GetRefreshToken(ctx, "gmail-refresh-token")
	if err != nil {
		log.Printf("Error retrieving refresh token: %v", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	// 3. Get Authenticated HTTP Client
	client := authMgr.GetHTTPClient(ctx, refreshToken)

	// 4. Create Gmail Client
	gmailClient, err := gmail.NewClient(ctx, client)
	if err != nil {
		log.Printf("Error creating Gmail client: %v", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	// 5. Call Renew Watch
	log.Printf("Renewing watch for topic: %s", topicName)
	resp, err := gmailClient.RenewWatch(topicName)
	if err != nil {
		log.Printf("Error renewing watch: %v", err)
		http.Error(w, fmt.Sprintf("Failed to renew watch: %v", err), http.StatusInternalServerError)
		return
	}

	// 6. Log & Save Results
	log.Printf("Successfully renewed watch. HistoryID: %d, Expiration: %d", resp.HistoryId, resp.Expiration)

	if err := repo.SaveWatchStatus(ctx, resp.HistoryId, resp.Expiration); err != nil {
		log.Printf("Warning: Failed to save watch status: %v", err)
		// We don't fail the request here as the primary action succeeded
	}

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(resp)
}
