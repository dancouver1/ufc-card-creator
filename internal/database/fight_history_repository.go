package database

import (
	"context"
	"fmt"

	"github.com/dancouver1/ufc-matchmaker/internal/models"
)

// GetLast5Fights retrieves the last 5 fights for a fighter
func (db *DB) GetLast5Fights(ctx context.Context, fighterID int) ([]models.FightHistory, error) {
	query := `
        SELECT id, fighter_id, opponent_name, result, method, round,
               fight_date, event_name, fight_order, created_at
        FROM fight_history
        WHERE fighter_id = $1
        ORDER BY fight_order DESC, fight_date DESC NULLS LAST
        LIMIT 5
    `

	rows, err := db.Pool.Query(ctx, query, fighterID)
	if err != nil {
		return nil, fmt.Errorf("query failed: %w", err)
	}
	defer rows.Close()

	var fights []models.FightHistory
	for rows.Next() {
		var f models.FightHistory
		err := rows.Scan(
			&f.ID, &f.FighterID, &f.OpponentName, &f.Result, &f.Method,
			&f.Round, &f.FightDate, &f.EventName, &f.FightOrder, &f.CreatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("scan failed: %w", err)
		}
		fights = append(fights, f)
	}

	return fights, nil
}

// CreateFightHistory inserts a new fight record
func (db *DB) CreateFightHistory(ctx context.Context, fight *models.FightHistory) error {
	query := `
        INSERT INTO fight_history (
            fighter_id, opponent_name, result, method, round,
            fight_date, event_name, fight_order
        ) VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
        RETURNING id, created_at
    `

	err := db.Pool.QueryRow(
		ctx, query,
		fight.FighterID, fight.OpponentName, fight.Result, fight.Method,
		fight.Round, fight.FightDate, fight.EventName, fight.FightOrder,
	).Scan(&fight.ID, &fight.CreatedAt)

	if err != nil {
		return fmt.Errorf("insert failed: %w", err)
	}

	return nil
}
