package models

import (
	"time"
)

type Card struct {
	ID         int        `json:"id"`
	CardName   string     `json:"card_name"`
	CardNumber *string    `json:"card_number,omitempty"`
	Location   *string    `json:"location,omitempty"`
	EventDate  *time.Time `json:"event_date,omitempty"`
	CreatedAt  time.Time  `json:"created_at"`
	UpdatedAt  time.Time  `json:"updated_at"`

	// For when we fetch card with matches
	Matches []Match `json:"matches,omitempty"`
}
