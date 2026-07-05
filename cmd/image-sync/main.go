package main

import (
	"context"
	"fmt"
	"io/ioutil"
	"log"
	"path/filepath"
	"strings"

	"github.com/dancouver1/ufc-card-creator/internal/config"
	"github.com/dancouver1/ufc-card-creator/internal/db"
	"github.com/dancouver1/ufc-card-creator/internal/domain/repository"
)

func main() {
	cfg := config.Load()

	// Connect to DB
	database, err := db.NewDB(cfg.DatabaseURL)
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}
	defer database.Close()

	repo := repository.New(database.Pool)

	imageDir := "./static/images/fighters"
	files, err := ioutil.ReadDir(imageDir)
	if err != nil {
		log.Fatalf("Failed to read image directory: %v", err)
	}

	ctx := context.Background()
	updatedCount := 0

	for _, file := range files {
		if file.IsDir() {
			continue
		}

		filename := file.Name()
		ext := filepath.Ext(filename)
		if ext != ".webp" && ext != ".png" && ext != ".jpg" && ext != ".jpeg" {
			continue
		}

		// Convert conor_mcgregor.webp -> Conor McGregor
		baseName := filename[:len(filename)-len(ext)]
		parts := strings.Split(baseName, "_")
		for i, part := range parts {
			parts[i] = strings.Title(part)
		}
		fighterName := strings.Join(parts, " ")

		imgPath := fmt.Sprintf("/static/images/fighters/%s", filename)

		rowsAffected, err := repo.Fighters.UpdateFighterImage(ctx, fighterName, imgPath)
		if err != nil {
			log.Printf("Failed to update fighter %s: %v", fighterName, err)
			continue
		}

		if rowsAffected > 0 {
			fmt.Printf("✅ Updated image for: %s\n", fighterName)
			updatedCount++
		}
	}

	fmt.Printf("\n--- Sync Complete ---\nSuccessfully updated %d fighters.\n", updatedCount)
}
