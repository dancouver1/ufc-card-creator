package models

import (
	"fmt"
	"time"
)

type Fighter struct {
	ID              int        `json:"id"`
	Name            string     `json:"name"`
	Nickname        *string    `json:"nickname,omitempty"`
	HeightFeet      *int       `json:"height_feet,omitempty"`
	HeightInches    *int       `json:"height_inches,omitempty"`
	WeightLbs       *int       `json:"weight_lbs,omitempty"`
	ReachCm         *int       `json:"reach_cm,omitempty"`
	LegReachCm      *int       `json:"leg_reach_cm,omitempty"`
	WeightClass     string     `json:"weight_class"`
	Stance          *string    `json:"stance,omitempty"`
	Wins            int        `json:"wins"`
	Losses          int        `json:"losses"`
	Draws           int        `json:"draws"`
	DateOfBirth     *time.Time `json:"date_of_birth,omitempty"`
	Nationality     *string    `json:"nationality,omitempty"`
	FighterImageURL *string    `json:"fighter_image_url,omitempty"`
	IsActive        bool       `json:"is_active"`
	LastFightDate   *time.Time `json:"last_fight_date,omitempty"`
	CreatedAt       time.Time  `json:"created_at"`
	UpdatedAt       time.Time  `json:"updated_at"`

	// NEW: Last 5 fights (populated when needed)
	Last5Fights []FightHistory `json:"last_5_fights,omitempty"`
}

// GetRecord returns the fighter's record as a string (e.g., "20-5-0")
func (f *Fighter) GetRecord() string {
	return fmt.Sprintf("%d-%d-%d", f.Wins, f.Losses, f.Draws)
}

// GetHeight returns the fighter's height as a string (e.g., "5'11\"")
func (f *Fighter) GetHeight() string {
	if f.HeightFeet == nil {
		return "N/A"
	}
	inches := 0
	if f.HeightInches != nil {
		inches = *f.HeightInches
	}
	return fmt.Sprintf("%d'%d\"", *f.HeightFeet, inches)
}

// GetReach returns reach in inches (converted from cm)
func (f *Fighter) GetReachInches() string {
	if f.ReachCm == nil {
		return "N/A"
	}
	inches := float64(*f.ReachCm) / 2.54
	return fmt.Sprintf("%.1f\"", inches)
}
