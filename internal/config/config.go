package config

import (
	"os"

	"github.com/joho/godotenv"
)

// Config holds application configuration loaded from the environment.
type Config struct {
	Port        string
	DatabaseURL string
}

// Load reads a local .env file (if present) and returns the application config.
func Load() *Config {
	// Only works locally, not in Docker.
	_ = godotenv.Load()

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	return &Config{
		Port:        port,
		DatabaseURL: os.Getenv("DATABASE_URL"),
	}
}
