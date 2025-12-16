package handlers

import (
	"net/http"
)

// HandleHome renders the home page
func (h *Handler) HandleHome(w http.ResponseWriter, r *http.Request) {
	err := h.Templates.ExecuteTemplate(w, "index.html", nil)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}
