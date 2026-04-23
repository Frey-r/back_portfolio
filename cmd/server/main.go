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

	// 2. Initialize Database
	db, err := database.New(cfg.SQLitePath)
	if err != nil {
		log.Fatalf("Critical error: Could not initialize database: %v", err)
	}
	defer db.Close()
	log.Println("Database initialized successfully")

	// Initialize Services
	contactSvc := services.NewContactService(db)
	authSvc := services.NewAuthService(db)
	linkedinSvc := services.NewLinkedInService(cfg, authSvc)
	githubSvc := services.NewGitHubService(cfg.GitHubPAT)

	// Initialize Handlers
	healthHandler := handlers.NewHealthHandler(db)
	statusHandler := handlers.NewStatusHandler()
	contactHandler := handlers.NewContactHandler(contactSvc)
	linkedinHandler := handlers.NewLinkedInHandler(linkedinSvc)
	authHandler := handlers.NewAuthHandler(cfg, authSvc)
	githubHandler := handlers.NewGitHubHandler(githubSvc)

	// Initialize Rate Limiter
	contactRateLimiter := middleware.NewRateLimiter(cfg.RateLimitContact, cfg.RateLimitContact)

	// --- Routes ---
	mux := http.NewServeMux()

	// Health & Status
	mux.HandleFunc("GET /api/health", healthHandler.HandleHealth)
	mux.HandleFunc("GET /api/status", statusHandler.HandleStatus)

	// LinkedIn
	mux.HandleFunc("GET /api/linkedin/top-post", linkedinHandler.HandleTopPost)

	// Auth (LinkedIn OAuth 2.0)
	mux.HandleFunc("GET /api/auth/linkedin/login", authHandler.HandleLinkedInLogin)
	mux.HandleFunc("GET /api/auth/linkedin/callback", authHandler.HandleLinkedInCallback)

	// GitHub (Rate Limited)
	mux.HandleFunc("GET /api/github/last-push", contactRateLimiter.Limit(githubHandler.HandleLastPush))

	// Contact Form (Rate Limited)
	mux.HandleFunc("POST /api/contact", contactRateLimiter.Limit(contactHandler.HandleSubmit))

	// 7. Apply Global Middleware
	handler := middleware.CORS(cfg.AllowedOrigins)(mux)
	handler = middleware.Logger(handler)

	// 8. Server Configuration
	server := &http.Server{
		Addr:    cfg.Port,
		Handler: handler,

		ReadTimeout:  15 * time.Second,
		WriteTimeout: 15 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	// 9. Graceful Shutdown Setup
	done := make(chan os.Signal, 1)
	signal.Notify(done, os.Interrupt, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		log.Printf("Starting Portfolio Backend on port :%s", cfg.Port)
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("Server failed: %v", err)
		}
	}()

	// Wait for termination signal
	<-done
	log.Println("Gracefully shutting down...")

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	if err := server.Shutdown(ctx); err != nil {
		log.Fatalf("Server forced to shutdown: %v", err)
	}

	log.Println("Server exited properly")
}
