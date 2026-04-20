package handlers

import (
	"net/http"
	"time"

	"github.com/eduardobachmann/back-portfolio/internal/database"
	"github.com/eduardobachmann/back-portfolio/internal/models"
)

var startTime = time.Now()

type HealthHandler struct {
	db *database.DB
}

func NewHealthHandler(db *database.DB) *HealthHandler {
	return &HealthHandler{db: db}
}

func (h *HealthHandler) HandleHealth(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Double check DB connection
	err := h.db.Ping()
	dbStatus := "ok"
	if err != nil {
		dbStatus = "error"
	}

	uptime := time.Since(startTime)
	
	data := map[string]interface{}{
		"status":   "ok",
		"database": dbStatus,
		"uptime":   uptime.String(),
		"time":     time.Now().Format(time.RFC3339),
	}

	writeJSON(w, http.StatusOK, models.JSONResponse{
		Success: true,
		Data:    data,
	})
}

