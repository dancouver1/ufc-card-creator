package repository

import (
	"context"
	"fmt"

	"github.com/dancouver1/ufc-card-creator/internal/models"
	"github.com/jackc/pgx/v5/pgxpool"
)

type MatchRepository struct {
	pool *pgxpool.Pool
}

// CreateMatch inserts a new match into the database
func (r *MatchRepository) CreateMatch(ctx context.Context, match *models.Match) error {
	// Start a transaction to ensure atomicity
	tx, err := r.pool.Begin(ctx)
	if err != nil {
		return fmt.Errorf("failed to start transaction: %w", err)
	}
	defer tx.Rollback(ctx)

	// If this is a main event, unset any existing main event on this card
	if match.IsMainEvent {
		_, err = tx.Exec(ctx, "UPDATE matches SET is_main_event = false WHERE card_id = $1", match.CardID)
		if err != nil {
			return fmt.Errorf("failed to unset existing main event: %w", err)
		}
	}

	// If this is a co-main event, unset any existing co-main event on this card
	if match.IsCoMainEvent {
		_, err = tx.Exec(ctx, "UPDATE matches SET is_co_main_event = false WHERE card_id = $1", match.CardID)
		if err != nil {
			return fmt.Errorf("failed to unset existing co-main event: %w", err)
		}
	}

	query := `
        INSERT INTO matches (
            card_id, fighter1_id, fighter2_id, fight_order,
            weight_class, is_title_fight, rounds, prediction,
            is_main_event, is_co_main_event, title_type, card_part
        ) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12)
        RETURNING id, created_at, updated_at
    `

	err = tx.QueryRow(
		ctx, query,
		match.CardID, match.Fighter1ID, match.Fighter2ID, match.FightOrder,
		match.WeightClass, match.IsTitleFight, match.Rounds, match.Prediction,
		match.IsMainEvent, match.IsCoMainEvent, match.TitleType, match.CardPart,
	).Scan(&match.ID, &match.CreatedAt, &match.UpdatedAt)

	if err != nil {
		return fmt.Errorf("insert failed: %w", err)
	}

	return tx.Commit(ctx)
}

// GetMatchesByCardID retrieves all matches for a specific card with fighter details
func (r *MatchRepository) GetMatchesByCardID(ctx context.Context, cardID int) ([]models.Match, error) {
	query := `
        SELECT
            m.id, m.card_id, m.fighter1_id, m.fighter2_id, m.fight_order,
            m.weight_class, m.is_title_fight, m.rounds, m.prediction,
            m.is_main_event, m.is_co_main_event, m.title_type, m.card_part,
            m.created_at, m.updated_at,
            f1.id, f1.name, f1.nickname, f1.height_feet, f1.height_inches,
            f1.weight_lbs, f1.reach_cm, f1.weight_class, f1.stance,
            f1.wins, f1.losses, f1.draws, f1.fighter_image_url,
            f2.id, f2.name, f2.nickname, f2.height_feet, f2.height_inches,
            f2.weight_lbs, f2.reach_cm, f2.weight_class, f2.stance,
            f2.wins, f2.losses, f2.draws, f2.fighter_image_url
        FROM matches m
        JOIN fighters f1 ON m.fighter1_id = f1.id
        JOIN fighters f2 ON m.fighter2_id = f2.id
        WHERE m.card_id = $1
        ORDER BY
            m.is_main_event DESC,
            m.is_co_main_event DESC,
            m.id DESC
    `

	rows, err := r.pool.Query(ctx, query, cardID)
	if err != nil {
		return nil, fmt.Errorf("query failed: %w", err)
	}
	defer rows.Close()

	var matches []models.Match
	for rows.Next() {
		var m models.Match
		var f1, f2 models.Fighter

		err := rows.Scan(
			&m.ID, &m.CardID, &m.Fighter1ID, &m.Fighter2ID, &m.FightOrder,
			&m.WeightClass, &m.IsTitleFight, &m.Rounds, &m.Prediction,
			&m.IsMainEvent, &m.IsCoMainEvent, &m.TitleType, &m.CardPart,
			&m.CreatedAt, &m.UpdatedAt,
			&f1.ID, &f1.Name, &f1.Nickname, &f1.HeightFeet, &f1.HeightInches,
			&f1.WeightLbs, &f1.ReachCm, &f1.WeightClass, &f1.Stance,
			&f1.Wins, &f1.Losses, &f1.Draws, &f1.FighterImageURL,
			&f2.ID, &f2.Name, &f2.Nickname, &f2.HeightFeet, &f2.HeightInches,
			&f2.WeightLbs, &f2.ReachCm, &f2.WeightClass, &f2.Stance,
			&f2.Wins, &f2.Losses, &f2.Draws, &f2.FighterImageURL,
		)
		if err != nil {
			return nil, fmt.Errorf("scan failed: %w", err)
		}

		f1.SetDefaultImageURL()
		f2.SetDefaultImageURL()

		m.Fighter1 = &f1
		m.Fighter2 = &f2
		matches = append(matches, m)
	}

	return matches, nil
}

// UpdateMatchPrediction updates the prediction for a match
func (r *MatchRepository) UpdateMatchPrediction(ctx context.Context, matchID int, predictionFighterID *int) error {
	query := `
        UPDATE matches
        SET prediction = $1, updated_at = CURRENT_TIMESTAMP
        WHERE id = $2
    `

	_, err := r.pool.Exec(ctx, query, predictionFighterID, matchID)
	if err != nil {
		return fmt.Errorf("update failed: %w", err)
	}

	return nil
}

// DeleteMatch deletes a specific match
func (r *MatchRepository) DeleteMatch(ctx context.Context, matchID int) error {
	query := `DELETE FROM matches WHERE id = $1`
	_, err := r.pool.Exec(ctx, query, matchID)
	if err != nil {
		return fmt.Errorf("delete failed: %w", err)
	}
	return nil
}
