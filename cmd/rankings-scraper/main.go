// Command rankings-scraper runs a single scrape of https://www.ufc.com/rankings,
// storing both the media-panel and META (AI) rankings tables in the database.
// The running server (cmd/server) also scrapes on its own schedule; this
// binary is for manual/on-demand runs, e.g. from a cron job or CI.
package main

import (
	"context"
	"log"

	"github.com/dancouver1/ufc-card-creator/internal/config"
	"github.com/dancouver1/ufc-card-creator/internal/db"
	"github.com/dancouver1/ufc-card-creator/internal/domain/repository"
	"github.com/dancouver1/ufc-card-creator/internal/rankings"
)

func main() {
	cfg := config.Load()
	database, err := db.NewDB(cfg.DatabaseURL)
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}
	defer database.Close()

	repo := repository.New(database.Pool)

	if err := rankings.Run(context.Background(), repo); err != nil {
		log.Fatalf("Scrape failed: %v", err)
	}
}
