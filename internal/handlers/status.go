package handlers

import (
	"net/http"
	"time"

	"github.com/eduardobachmann/back-portfolio/internal/models"
)

type StatusHandler struct{}

func NewStatusHandler() *StatusHandler {
	return &StatusHandler{}
}

func (h *StatusHandler) HandleStatus(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	status := models.StatusInfo{
		Available:   true,
		Message:     "Actualmente aceptando nuevos proyectos.",
		LastUpdated: time.Now().Format(time.RFC3339),
	}

	writeJSON(w, http.StatusOK, models.JSONResponse{
		Success: true,
		Data:    status,
	})
}
