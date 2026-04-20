package handlers

import (
	"net/http"

	"github.com/eduardobachmann/back-portfolio/internal/models"
)

func HandleStatus(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	status := models.StatusInfo{
		Available:   true,
		Message:     "Aceptando nuevos proyectos para Q3 2025",
		AcceptingBy: "2025-07-01",
	}

	writeJSON(w, http.StatusOK, models.JSONResponse{
		Success: true,
		Data:    status,
	})
}
