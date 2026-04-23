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
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]string{
			"error": err.Error(),
		})
		return
	}
	
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(push)
}
