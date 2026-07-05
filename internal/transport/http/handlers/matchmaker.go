package handlers

import (
	"context"
	"net/http"
)

// HandleMatchmakerPage renders the matchmaker page
func (h *Handler) HandleMatchmakerPage(w http.ResponseWriter, r *http.Request) {
	ctx := context.Background()

	// Get all cards for the dropdown
	cards, err := h.Repo.Cards.GetAllCards(ctx)
	if err != nil {
		http.Error(w, "Failed to fetch cards", http.StatusInternalServerError)
		return
	}

	data := map[string]interface{}{
		"Cards": cards,
	}

	err = h.Templates.ExecuteTemplate(w, "matchmaker.html", data)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}
