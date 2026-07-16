package repository

import (
	"context"
	"fmt"

	"github.com/dancouver1/ufc-card-creator/internal/models"
	"github.com/jackc/pgx/v5/pgxpool"
)

type RankingRepository struct {
	pool *pgxpool.Pool
}

// GetByType returns all rankings for a rank type ("media" or "meta"),
// ordered by division then position, with the champion (position 0) first
// in each division.
func (r *RankingRepository) GetByType(ctx context.Context, rankType models.RankType) ([]models.Ranking, error) {
	query := `
        SELECT id, rank_type, division, position, fighter_name, fighter_slug,
               is_champion, movement, champion_image_url, updated_at
        FROM rankings
        WHERE rank_type = $1
        ORDER BY division ASC, position ASC
    `

	rows, err := r.pool.Query(ctx, query, rankType)
	if err != nil {
		return nil, fmt.Errorf("query failed: %w", err)
	}
	defer rows.Close()

	var rankings []models.Ranking
	for rows.Next() {
		var rk models.Ranking
		var fighterSlug, championImageURL *string
		if err := rows.Scan(
			&rk.ID, &rk.RankType, &rk.Division, &rk.Position, &rk.FighterName,
			&fighterSlug, &rk.IsChampion, &rk.Movement, &championImageURL, &rk.UpdatedAt,
		); err != nil {
			return nil, fmt.Errorf("scan failed: %w", err)
		}
		if fighterSlug != nil {
			rk.FighterSlug = *fighterSlug
		}
		if championImageURL != nil {
			rk.ChampionImageURL = *championImageURL
		}
		rankings = append(rankings, rk)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("rows iteration failed: %w", err)
	}

	return rankings, nil
}

// LastUpdated returns the most recent updated_at across all rankings, or the
// zero value if the table is empty.
func (r *RankingRepository) LastUpdated(ctx context.Context) (*models.Ranking, error) {
	var rk models.Ranking
	err := r.pool.QueryRow(ctx, `SELECT updated_at FROM rankings ORDER BY updated_at DESC LIMIT 1`).Scan(&rk.UpdatedAt)
	if err != nil {
		return nil, err
	}
	return &rk, nil
}

// ReplaceAll atomically replaces every ranking row for the given rank type.
// Used by the scraper: it deletes the old snapshot and inserts the freshly
// scraped one in a single transaction so readers never see a partial table.
func (r *RankingRepository) ReplaceAll(ctx context.Context, rankType models.RankType, rankings []models.Ranking) error {
	tx, err := r.pool.Begin(ctx)
	if err != nil {
		return fmt.Errorf("begin tx failed: %w", err)
	}
	defer tx.Rollback(ctx)

	if _, err := tx.Exec(ctx, `DELETE FROM rankings WHERE rank_type = $1`, rankType); err != nil {
		return fmt.Errorf("delete failed: %w", err)
	}

	for _, rk := range rankings {
		_, err := tx.Exec(ctx, `
            INSERT INTO rankings (
                rank_type, division, position, fighter_name, fighter_slug,
                is_champion, movement, champion_image_url
            ) VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
        `, rankType, rk.Division, rk.Position, rk.FighterName, rk.FighterSlug,
			rk.IsChampion, rk.Movement, rk.ChampionImageURL)
		if err != nil {
			return fmt.Errorf("insert failed for %s #%d %s: %w", rk.Division, rk.Position, rk.FighterName, err)
		}
	}

	if err := tx.Commit(ctx); err != nil {
		return fmt.Errorf("commit failed: %w", err)
	}

	return nil
}
