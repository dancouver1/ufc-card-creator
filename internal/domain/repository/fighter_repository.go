package repository

import (
	"context"
	"fmt"

	"github.com/dancouver1/ufc-card-creator/internal/models"
	"github.com/jackc/pgx/v5/pgxpool"
)

type FighterRepository struct {
	pool *pgxpool.Pool
}

// GetAllFighters retrieves all fighters from the database
func (r *FighterRepository) GetAllFighters(ctx context.Context) ([]models.Fighter, error) {
	query := `
        SELECT id, name, nickname, height_feet, height_inches, weight_lbs,
               reach_cm, leg_reach_cm, weight_class, stance, wins, losses, draws,
               date_of_birth, nationality, fighter_image_url, is_active,
               last_fight_date, created_at, updated_at
        FROM fighters
        WHERE is_active = true
        ORDER BY name ASC
    `

	rows, err := r.pool.Query(ctx, query)
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
func (r *FighterRepository) GetFighterByID(ctx context.Context, id int) (*models.Fighter, error) {
	query := `
        SELECT id, name, nickname, height_feet, height_inches, weight_lbs,
               reach_cm, leg_reach_cm, weight_class, stance, wins, losses, draws,
               date_of_birth, nationality, fighter_image_url, is_active,
               last_fight_date, created_at, updated_at
        FROM fighters
        WHERE id = $1
    `

	var f models.Fighter
	err := r.pool.QueryRow(ctx, query, id).Scan(
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
func (r *FighterRepository) GetFightersByWeightClass(ctx context.Context, weightClass string) ([]models.Fighter, error) {
	query := `
        SELECT id, name, nickname, height_feet, height_inches, weight_lbs,
               reach_cm, leg_reach_cm, weight_class, stance, wins, losses, draws,
               date_of_birth, nationality, fighter_image_url, is_active,
               last_fight_date, created_at, updated_at
        FROM fighters
        WHERE weight_class = $1 AND is_active = true
        ORDER BY name ASC
    `

	rows, err := r.pool.Query(ctx, query, weightClass)
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
func (r *FighterRepository) SearchFighters(ctx context.Context, searchTerm string) ([]models.Fighter, error) {
	return r.SearchFightersWithFilters(ctx, searchTerm, "")
}

// SearchFightersWithFilters searches for fighters by name and/or weight class
func (r *FighterRepository) SearchFightersWithFilters(ctx context.Context, searchTerm string, weightClass string) ([]models.Fighter, error) {
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

	rows, err := r.pool.Query(ctx, query, params...)
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
func (r *FighterRepository) CreateFighter(ctx context.Context, fighter *models.Fighter) error {
	query := `
        INSERT INTO fighters (
            name, nickname, height_feet, height_inches, weight_lbs,
            reach_cm, leg_reach_cm, weight_class, stance, wins, losses, draws,
            date_of_birth, nationality, fighter_image_url, is_active, last_fight_date
        ) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15, $16, $17)
        RETURNING id, created_at, updated_at
    `

	err := r.pool.QueryRow(
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
func (r *FighterRepository) UpsertFighter(ctx context.Context, fighter *models.Fighter) error {
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
            leg_reach_cm = COALESCE(EXCLUDED.leg_reach_cm, fighters.leg_reach_cm),
            weight_class = EXCLUDED.weight_class,
            stance = COALESCE(EXCLUDED.stance, fighters.stance),
            wins = EXCLUDED.wins,
            losses = EXCLUDED.losses,
            draws = EXCLUDED.draws,
            date_of_birth = COALESCE(EXCLUDED.date_of_birth, fighters.date_of_birth),
            nationality = COALESCE(EXCLUDED.nationality, fighters.nationality),
            fighter_image_url = COALESCE(EXCLUDED.fighter_image_url, fighters.fighter_image_url),
            is_active = EXCLUDED.is_active,
            last_fight_date = COALESCE(EXCLUDED.last_fight_date, fighters.last_fight_date),
            updated_at = CURRENT_TIMESTAMP
        RETURNING id, created_at, updated_at
    `

	err := r.pool.QueryRow(
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

// UpdateFighterImage sets a fighter's image URL by name. Returns the number of rows affected.
func (r *FighterRepository) UpdateFighterImage(ctx context.Context, name string, imagePath string) (int64, error) {
	query := "UPDATE fighters SET fighter_image_url = $1 WHERE name = $2"
	res, err := r.pool.Exec(ctx, query, imagePath, name)
	if err != nil {
		return 0, fmt.Errorf("update failed: %w", err)
	}
	return res.RowsAffected(), nil
}
