package database

import (
	"context"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/dancouver1/ufc-card-creator/internal/models"
	"github.com/jackc/pgx/v5/pgxpool"
)

type DB struct {
	Pool *pgxpool.Pool
}

// NewDB creates a new database connection pool
func NewDB() (*DB, error) {
	databaseURL := os.Getenv("DATABASE_URL")
	if databaseURL == "" {
		return nil, fmt.Errorf("DATABASE_URL environment variable is not set")
	}

	config, err := pgxpool.ParseConfig(databaseURL)
	if err != nil {
		return nil, fmt.Errorf("unable to parse DATABASE_URL: %w", err)
	}

	// Configure pool settings
	config.MaxConns = 25
	config.MinConns = 5

	pool, err := pgxpool.NewWithConfig(context.Background(), config)
	if err != nil {
		return nil, fmt.Errorf("unable to create connection pool: %w", err)
	}

	// Test the connection
	if err := pool.Ping(context.Background()); err != nil {
		return nil, fmt.Errorf("unable to ping database: %w", err)
	}

	log.Println("Successfully connected to database")

	return &DB{Pool: pool}, nil
}

// Close closes the database connection pool
func (db *DB) Close() {
	db.Pool.Close()
}

// Health checks if the database connection is alive
func (db *DB) Health(ctx context.Context) error {
	return db.Pool.Ping(ctx)
}

// ============== FIGHTER METHODS ==============

// GetAllFighters retrieves all fighters from the database
func (db *DB) GetAllFighters(ctx context.Context) ([]models.Fighter, error) {
	query := `
        SELECT id, name, nickname, height_feet, height_inches, weight_lbs,
               reach_cm, leg_reach_cm, weight_class, stance, wins, losses, draws,
               date_of_birth, nationality, fighter_image_url, is_active,
               last_fight_date, created_at, updated_at
        FROM fighters
        WHERE is_active = true
        ORDER BY name ASC
    `

	rows, err := db.Pool.Query(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("query failed: %w", err)
	}
	defer rows.Close()

	var fighters []models.Fighter
	for rows.Next() {
		var f models.Fighter
		err := rows.Scan(
			&f.ID, &f.Name, &f.Nickname, &f.HeightFeet, &f.HeightInches,
			&f.WeightLbs, &f.ReachCm, &f.LegReachCm, &f.WeightClass, &f.Stance,
			&f.Wins, &f.Losses, &f.Draws, &f.DateOfBirth, &f.Nationality,
			&f.FighterImageURL, &f.IsActive, &f.LastFightDate,
			&f.CreatedAt, &f.UpdatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("scan failed: %w", err)
		}
		f.SetDefaultImageURL()
		fighters = append(fighters, f)
	}

	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("rows iteration failed: %w", err)
	}

	return fighters, nil
}

// GetFighterByID retrieves a single fighter by ID
func (db *DB) GetFighterByID(ctx context.Context, id int) (*models.Fighter, error) {
	query := `
        SELECT id, name, nickname, height_feet, height_inches, weight_lbs,
               reach_cm, leg_reach_cm, weight_class, stance, wins, losses, draws,
               date_of_birth, nationality, fighter_image_url, is_active,
               last_fight_date, created_at, updated_at
        FROM fighters
        WHERE id = $1
    `

	var f models.Fighter
	err := db.Pool.QueryRow(ctx, query, id).Scan(
		&f.ID, &f.Name, &f.Nickname, &f.HeightFeet, &f.HeightInches,
		&f.WeightLbs, &f.ReachCm, &f.LegReachCm, &f.WeightClass, &f.Stance,
		&f.Wins, &f.Losses, &f.Draws, &f.DateOfBirth, &f.Nationality,
		&f.FighterImageURL, &f.IsActive, &f.LastFightDate,
		&f.CreatedAt, &f.UpdatedAt,
	)

	if err != nil {
		return nil, fmt.Errorf("query failed: %w", err)
	}

	f.SetDefaultImageURL()
	return &f, nil
}

// GetFightersByWeightClass retrieves fighters filtered by weight class
func (db *DB) GetFightersByWeightClass(ctx context.Context, weightClass string) ([]models.Fighter, error) {
	query := `
        SELECT id, name, nickname, height_feet, height_inches, weight_lbs,
               reach_cm, leg_reach_cm, weight_class, stance, wins, losses, draws,
               date_of_birth, nationality, fighter_image_url, is_active,
               last_fight_date, created_at, updated_at
        FROM fighters
        WHERE weight_class = $1 AND is_active = true
        ORDER BY name ASC
    `

	rows, err := db.Pool.Query(ctx, query, weightClass)
	if err != nil {
		return nil, fmt.Errorf("query failed: %w", err)
	}
	defer rows.Close()

	var fighters []models.Fighter
	for rows.Next() {
		var f models.Fighter
		err := rows.Scan(
			&f.ID, &f.Name, &f.Nickname, &f.HeightFeet, &f.HeightInches,
			&f.WeightLbs, &f.ReachCm, &f.LegReachCm, &f.WeightClass, &f.Stance,
			&f.Wins, &f.Losses, &f.Draws, &f.DateOfBirth, &f.Nationality,
			&f.FighterImageURL, &f.IsActive, &f.LastFightDate,
			&f.CreatedAt, &f.UpdatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("scan failed: %w", err)
		}
		f.SetDefaultImageURL()
		fighters = append(fighters, f)
	}

	return fighters, nil
}

