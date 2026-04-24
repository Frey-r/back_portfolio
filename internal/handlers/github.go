package handlers

import (
	"encoding/json"
	"net/http"

	"github.com/eduardobachmann/back-portfolio/internal/services"
)

type GitHubHandler struct {
	svc *services.GitHubService
}

func NewGitHubHandler(svc *services.GitHubService) *GitHubHandler {
	return &GitHubHandler{svc: svc}
}

func (h *GitHubHandler) HandleLastPush(w http.ResponseWriter, r *http.Request) {
	push, err := h.svc.GetLastPush()
	if err != nil {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]string{
			"error": err.Error(),
		})
		return
	}

	w.Header().Set("Content-Type", "application/json")

	// No push events found — return a friendly fallback
	if push == nil {
		json.NewEncoder(w).Encode(map[string]interface{}{
			"repo":      "No recent activity",
			"branch":    "",
			"message":   "Sin push recientes",
			"sha":       "",
			"author":    "",
			"pushed_at": nil,
		})
		return
	}

	json.NewEncoder(w).Encode(push)
}
