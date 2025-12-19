package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/go-chi/chi/v5"
	chimiddleware "github.com/go-chi/chi/v5/middleware"
	"github.com/joho/godotenv"

	"gagarin-soft/internal/admin/config"
	"gagarin-soft/internal/admin/handlers"
	"gagarin-soft/internal/admin/middleware"
	"gagarin-soft/internal/admin/storage"
)

func main() {
	_ = godotenv.Load() // Ignore error if .env doesn't exist
	cfg := config.Load()

	// Build DB Connection String
	var connString string
	if cfg.InstanceConnectionName != "" {
		// When using connector, host/port are ignored in DSN, user/pass/db matter
		// DSN format for pgx with connector often just needs user/pass/db
		connString = fmt.Sprintf("user=%s password=%s dbname=%s", cfg.DBUser, cfg.DBPass, cfg.DBName)
	} else {
		// Local TCP
		dbHost := os.Getenv("DB_HOST")
		if dbHost == "" {
			dbHost = "localhost"
		}
		connString = fmt.Sprintf("postgres://%s:%s@%s:5432/%s",
			cfg.DBUser, cfg.DBPass, dbHost, cfg.DBName)
	}

	ctx := context.Background()
	store, err := storage.New(ctx, connString, cfg.InstanceConnectionName)
	if err != nil {
		log.Fatalf("Failed to connect to storage: %v", err)
	}
	defer store.Close()

	h := handlers.NewHandler(cfg, store)
	iap := middleware.NewIAPMiddleware(cfg.AdminAllowlist, cfg.AppEnv)

	r := chi.NewRouter()
	r.Use(chimiddleware.Logger)
	r.Use(chimiddleware.Recoverer)
	r.Use(chimiddleware.Timeout(60 * time.Second))

	// Health check (Public/Internal)
	r.Post("/health", h.Health) // User requested POST but standard is GET... implementing POST as requested
	r.Get("/health", h.Health)  // Also support GET for convenience

	// Admin API Protected by IAP
	r.Group(func(r chi.Router) {
		r.Use(iap.Middleware)

		r.Route("/admin", func(r chi.Router) {
			r.Get("/stats", h.GetStats)

			r.Get("/filters", h.GetFilters)
			r.Post("/filters", h.CreateFilter)
			r.Patch("/filters/{id}", h.UpdateFilter)
			r.Delete("/filters/{id}", h.DeleteFilter)

			r.Get("/events", h.GetEvents)

			r.Post("/actions/{action}", h.TriggerAction) // renew-watch, resync, reprocess
		})
	})

	// TODO: Serve Frontend Static Files here if single container
	// For now, API only. We can add static file server later or user can deploy separately.
	// User said: "Can separate or single container".
	// Implementation Plan said: "Go server will serve the web/out"
	// Let's add simple file server for "dist" or "out" folder if it exists.

	fs := http.FileServer(http.Dir("./web/out"))
	// We only serve static files for paths NOT starting with /admin and NOT /health
	// But actually Next.js export produces .html files.
	// Simplest: Serve everything else as file server?

	// A wildcard handler for frontend - careful not to mask API
	r.Get("/*", func(w http.ResponseWriter, r *http.Request) {
		if _, err := os.Stat("./web/out" + r.URL.Path); err == nil {
			fs.ServeHTTP(w, r)
			return
		}
		// Fallback to index.html for client-side routing
		http.ServeFile(w, r, "./web/out/index.html")
	})

	log.Printf("Starting Admin Service on :%s", cfg.Port)
	if err := http.ListenAndServe(":"+cfg.Port, r); err != nil {
		log.Fatalf("Server failed: %v", err)
	}
}