// SearchFighters searches for fighters by name (fuzzy search)
func (db *DB) SearchFighters(ctx context.Context, searchTerm string) ([]models.Fighter, error) {
	return db.SearchFightersWithFilters(ctx, searchTerm, "")
}

// SearchFightersWithFilters searches for fighters by name and/or weight class
func (db *DB) SearchFightersWithFilters(ctx context.Context, searchTerm string, weightClass string) ([]models.Fighter, error) {
	query := `
        SELECT id, name, nickname, height_feet, height_inches, weight_lbs,
               reach_cm, leg_reach_cm, weight_class, stance, wins, losses, draws,
               date_of_birth, nationality, fighter_image_url, is_active,
               last_fight_date, created_at, updated_at
        FROM fighters
        WHERE is_active = true
    `
	params := []interface{}{}
	paramCount := 1

	if searchTerm != "" {
		query += fmt.Sprintf(" AND (name ILIKE $%d OR nickname ILIKE $%d)", paramCount, paramCount)
		params = append(params, "%"+searchTerm+"%")
		paramCount++
	}

	if weightClass != "" && weightClass != "all" {
		query += fmt.Sprintf(" AND weight_class = $%d", paramCount)
		params = append(params, weightClass)
		paramCount++
	}

	query += " ORDER BY name ASC LIMIT 50"

	rows, err := db.Pool.Query(ctx, query, params...)
	if err != nil {
		return nil, fmt.Errorf("query failed: %w", err)
	}
	defer rows.Close()

	var fighters []models.Fighter
	for rows.Next() {
		var f models.Fighter
		err := rows.Scan(
			&f.ID, &f.Name, &f.Nickname, &f.HeightFeet, &f.HeightInches,
			&f.WeightLbs, &f.ReachCm, &f.LegReachCm, &f.WeightClass, &f.Stance,
			&f.Wins, &f.Losses, &f.Draws, &f.DateOfBirth, &f.Nationality,
			&f.FighterImageURL, &f.IsActive, &f.LastFightDate,
			&f.CreatedAt, &f.UpdatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("scan failed: %w", err)
		}
		f.SetDefaultImageURL()
		fighters = append(fighters, f)
	}

	return fighters, nil
}

// CreateFighter inserts a new fighter into the database
func (db *DB) CreateFighter(ctx context.Context, fighter *models.Fighter) error {
	query := `
        INSERT INTO fighters (
            name, nickname, height_feet, height_inches, weight_lbs,
            reach_cm, leg_reach_cm, weight_class, stance, wins, losses, draws,
            date_of_birth, nationality, fighter_image_url, is_active, last_fight_date
        ) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15, $16, $17)
        RETURNING id, created_at, updated_at
    `

	err := db.Pool.QueryRow(
		ctx, query,
		fighter.Name, fighter.Nickname, fighter.HeightFeet, fighter.HeightInches,
		fighter.WeightLbs, fighter.ReachCm, fighter.LegReachCm, fighter.WeightClass,
		fighter.Stance, fighter.Wins, fighter.Losses, fighter.Draws,
		fighter.DateOfBirth, fighter.Nationality, fighter.FighterImageURL,
		fighter.IsActive, fighter.LastFightDate,
	).Scan(&fighter.ID, &fighter.CreatedAt, &fighter.UpdatedAt)

	if err != nil {
		return fmt.Errorf("insert failed: %w", err)
	}

	return nil
}

