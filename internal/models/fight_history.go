package models

import (
	"time"
)

type FightHistory struct {
	ID           int        `json:"id"`
	FighterID    int        `json:"fighter_id"`
	OpponentName string     `json:"opponent_name"`
	Result       string     `json:"result"` // 'win', 'loss', 'draw', 'nc'
	Method       *string    `json:"method,omitempty"`
	Round        *int       `json:"round,omitempty"`
	FightDate    *time.Time `json:"fight_date,omitempty"`
	EventName    *string    `json:"event_name,omitempty"`
	FightOrder   int        `json:"fight_order"`
	CreatedAt    time.Time  `json:"created_at"`
}

// GetResultColor returns CSS color class based on result
func (fh *FightHistory) GetResultColor() string {
	switch fh.Result {
	case "win":
		return "green"
	case "loss":
		return "red"
	case "draw":
		return "yellow"
	default:
		return "gray"
	}
}
