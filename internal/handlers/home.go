package handlers

import (
	"log"
	"net/http"
)

// HandleHome renders the home page
func (h *Handler) HandleHome(w http.ResponseWriter, r *http.Request) {
	err := h.Templates.ExecuteTemplate(w, "index.html", nil)
	if err != nil {
		log.Printf("Error rendering home: %v", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}
