package handlers

import (
	"context"
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"
)

// HandleCardDetailPage renders the card detail page
func (h *Handler) HandleCardDetailPage(w http.ResponseWriter, r *http.Request) {
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

	data := map[string]interface{}{
		"Card": card,
	}

	err = h.Templates.ExecuteTemplate(w, "card_detail.html", data)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}
