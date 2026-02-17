package handlers

import (
	"context"
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/dancouver1/ufc-card-creator/internal/models"
	"github.com/go-chi/chi/v5"
)

// HandleFightersPage renders the fighters list page
func (h *Handler) HandleFightersPage(w http.ResponseWriter, r *http.Request) {
	ctx := context.Background()

	searchTerm := r.URL.Query().Get("search")
	weightClass := r.URL.Query().Get("weight_class")

	fighters, err := h.DB.SearchFightersWithFilters(ctx, searchTerm, weightClass)
	if err != nil {
		http.Error(w, "Failed to fetch fighters", http.StatusInternalServerError)
		return
	}

	data := map[string]interface{}{
		"Fighters":    fighters,
		"SearchTerm":  searchTerm,
		"WeightClass": weightClass,
	}

	err = h.Templates.ExecuteTemplate(w, "fighters.html", data)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}

// HandleGetFighters returns fighters as JSON (for API calls)
func (h *Handler) HandleGetFighters(w http.ResponseWriter, r *http.Request) {
	ctx := context.Background()

	searchTerm := r.URL.Query().Get("search")
	weightClass := r.URL.Query().Get("weight_class")

	fighters, err := h.DB.SearchFightersWithFilters(ctx, searchTerm, weightClass)

	if err != nil {
		http.Error(w, "Failed to fetch fighters", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(fighters)
}

// HandleGetFighterByID returns a single fighter with last 5 fights
func (h *Handler) HandleGetFighterByID(w http.ResponseWriter, r *http.Request) {
	ctx := context.Background()

	idStr := chi.URLParam(r, "id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		http.Error(w, "Invalid fighter ID", http.StatusBadRequest)
		return
	}

	fighter, err := h.DB.GetFighterByID(ctx, id)
	if err != nil {
		http.Error(w, "Fighter not found", http.StatusNotFound)
		return
	}

	// Get last 5 fights
	fights, err := h.DB.GetLast5Fights(ctx, id)
	if err != nil {
		// Continue even if fights fetch fails
		fights = []models.FightHistory{}
	}
	fighter.Last5Fights = fights

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(fighter)
}

// HandleSearchFighters returns filtered HTML for HTMX
func (h *Handler) HandleSearchFighters(w http.ResponseWriter, r *http.Request) {
	ctx := context.Background()

	searchTerm := r.URL.Query().Get("search")
	weightClass := r.URL.Query().Get("weight_class")

	fighters, err := h.DB.SearchFightersWithFilters(ctx, searchTerm, weightClass)
	if err != nil {
		http.Error(w, "Search failed", http.StatusInternalServerError)
		return
	}

	// Return just the fighters grid HTML (partial template)
	// For now, we'll return JSON - you can create a partial template later
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(fighters)
}
