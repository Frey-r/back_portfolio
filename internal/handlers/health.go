package handlers

import (
	"net/http"
	"time"

	"github.com/eduardobachmann/back-portfolio/internal/models"
)

var startTime = time.Now()

func HandleHealth(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	uptime := time.Since(startTime)
	
	data := map[string]interface{}{
		"status": "ok",
		"uptime": uptime.String(),
		"time":   time.Now().Format(time.RFC3339),
	}

	writeJSON(w, http.StatusOK, models.JSONResponse{
		Success: true,
		Data:    data,
	})
}
