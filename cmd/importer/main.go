package main

import (
	"context"
	"encoding/csv"
	"flag"
	"fmt"
	"io"
	"log"
	"math"
	"os"
	"strconv"

	"github.com/dancouver1/ufc-card-creator/internal/config"
	"github.com/dancouver1/ufc-card-creator/internal/db"
	"github.com/dancouver1/ufc-card-creator/internal/domain/repository"
	"github.com/dancouver1/ufc-card-creator/internal/models"
)

func main() {
	filePath := flag.String("file", "", "Path to the CSV file")
	flag.Parse()

	if *filePath == "" {
		log.Fatal("Please provide a file path using -file")
	}

	cfg := config.Load()

	// Connect to DB
	database, err := db.NewDB(cfg.DatabaseURL)
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}
	defer database.Close()

	repo := repository.New(database.Pool)

	file, err := os.Open(*filePath)
	if err != nil {
		log.Fatalf("Failed to open file: %v", err)
	}
	defer file.Close()

	reader := csv.NewReader(file)
	header, err := reader.Read()
	if err != nil {
		log.Fatalf("Failed to read header: %v", err)
	}

	colMap := make(map[string]int)
	for i, name := range header {
		colMap[name] = i
	}

	ctx := context.Background()
	count := 0

	for {
		record, err := reader.Read()
		if err == io.EOF {
			break
		}
		if err != nil {
			log.Printf("Error reading record: %v", err)
			continue
		}

		fighter := parseFighter(record, colMap)
		if fighter.Name == "" {
			continue
		}

		err = repo.Fighters.UpsertFighter(ctx, fighter)
		if err != nil {
			log.Printf("Failed to upsert fighter %s: %v", fighter.Name, err)
			continue
		}
		count++
		if count%100 == 0 {
			fmt.Printf("Processed %d fighters...\n", count)
		}
	}

	fmt.Printf("Successfully imported/updated %d fighters!\n", count)
}

func parseFighter(record []string, colMap map[string]int) *models.Fighter {
	name := record[colMap["name"]]
	nickname := record[colMap["nick_name"]]

	wins, _ := strconv.Atoi(record[colMap["wins"]])
	losses, _ := strconv.Atoi(record[colMap["losses"]])
	draws, _ := strconv.Atoi(record[colMap["draws"]])

	// Height cm -> feet/inches
	heightCm, _ := strconv.ParseFloat(record[colMap["height"]], 64)
	var ft, in int
	if heightCm > 0 {
		totalInches := heightCm / 2.54
		ft = int(totalInches / 12)
		in = int(math.Round(totalInches - float64(ft*12)))
		if in == 12 {
			ft++
			in = 0
		}
	}

	// Weight kg -> lbs
	weightKg, _ := strconv.ParseFloat(record[colMap["weight"]], 64)
	weightLbs := int(math.Round(weightKg * 2.20462))

	// Reach cm
	reachCmF, _ := strconv.ParseFloat(record[colMap["reach"]], 64)
	reachCm := int(math.Round(reachCmF))

	stance := record[colMap["stance"]]

	f := &models.Fighter{
		Name:     name,
		Wins:     wins,
		Losses:   losses,
		Draws:    draws,
		IsActive: true,
	}

	if nickname != "" {
		f.Nickname = &nickname
	}
	if ft > 0 {
		f.HeightFeet = &ft
		f.HeightInches = &in
	}
	if weightLbs > 0 {
		f.WeightLbs = &weightLbs
		f.WeightClass = inferWeightClass(weightLbs)
	} else {
		f.WeightClass = "Unknown"
	}
	if reachCm > 0 {
		f.ReachCm = &reachCm
	}
	if stance != "" && stance != "N/A" {
		f.Stance = &stance
	}

	return f
}

func inferWeightClass(lbs int) string {
	switch {
	case lbs <= 115:
		return "Strawweight"
	case lbs <= 125:
		return "Flyweight"
	case lbs <= 135:
		return "Bantamweight"
	case lbs <= 145:
		return "Featherweight"
	case lbs <= 155:
		return "Lightweight"
	case lbs <= 170:
		return "Welterweight"
	case lbs <= 185:
		return "Middleweight"
	case lbs <= 205:
		return "Light Heavyweight"
	default:
		return "Heavyweight"
	}
}
