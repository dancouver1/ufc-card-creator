package handlers

import (
	"context"
	"encoding/json"
	"log"
	"net/http"
	"strconv"

	"github.com/dancouver1/ufc-card-creator/internal/models"
	"github.com/go-chi/chi/v5"
)

// HandleCardsPage renders the cards list page
func (h *Handler) HandleCardsPage(w http.ResponseWriter, r *http.Request) {
	ctx := context.Background()

	cards, err := h.Repo.Cards.GetAllCards(ctx)
	if err != nil {
		http.Error(w, "Failed to fetch cards", http.StatusInternalServerError)
		return
	}

	data := map[string]interface{}{
		"Cards": cards,
	}

	// Debug: log the number of cards
	log.Printf("Rendering cards page with %d cards", len(cards))

	err = h.Templates.ExecuteTemplate(w, "cards.html", data)
	if err != nil {
		log.Printf("Error executing cards template: %v", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}

// HandleCreateCard creates a new card
//
// @Summary      Create a card
// @Description  Creates a new fight card
// @Tags         cards
// @Accept       json
// @Produce      json
// @Param        card  body      models.Card  true  "Card to create (card_name required)"
// @Success      201   {object}  models.Card
// @Failure      400   {string}  string  "Invalid request body / card name is required"
// @Failure      500   {string}  string  "Failed to create card"
// @Router       /cards [post]
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

	if err := h.Repo.Cards.CreateCard(ctx, &card); err != nil {
		http.Error(w, "Failed to create card", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(card)
}

// HandleGetCardByID returns a card with all its matches
//
// @Summary      Get card by ID
// @Description  Returns a card along with all its matches and fighter details
// @Tags         cards
// @Produce      json
// @Param        id   path      int  true  "Card ID"
// @Success      200  {object}  models.Card
// @Failure      400  {string}  string  "Invalid card ID"
// @Failure      404  {string}  string  "Card not found"
// @Router       /cards/{id} [get]
func (h *Handler) HandleGetCardByID(w http.ResponseWriter, r *http.Request) {
	ctx := context.Background()

	idStr := chi.URLParam(r, "id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		http.Error(w, "Invalid card ID", http.StatusBadRequest)
		return
	}

	card, err := h.Repo.Cards.GetCardByID(ctx, id)
	if err != nil {
		http.Error(w, "Card not found", http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(card)
}

// HandleDeleteCard deletes a card
//
// @Summary      Delete a card
// @Description  Deletes a card and all of its matches
// @Tags         cards
// @Param        id   path  int  true  "Card ID"
// @Success      200  {string}  string  "OK"
// @Failure      400  {string}  string  "Invalid card ID"
// @Failure      500  {string}  string  "Failed to delete card"
// @Router       /cards/{id} [delete]
func (h *Handler) HandleDeleteCard(w http.ResponseWriter, r *http.Request) {
	ctx := context.Background()
	idStr := chi.URLParam(r, "id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		http.Error(w, "Invalid card ID", http.StatusBadRequest)
		return
	}

	if err := h.Repo.Cards.DeleteCard(ctx, id); err != nil {
		http.Error(w, "Failed to delete card", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
}
