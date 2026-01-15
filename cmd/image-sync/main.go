package main

import (
	"context"
	"fmt"
	"io/ioutil"
	"log"
	"path/filepath"
	"strings"

	"github.com/dancouver1/ufc-card-creator/internal/database"
	"github.com/joho/godotenv"
)

func main() {
	// Load .env
	_ = godotenv.Load()

	// Connect to DB
	db, err := database.NewDB()
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}
	defer db.Close()

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

		// Update DB
		query := "UPDATE fighters SET fighter_image_url = $1 WHERE name = $2"
		res, err := db.Pool.Exec(ctx, query, imgPath, fighterName)
		if err != nil {
			log.Printf("Failed to update fighter %s: %v", fighterName, err)
			continue
		}

		rowsAffected := res.RowsAffected()
		if rowsAffected > 0 {
			fmt.Printf("✅ Updated image for: %s\n", fighterName)
			updatedCount++
		}
	}

	fmt.Printf("\n--- Sync Complete ---\nSuccessfully updated %d fighters.\n", updatedCount)
}
