package repository

import "github.com/jackc/pgx/v5/pgxpool"

// Repository aggregates all per-entity repositories over a shared connection pool.
type Repository struct {
	Fighters     *FighterRepository
	Cards        *CardRepository
	Matches      *MatchRepository
	FightHistory *FightHistoryRepository
	Rankings     *RankingRepository
}

// New builds a Repository backed by the given pool.
func New(pool *pgxpool.Pool) *Repository {
	matches := &MatchRepository{pool: pool}

	return &Repository{
		Fighters:     &FighterRepository{pool: pool},
		Cards:        &CardRepository{pool: pool, matches: matches},
		Matches:      matches,
		FightHistory: &FightHistoryRepository{pool: pool},
		Rankings:     &RankingRepository{pool: pool},
	}
}
