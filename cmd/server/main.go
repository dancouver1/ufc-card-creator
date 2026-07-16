package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/dancouver1/ufc-card-creator/internal/config"
	"github.com/dancouver1/ufc-card-creator/internal/db"
	"github.com/dancouver1/ufc-card-creator/internal/domain/repository"
	"github.com/dancouver1/ufc-card-creator/internal/rankings"
	transporthttp "github.com/dancouver1/ufc-card-creator/internal/transport/http"
	"github.com/dancouver1/ufc-card-creator/internal/transport/http/handlers"

	_ "github.com/dancouver1/ufc-card-creator/docs"
)

// @title UFC Card Creator API
// @version 1.0
// @description Backend API for browsing UFC fighters and building matchmaker cards.
// @BasePath /api
func main() {
	cfg := config.Load()

	// Initialize database connection
	database, err := db.NewDB(cfg.DatabaseURL)
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}
	defer database.Close()

	repo := repository.New(database.Pool)
	h := handlers.NewHandler(repo)
	r := transporthttp.NewRouter(h, database.Health)

	// Keep the rankings page in sync with ufc.com/rankings by re-scraping on
	// a schedule (ufc.com offers no update webhook, so periodic polling is
	// the closest available option).
	schedulerCtx, stopScheduler := context.WithCancel(context.Background())
	defer stopScheduler()
	rankings.StartScheduler(schedulerCtx, repo, cfg.RankingsScrapeInterval)

	// Create server
	server := &http.Server{
		Addr:         ":" + cfg.Port,
		Handler:      r,
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 15 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	// Start server in a goroutine
	go func() {
		log.Printf("Server starting on port %s...", cfg.Port)
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("Server failed to start: %v", err)
		}
	}()

	// Graceful shutdown
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	log.Println("Shutting down server...")

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	if err := server.Shutdown(ctx); err != nil {
		log.Fatalf("Server forced to shutdown: %v", err)
	}

	log.Println("Server exited")
}
