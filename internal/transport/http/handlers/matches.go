package handlers

import (
	"context"
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/dancouver1/ufc-card-creator/internal/models"
	"github.com/go-chi/chi/v5"
)

// HandleCreateMatch creates a new match
//
// @Summary      Create a match
// @Description  Creates a matchup between two fighters on a card
// @Tags         matches
// @Accept       json
// @Produce      json
// @Param        match  body      models.Match  true  "Match to create (card_id, fighter1_id, fighter2_id required)"
// @Success      201    {object}  models.Match
// @Failure      400    {string}  string  "Invalid request body / missing fighters or card"
// @Failure      500    {string}  string  "Failed to create match"
// @Router       /matches [post]
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
	if err := h.Repo.Matches.CreateMatch(ctx, &match); err != nil {
		http.Error(w, "Failed to create match", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(match)
}

// UpdateMatchPredictionRequest is the body for HandleUpdateMatchPrediction.
type UpdateMatchPredictionRequest struct {
	MatchID    int  `json:"match_id"`
	Prediction *int `json:"prediction"`
}

// HandleUpdateMatchPrediction updates the prediction for a match
//
// @Summary      Update match prediction
// @Description  Sets (or clears, if prediction is null) the predicted winner for a match
// @Tags         matches
// @Accept       json
// @Produce      json
// @Param        prediction  body      UpdateMatchPredictionRequest  true  "Match ID and predicted winner's fighter ID"
// @Success      200  {string}  string  "Prediction updated"
// @Failure      400  {string}  string  "Invalid request body"
// @Failure      500  {string}  string  "Failed to update prediction"
// @Router       /matches/prediction [put]
func (h *Handler) HandleUpdateMatchPrediction(w http.ResponseWriter, r *http.Request) {
	ctx := context.Background()

	var req UpdateMatchPredictionRequest

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	if err := h.Repo.Matches.UpdateMatchPrediction(ctx, req.MatchID, req.Prediction); err != nil {
		http.Error(w, "Failed to update prediction", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
	w.Write([]byte(`{"message": "Prediction updated"}`))
}

// HandleDeleteMatch deletes a match
//
// @Summary      Delete a match
// @Description  Deletes a single match
// @Tags         matches
// @Param        id   path  int  true  "Match ID"
// @Success      200  {string}  string  "OK"
// @Failure      400  {string}  string  "Invalid match ID"
// @Failure      500  {string}  string  "Failed to delete match"
// @Router       /matches/{id} [delete]
func (h *Handler) HandleDeleteMatch(w http.ResponseWriter, r *http.Request) {
	ctx := context.Background()
	idStr := chi.URLParam(r, "id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		http.Error(w, "Invalid match ID", http.StatusBadRequest)
		return
	}

	if err := h.Repo.Matches.DeleteMatch(ctx, id); err != nil {
		http.Error(w, "Failed to delete match", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
}
