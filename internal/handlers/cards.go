package handlers

import (
	"context"
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"
	"github.com/dancouver1/ufc-card-creator/internal/models"
)

// HandleCardsPage renders the cards list page
func (h *Handler) HandleCardsPage(w http.ResponseWriter, r *http.Request) {
	ctx := context.Background()

	cards, err := h.DB.GetAllCards(ctx)
	if err != nil {
		http.Error(w, "Failed to fetch cards", http.StatusInternalServerError)
		return
	}

	data := map[string]interface{}{
		"Cards": cards,
	}

	err = h.Templates.ExecuteTemplate(w, "cards.html", data)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}

// HandleCreateCard creates a new card
func (h *Handler) HandleCreateCard(w http.ResponseWriter, r *http.Request) {
	ctx := context.Background()

	var card models.Card
	if err := json.NewDecoder(r.Body).Decode(&card); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	if card.CardName == "" {
		http.Error(w, "Card name is required", http.StatusBadRequest)
		return
	}

	if err := h.DB.CreateCard(ctx, &card); err != nil {
		http.Error(w, "Failed to create card", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(card)
}

// HandleGetCardByID returns a card with all its matches
func (h *Handler) HandleGetCardByID(w http.ResponseWriter, r *http.Request) {
	ctx := context.Background()

	idStr := chi.URLParam(r, "id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		http.Error(w, "Invalid card ID", http.StatusBadRequest)
		return
	}

	card, err := h.DB.GetCardByID(ctx, id)
	if err != nil {
		http.Error(w, "Card not found", http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(card)
}
