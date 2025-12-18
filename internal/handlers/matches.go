package handlers

import (
	"context"
	"encoding/json"
	"net/http"

	"github.com/dancouver1/ufc-card-creator/internal/models"
)

// HandleCreateMatch creates a new match
func (h *Handler) HandleCreateMatch(w http.ResponseWriter, r *http.Request) {
	ctx := context.Background()

	var match models.Match
	if err := json.NewDecoder(r.Body).Decode(&match); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	// Validate that both fighters are selected
	if match.Fighter1ID == 0 || match.Fighter2ID == 0 {
		http.Error(w, "Both fighters must be selected", http.StatusBadRequest)
		return
	}

	// Set default values if not provided
	if match.Rounds == 0 {
		match.Rounds = 3
	}

	// Validate that a card is selected (sent from frontend)
	if match.CardID == nil || *match.CardID <= 0 {
		http.Error(w, "Please select a card to assign this match to", http.StatusBadRequest)
		return
	}

	// Create the match
	if err := h.DB.CreateMatch(ctx, &match); err != nil {
		http.Error(w, "Failed to create match", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(match)
}

// HandleUpdateMatchPrediction updates the prediction for a match
func (h *Handler) HandleUpdateMatchPrediction(w http.ResponseWriter, r *http.Request) {
	ctx := context.Background()

	var req struct {
		MatchID    int  `json:"match_id"`
		Prediction *int `json:"prediction"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	if err := h.DB.UpdateMatchPrediction(ctx, req.MatchID, req.Prediction); err != nil {
		http.Error(w, "Failed to update prediction", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
	w.Write([]byte(`{"message": "Prediction updated"}`))
}
