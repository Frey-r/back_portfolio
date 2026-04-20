package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/eduardobachmann/back-portfolio/internal/config"
	"github.com/eduardobachmann/back-portfolio/internal/database"
	"github.com/eduardobachmann/back-portfolio/internal/handlers"
	"github.com/eduardobachmann/back-portfolio/internal/middleware"
	"github.com/eduardobachmann/back-portfolio/internal/services"
)

func main() {
	// 1. Load Configuration
	cfg := config.Load()
	log.Printf("Starting Portfolio Backend on port %s", cfg.Port)

	// Ensure data directory exists
	if err := os.MkdirAll("./data", 0755); err != nil {
		log.Fatalf("Failed to create data directory: %v", err)
	}

	// 2. Initialize Database
	db, err := database.InitDB(cfg.SQLitePath)
	if err != nil {
		log.Fatalf("Database initialization failed: %v", err)
	}
	defer db.Close()
	log.Println("Database initialized successfully")

	// 3. Initialize Services
	contactSvc := services.NewContactService(db)
	linkedinSvc := services.NewLinkedInService(cfg.LinkedInAccessToken)

	// 4. Initialize Handlers
	contactHandler := handlers.NewContactHandler(contactSvc)
	linkedinHandler := handlers.NewLinkedInHandler(linkedinSvc)

	// 5. Initialize Rate Limiter
	rateLimiter := middleware.NewRateLimiter(cfg.RateLimitContact, cfg.RateLimitContact) // burst = limit
	
	// 6. Setup Router
	// Using Go 1.22+ enhanced ServeMux
	mux := http.NewServeMux()

	// Public Routes
	mux.HandleFunc("GET /api/health", handlers.HandleHealth)
	mux.HandleFunc("GET /api/status", handlers.HandleStatus)
	mux.HandleFunc("GET /api/linkedin/top-post", linkedinHandler.HandleTopPost)

	// Rate Limited Routes
	mux.HandleFunc("POST /api/contact", rateLimiter.Limit(contactHandler.HandleSubmit))

	// 7. Apply Global Middleware
	handler := middleware.CORS(cfg.AllowedOrigins)(mux)
	handler = middleware.Logger(handler) // Logger runs first (wraps CORS)

	// 8. Server Setup with Graceful Shutdown
	srv := &http.Server{
		Addr:         cfg.Port,
		Handler:      handler,
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 10 * time.Second,
		IdleTimeout:  120 * time.Second,
	}

	// Channel to listen for errors coming from the listener
	serverErrors := make(chan error, 1)

	// Start server
	go func() {
		log.Printf("Server listening on %s", srv.Addr)
		serverErrors <- srv.ListenAndServe()
	}()

	// Channel to listen for an interrupt or terminate signal from the OS
	osSignals := make(chan os.Signal, 1)
	signal.Notify(osSignals, os.Interrupt, syscall.SIGTERM)

	// Block until a signal is received
	select {
	case err := <-serverErrors:
		log.Fatalf("Error starting server: %v", err)

	case sig := <-osSignals:
		log.Printf("Received signal: %v. Starting graceful shutdown...", sig)

		// Create a deadline to wait for
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()

		if err := srv.Shutdown(ctx); err != nil {
			log.Printf("Graceful shutdown did not complete in time: %v", err)
			if err := srv.Close(); err != nil {
				log.Fatalf("Could not stop server gracefully: %v", err)
			}
		}
	}

	log.Println("Server stopped")
}