// UpsertFighter inserts or updates a fighter based on name
func (db *DB) UpsertFighter(ctx context.Context, fighter *models.Fighter) error {
	query := `
        INSERT INTO fighters (
            name, nickname, height_feet, height_inches, weight_lbs,
            reach_cm, leg_reach_cm, weight_class, stance, wins, losses, draws,
            date_of_birth, nationality, fighter_image_url, is_active, last_fight_date
        ) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15, $16, $17)
        ON CONFLICT (name) DO UPDATE SET
            nickname = EXCLUDED.nickname,
            height_feet = COALESCE(EXCLUDED.height_feet, fighters.height_feet),
            height_inches = COALESCE(EXCLUDED.height_inches, fighters.height_inches),
            weight_lbs = COALESCE(EXCLUDED.weight_lbs, fighters.weight_lbs),
            reach_cm = COALESCE(EXCLUDED.reach_cm, fighters.reach_cm),
            weight_class = EXCLUDED.weight_class,
            stance = EXCLUDED.stance,
            wins = EXCLUDED.wins,
            losses = EXCLUDED.losses,
            draws = EXCLUDED.draws,
            updated_at = CURRENT_TIMESTAMP
        RETURNING id, created_at, updated_at
    `

	err := db.Pool.QueryRow(
		ctx, query,
		fighter.Name, fighter.Nickname, fighter.HeightFeet, fighter.HeightInches,
		fighter.WeightLbs, fighter.ReachCm, fighter.LegReachCm, fighter.WeightClass,
		fighter.Stance, fighter.Wins, fighter.Losses, fighter.Draws,
		fighter.DateOfBirth, fighter.Nationality, fighter.FighterImageURL,
		fighter.IsActive, fighter.LastFightDate,
	).Scan(&fighter.ID, &fighter.CreatedAt, &fighter.UpdatedAt)

	if err != nil {
		return fmt.Errorf("upsert failed: %w", err)
	}

	return nil
}

// ============== CARD METHODS ==============

// CreateCard inserts a new card into the database
func (db *DB) CreateCard(ctx context.Context, card *models.Card) error {
	query := `
        INSERT INTO cards (card_name, card_number, location, event_date)
        VALUES ($1, $2, $3, $4)
        RETURNING id, created_at, updated_at
    `

	err := db.Pool.QueryRow(
		ctx, query,
		card.CardName, card.CardNumber, card.Location, card.EventDate,
	).Scan(&card.ID, &card.CreatedAt, &card.UpdatedAt)

	if err != nil {
		return fmt.Errorf("insert failed: %w", err)
	}

	return nil
}

// GetOrCreateActiveCard retrieves the most recent card or creates a new one if none exists
func (db *DB) GetOrCreateActiveCard(ctx context.Context) (int, error) {
	// Try to get the most recent card
	var cardID int
	query := `
        SELECT id FROM cards
        ORDER BY created_at DESC
        LIMIT 1
    `

	err := db.Pool.QueryRow(ctx, query).Scan(&cardID)
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

	if err := db.CreateCard(ctx, &card); err != nil {
		return 0, fmt.Errorf("failed to create card: %w", err)
	}

	return card.ID, nil
}

// GetCardByID retrieves a card with all its matches
func (db *DB) GetCardByID(ctx context.Context, id int) (*models.Card, error) {
	cardQuery := `
        SELECT id, card_name, card_number, location, event_date, created_at, updated_at
        FROM cards
        WHERE id = $1
    `

	var card models.Card
	err := db.Pool.QueryRow(ctx, cardQuery, id).Scan(
		&card.ID, &card.CardName, &card.CardNumber, &card.Location,
		&card.EventDate, &card.CreatedAt, &card.UpdatedAt,
	)
	if err != nil {
		return nil, fmt.Errorf("card query failed: %w", err)
	}

	// Get all matches for this card
	matches, err := db.GetMatchesByCardID(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("matches query failed: %w", err)
	}

	card.Matches = matches

	return &card, nil
}

// GetAllCards retrieves all cards
func (db *DB) GetAllCards(ctx context.Context) ([]models.Card, error) {
	query := `
        SELECT id, card_name, card_number, location, event_date, created_at, updated_at
        FROM cards
        ORDER BY created_at DESC
    `

	rows, err := db.Pool.Query(ctx, query)
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
func (db *DB) DeleteCard(ctx context.Context, cardID int) error {
	// Start a transaction to ensure both card and matches are deleted
	tx, err := db.Pool.Begin(ctx)
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

// ============== MATCH METHODS ==============

// CreateMatch inserts a new match into the database
func (db *DB) CreateMatch(ctx context.Context, match *models.Match) error {
	// Start a transaction to ensure atomicity
	tx, err := db.Pool.Begin(ctx)
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
func (db *DB) GetMatchesByCardID(ctx context.Context, cardID int) ([]models.Match, error) {
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

	rows, err := db.Pool.Query(ctx, query, cardID)
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
func (db *DB) UpdateMatchPrediction(ctx context.Context, matchID int, predictionFighterID *int) error {
	query := `
        UPDATE matches
        SET prediction = $1, updated_at = CURRENT_TIMESTAMP
        WHERE id = $2
    `

	_, err := db.Pool.Exec(ctx, query, predictionFighterID, matchID)
	if err != nil {
		return fmt.Errorf("update failed: %w", err)
	}

	return nil
}

// DeleteMatch deletes a specific match
func (db *DB) DeleteMatch(ctx context.Context, matchID int) error {
	query := `DELETE FROM matches WHERE id = $1`
	_, err := db.Pool.Exec(ctx, query, matchID)
	if err != nil {
		return fmt.Errorf("delete failed: %w", err)
	}
	return nil
}

// ============== FIGHT HISTORY METHODS ==============

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
