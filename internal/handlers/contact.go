package handlers

import (
	"encoding/json"
	"log"
	"net/http"
	"regexp"
	"strings"

	"github.com/eduardobachmann/back-portfolio/internal/models"
	"github.com/eduardobachmann/back-portfolio/internal/services"
)

type ContactHandler struct {
	svc *services.ContactService
}

func NewContactHandler(svc *services.ContactService) *ContactHandler {
	return &ContactHandler{svc: svc}
}

func (h *ContactHandler) HandleSubmit(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req models.ContactRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSON(w, http.StatusBadRequest, models.JSONResponse{Success: false, Error: "Invalid JSON format"})
		return
	}

	// Basic Validation
	req.Name = strings.TrimSpace(req.Name)
	req.Email = strings.TrimSpace(req.Email)
	req.Message = strings.TrimSpace(req.Message)

	if req.Name == "" || req.Message == "" {
		writeJSON(w, http.StatusBadRequest, models.JSONResponse{Success: false, Error: "Name and message are required"})
		return
	}

	if !isValidEmail(req.Email) {
		writeJSON(w, http.StatusBadRequest, models.JSONResponse{Success: false, Error: "Invalid email format"})
		return
	}

	// Extract context metadata
	ip := r.Header.Get("X-Forwarded-For")
	if ip == "" {
		ip = r.RemoteAddr
	}
	userAgent := r.UserAgent()

	// Save to DB
	id, err := h.svc.Create(r.Context(), &req, ip, userAgent)
	if err != nil {
		log.Printf("Failed to insert contact: %v\n", err)
		writeJSON(w, http.StatusInternalServerError, models.JSONResponse{Success: false, Error: "Internal server error"})
		return
	}

	// TODO: Send notification email asynchronously if configured

	writeJSON(w, http.StatusCreated, models.JSONResponse{
		Success: true,
		Data: map[string]interface{}{
			"id":      id,
			"message": "Contact received successfully",
		},
	})
}

// Helpers
var emailRegex = regexp.MustCompile(`^[a-z0-9._%+\-]+@[a-z0-9.\-]+\.[a-z]{2,4}$`)

func isValidEmail(e string) bool {
	return emailRegex.MatchString(strings.ToLower(e))
}

func writeJSON(w http.ResponseWriter, status int, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	if err := json.NewEncoder(w).Encode(data); err != nil {
		log.Printf("Failed to encode JSON: %v\n", err)
	}
}
