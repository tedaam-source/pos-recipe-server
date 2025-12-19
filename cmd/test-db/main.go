package main

import (
	"context"
	"log"

	"gagarin-soft/internal/config"
	"gagarin-soft/internal/storage"

	"github.com/joho/godotenv"
)

func main() {
	// Load .env
	if err := godotenv.Load(); err != nil {
		log.Println("Warning: No .env file found (or error loading it)")
	}
	cfg := config.Load()
	log.Printf("Testing connection to: %s User: %s", cfg.InstanceConnectionName, cfg.DBUser)

	_, cleanup, err := storage.NewPostgresRepository(
		context.Background(),
		cfg.InstanceConnectionName,
		cfg.DBUser,
		cfg.DBPass,
		cfg.DBName,
	)
	if err != nil {
		log.Fatalf("❌ Connection FAILED: %v\nHint: Ensure you have run 'gcloud auth application-default login' if running locally.", err)
	}
	defer func() {
		if cleanup != nil {
			if err := cleanup(); err != nil {
				log.Printf("Cleanup error: %v", err)
			}
		}
	}()

	log.Println("✅ Connection SUCCESSFUL!")
}
