package handlers

import (
    "encoding/json"
    "html/template"
    "net/http"
    
    "github.com/dancouver1/ufc-card-creator/internal/database"
)

type Handler struct {
    DB        *database.DB
    Templates *template.Template
}

// NewHandler creates a new handler instance
func NewHandler(db *database.DB) *Handler {
    // Parse all templates - no custom functions needed for now
    tmpl := template.Must(template.ParseGlob("templates/*.html"))
    
    return &Handler{
        DB:        db,
        Templates: tmpl,
    }
}

// respondWithError sends an error response
func (h *Handler) respondWithError(w http.ResponseWriter, code int, message string) {
    w.WriteHeader(code)
    w.Write([]byte(message))
}

// respondWithJSON sends a JSON response
func (h *Handler) respondWithJSON(w http.ResponseWriter, code int, payload interface{}) {
    w.Header().Set("Content-Type", "application/json")
    w.WriteHeader(code)
    json.NewEncoder(w).Encode(payload)
}