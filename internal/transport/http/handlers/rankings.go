package handlers

import (
	"context"
	"net/http"

	"github.com/dancouver1/ufc-card-creator/internal/models"
)

// HandleRankingsPage renders the UFC rankings page with both the media-panel
// and META (AI) rankings tables; the client-side toggle switches between
// the two without a page reload.
func (h *Handler) HandleRankingsPage(w http.ResponseWriter, r *http.Request) {
	ctx := context.Background()

	media, err := h.Repo.Rankings.GetByType(ctx, models.RankTypeMedia)
	if err != nil {
		http.Error(w, "Failed to fetch media rankings", http.StatusInternalServerError)
		return
	}

	meta, err := h.Repo.Rankings.GetByType(ctx, models.RankTypeMeta)
	if err != nil {
		http.Error(w, "Failed to fetch meta rankings", http.StatusInternalServerError)
		return
	}

	lastUpdated, _ := h.Repo.Rankings.LastUpdated(ctx)

	data := map[string]interface{}{
		"MediaDivisions": groupByDivision(media),
		"MetaDivisions":  groupByDivision(meta),
		"LastUpdated":    lastUpdated,
	}

	if err := h.Templates.ExecuteTemplate(w, "rankings.html", data); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}

// groupByDivision turns a flat, division/position-ordered ranking list into
// one DivisionRankings per division, splitting off the champion (position 0).
func groupByDivision(rankings []models.Ranking) []models.DivisionRankings {
	var divisions []models.DivisionRankings
	var current *models.DivisionRankings

	for i := range rankings {
		rk := rankings[i]
		if current == nil || current.Division != rk.Division {
			divisions = append(divisions, models.DivisionRankings{Division: rk.Division})
			current = &divisions[len(divisions)-1]
		}
		if rk.IsChampion {
			champ := rk
			current.Champion = &champ
			continue
		}
		current.Contenders = append(current.Contenders, rk)
	}

	return divisions
}
