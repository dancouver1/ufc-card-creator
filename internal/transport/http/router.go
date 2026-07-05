package http

import (
	"context"
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"
	chimiddleware "github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/httprate"

	"github.com/dancouver1/ufc-card-creator/internal/transport/http/handlers"
)

// NewRouter builds the application's chi router: middleware, page routes, API routes,
// the health check, and static file serving.
func NewRouter(h *handlers.Handler, healthCheck func(ctx context.Context) error) chi.Router {
	r := chi.NewRouter()

	// Middleware
	r.Use(chimiddleware.Logger)
	r.Use(chimiddleware.Recoverer)
	r.Use(chimiddleware.RequestID)
	r.Use(chimiddleware.RealIP)
	r.Use(chimiddleware.Timeout(60 * time.Second))

	r.Use(httprate.LimitByIP(100, time.Minute))

	// Security Headers Middleware
	r.Use(SecurityHeaders)

	// Page Routes
	r.Get("/", h.HandleHome)
	r.Get("/fighters", h.HandleFightersPage)
	r.Get("/matchmaker", h.HandleMatchmakerPage)
	r.Get("/cards", h.HandleCardsPage)
	r.Get("/card/{id}", h.HandleCardDetailPage)

	// API Routes
	r.Route("/api", func(r chi.Router) {
		// Fighters
		r.Get("/fighters", h.HandleGetFighters)
		r.Get("/fighter/{id}", h.HandleGetFighterByID)
		r.Get("/fighters/search", h.HandleSearchFighters)

		// Matches
		r.Post("/matches", h.HandleCreateMatch)
		r.Put("/matches/prediction", h.HandleUpdateMatchPrediction)
		r.Delete("/matches/{id}", h.HandleDeleteMatch)

		// Cards
		r.Post("/cards", h.HandleCreateCard)
		r.Get("/cards/{id}", h.HandleGetCardByID)
		r.Delete("/cards/{id}", h.HandleDeleteCard)
	})

	// Health check endpoint
	r.Get("/health", func(w http.ResponseWriter, r *http.Request) {
		ctx, cancel := context.WithTimeout(r.Context(), 2*time.Second)
		defer cancel()

		if err := healthCheck(ctx); err != nil {
			http.Error(w, "Database connection failed", http.StatusServiceUnavailable)
			return
		}

		w.WriteHeader(http.StatusOK)
		w.Write([]byte("OK"))
	})

	// Static files
	fileServer := http.FileServer(http.Dir("./static"))
	r.Handle("/static/*", http.StripPrefix("/static/", fileServer))

	return r
}
