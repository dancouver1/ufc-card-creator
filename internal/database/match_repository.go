package database
// CreateMatch inserts a new match into the database
func (db *DB) CreateMatch(ctx context.Context, match *models.Match) error {
	query := `
        INSERT INTO matches (
            card_id, fighter1_id, fighter2_id, fight_order,
            weight_class, is_title_fight, rounds, prediction
        ) VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
        RETURNING id, created_at, updated_at
    `

	err := db.Pool.QueryRow(
		ctx, query,
		match.CardID, match.Fighter1ID, match.Fighter2ID, match.FightOrder,
		match.WeightClass, match.IsTitleFight, match.Rounds, match.Prediction,
	).Scan(&match.ID, &match.CreatedAt, &match.UpdatedAt)

	if err != nil {
		return fmt.Errorf("insert failed: %w", err)
	}

	return nil
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