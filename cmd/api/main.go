package main

import (
	"context"
	"log"
	"net/http"
	"time"

	"gagarin-soft/internal/auth"
	"gagarin-soft/internal/config"
	"gagarin-soft/internal/handlers"
	"gagarin-soft/internal/services"
	"gagarin-soft/internal/storage"
)

func main() {
	// 1. Load Config
	cfg := config.Load()

	// 2. Initialize Auth Manager
	ctx := context.Background()
	var authManager auth.TokenManager
	var err error

	if cfg.AppEnv == "local" {
		log.Println("Initializing MOCK Auth Manager (local env)")
		authManager = auth.NewMockManager()
	} else {
		log.Println("Initializing Google Cloud Auth Manager")
		authManager, err = auth.NewGoogleManager(ctx, cfg.ProjectID, cfg.OAuthClientID, cfg.OAuthClientSecret)
		if err != nil {
			log.Fatalf("Failed to initialize auth manager: %v", err)
		}
		defer authManager.Close()
	}

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

	// 4. Initialize Services and Handlers
	gmailService := services.NewGmailWatchService(cfg, authManager, repo)
	renewHandler := &handlers.RenewWatchHandler{Service: gmailService}

	// 5. Define Handlers
	mux := http.NewServeMux()

	mux.HandleFunc("POST /health", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("OK"))
	})

	mux.Handle("POST /renew-watch", renewHandler)

	// 6. Start Server
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
