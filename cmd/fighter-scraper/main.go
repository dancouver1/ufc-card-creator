// Command fighter-scraper scrapes the full UFC athlete roster from
// https://www.ufc.com/athletes/all and upserts active fighters into the
// database, skipping retired/cut fighters. Run cmd/image-scraper -all
// afterwards to backfill photos for newly added fighters.
package main

import (
	"context"
	"flag"
	"log"
	"time"

	"github.com/dancouver1/ufc-card-creator/internal/config"
	"github.com/dancouver1/ufc-card-creator/internal/db"
	"github.com/dancouver1/ufc-card-creator/internal/domain/repository"
	"github.com/dancouver1/ufc-card-creator/internal/fighterscraper"
)

func main() {
	delay := flag.Duration("delay", 750*time.Millisecond, "Delay between requests to ufc.com")
	maxPages := flag.Int("max-pages", 0, "Stop after this many listing pages (0 = scrape until the roster is exhausted)")
	dryRun := flag.Bool("dry-run", false, "Log active fighters instead of writing them to the database")
	flag.Parse()

	var repo *repository.Repository
	if !*dryRun {
		cfg := config.Load()
		database, err := db.NewDB(cfg.DatabaseURL)
		if err != nil {
			log.Fatalf("Failed to connect to database: %v", err)
		}
		defer database.Close()

		repo = repository.New(database.Pool)
	}

	opts := fighterscraper.Options{
		Delay:    *delay,
		MaxPages: *maxPages,
		DryRun:   *dryRun,
	}

	if _, err := fighterscraper.Run(context.Background(), repo, opts); err != nil {
		log.Fatalf("Scrape failed: %v", err)
	}
}
