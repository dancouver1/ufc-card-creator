package handlers

import (
	"net/http"
)

// HandleMatchmakerPage renders the matchmaker page
func (h *Handler) HandleMatchmakerPage(w http.ResponseWriter, r *http.Request) {
	err := h.Templates.ExecuteTemplate(w, "matchmaker.html", nil)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}
