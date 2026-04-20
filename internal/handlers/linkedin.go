package handlers

import (
	"net/http"

	"github.com/eduardobachmann/back-portfolio/internal/models"
	"github.com/eduardobachmann/back-portfolio/internal/services"
)

type LinkedInHandler struct {
	svc *services.LinkedInService
}

func NewLinkedInHandler(svc *services.LinkedInService) *LinkedInHandler {
	return &LinkedInHandler{svc: svc}
}

// HandleTopPost returns the top LinkedIn post data
func (h *LinkedInHandler) HandleTopPost(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	post := h.svc.GetTopPost(r.Context())
	
	writeJSON(w, http.StatusOK, models.JSONResponse{
		Success: true,
		Data:    post,
	})
}
