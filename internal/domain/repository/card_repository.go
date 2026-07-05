package repository

import (
	"context"
	"fmt"
	"time"

	"github.com/dancouver1/ufc-card-creator/internal/models"
	"github.com/jackc/pgx/v5/pgxpool"
)

type CardRepository struct {
	pool    *pgxpool.Pool
	matches *MatchRepository
}

// CreateCard inserts a new card into the database
func (r *CardRepository) CreateCard(ctx context.Context, card *models.Card) error {
	query := `
        INSERT INTO cards (card_name, card_number, location, event_date)
        VALUES ($1, $2, $3, $4)
        RETURNING id, created_at, updated_at
    `

	err := r.pool.QueryRow(
		ctx, query,
		card.CardName, card.CardNumber, card.Location, card.EventDate,
	).Scan(&card.ID, &card.CreatedAt, &card.UpdatedAt)

	if err != nil {
		return fmt.Errorf("insert failed: %w", err)
	}

	return nil
}

// GetOrCreateActiveCard retrieves the most recent card or creates a new one if none exists
func (r *CardRepository) GetOrCreateActiveCard(ctx context.Context) (int, error) {
	// Try to get the most recent card
	var cardID int
	query := `
        SELECT id FROM cards
        ORDER BY created_at DESC
        LIMIT 1
    `

	err := r.pool.QueryRow(ctx, query).Scan(&cardID)
	if err == nil {
		// Card found, return its ID
		return cardID, nil
	}

	// No card exists, create a new one
	now := time.Now()
	cardName := fmt.Sprintf("My Card - %s", now.Format("2006-01-02"))

	card := models.Card{
		CardName: cardName,
	}

	if err := r.CreateCard(ctx, &card); err != nil {
		return 0, fmt.Errorf("failed to create card: %w", err)
	}

	return card.ID, nil
}

// GetCardByID retrieves a card with all its matches
func (r *CardRepository) GetCardByID(ctx context.Context, id int) (*models.Card, error) {
	cardQuery := `
        SELECT id, card_name, card_number, location, event_date, created_at, updated_at
        FROM cards
        WHERE id = $1
    `

	var card models.Card
	err := r.pool.QueryRow(ctx, cardQuery, id).Scan(
		&card.ID, &card.CardName, &card.CardNumber, &card.Location,
		&card.EventDate, &card.CreatedAt, &card.UpdatedAt,
	)
	if err != nil {
		return nil, fmt.Errorf("card query failed: %w", err)
	}

	// Get all matches for this card
	matches, err := r.matches.GetMatchesByCardID(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("matches query failed: %w", err)
	}

	card.Matches = matches

	return &card, nil
}

// GetAllCards retrieves all cards
func (r *CardRepository) GetAllCards(ctx context.Context) ([]models.Card, error) {
	query := `
        SELECT id, card_name, card_number, location, event_date, created_at, updated_at
        FROM cards
        ORDER BY created_at DESC
    `

	rows, err := r.pool.Query(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("query failed: %w", err)
	}
	defer rows.Close()

	var cards []models.Card
	for rows.Next() {
		var c models.Card
		err := rows.Scan(
			&c.ID, &c.CardName, &c.CardNumber, &c.Location,
			&c.EventDate, &c.CreatedAt, &c.UpdatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("scan failed: %w", err)
		}
		cards = append(cards, c)
	}

	return cards, nil
}

// DeleteCard deletes a card and all its matches
func (r *CardRepository) DeleteCard(ctx context.Context, cardID int) error {
	// Start a transaction to ensure both card and matches are deleted
	tx, err := r.pool.Begin(ctx)
	if err != nil {
		return fmt.Errorf("failed to start transaction: %w", err)
	}
	defer tx.Rollback(ctx)

	// Delete matches first
	_, err = tx.Exec(ctx, "DELETE FROM matches WHERE card_id = $1", cardID)
	if err != nil {
		return fmt.Errorf("failed to delete matches: %w", err)
	}

	// Delete card
	_, err = tx.Exec(ctx, "DELETE FROM cards WHERE id = $1", cardID)
	if err != nil {
		return fmt.Errorf("failed to delete card: %w", err)
	}

	return tx.Commit(ctx)
}
