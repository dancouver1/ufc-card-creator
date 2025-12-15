package models

import (
	"time"
)

type Match struct {
	ID           int     `json:"id"`
	CardID       *int    `json:"card_id,omitempty"`
	Fighter1ID   int     `json:"fighter1_id"`
	Fighter2ID   int     `json:"fighter2_id"`
	FightOrder   *int    `json:"fight_order,omitempty"`
	WeightClass  *string `json:"weight_class,omitempty"`
	IsTitleFight bool    `json:"is_title_fight"`
	Rounds       int     `json:"rounds"`

	// NEW: Prediction field
	Prediction *int `json:"prediction,omitempty"` // Fighter ID of predicted winner (NULL if no prediction)

	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`

	// Populated when fetching match details
	Fighter1 *Fighter `json:"fighter1,omitempty"`
	Fighter2 *Fighter `json:"fighter2,omitempty"`
}


