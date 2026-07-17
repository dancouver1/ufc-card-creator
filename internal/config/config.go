package config

import (
	"os"
	"strconv"
	"time"

	"github.com/joho/godotenv"
)

// Config holds application configuration loaded from the environment.
type Config struct {
	Port                   string
	DatabaseURL            string
	RankingsScrapeInterval time.Duration
}

// Load reads a local .env file (if present) and returns the application config.
func Load() *Config {
	// Only works locally, not in Docker.
	_ = godotenv.Load()

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	interval := 6 * time.Hour
	if raw := os.Getenv("RANKINGS_SCRAPE_INTERVAL_HOURS"); raw != "" {
		if hours, err := strconv.Atoi(raw); err == nil && hours > 0 {
			interval = time.Duration(hours) * time.Hour
		}
	}

	return &Config{
		Port:                   port,
		DatabaseURL:            os.Getenv("DATABASE_URL"),
		RankingsScrapeInterval: interval,
	}
}
