package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/joho/godotenv"

	"github.com/dancouver1/ufc-card-creator/internal/database"
	"github.com/dancouver1/ufc-card-creator/internal/handlers"
)

func main() {
	// Load .env file (only works locally, not in Docker)
	_ = godotenv.Load()

	// Get port from environment
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	// Initialize database connection
	db, err := database.NewDB()
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}
	defer db.Close()

	// Initialize handlers
	h := handlers.NewHandler(db)

	// Initialize router
	r := chi.NewRouter()

	// Middleware
	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)
	r.Use(middleware.RequestID)
	r.Use(middleware.RealIP)
	r.Use(middleware.Timeout(60 * time.Second))

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

		if err := db.Health(ctx); err != nil {
			http.Error(w, "Database connection failed", http.StatusServiceUnavailable)
			return
		}

		w.WriteHeader(http.StatusOK)
		w.Write([]byte("OK"))
	})

	// Static files
	fileServer := http.FileServer(http.Dir("./static"))
	r.Handle("/static/*", http.StripPrefix("/static/", fileServer))

	// Create server
	server := &http.Server{
		Addr:         ":" + port,
		Handler:      r,
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 15 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	// Start server in a goroutine
	go func() {
		log.Printf("Server starting on port %s...", port)
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("Server failed to start: %v", err)
		}
	}()

	// Graceful shutdown
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	log.Println("Shutting down server...")

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	if err := server.Shutdown(ctx); err != nil {
		log.Fatalf("Server forced to shutdown: %v", err)
	}

	log.Println("Server exited")
}
