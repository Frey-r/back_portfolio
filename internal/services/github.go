package services

import (
	"encoding/json"
	"fmt"
	"net/http"
	"time"
)

type GitHubService struct {
	pat        string
	login      string
	httpClient *http.Client
}

type GitHubUser struct {
	Login string `json:"login"`
}

type PushEvent struct {
	ID        string    `json:"id"`
	Type      string    `json:"type"`
	Actor     Actor     `json:"actor"`
	Repo      Repo      `json:"repo"`
	Payload   Payload   `json:"payload"`
	CreatedAt time.Time `json:"created_at"`
}

type Actor struct {
	Login string `json:"login"`
}

type Repo struct {
	Name string `json:"name"`
}

type Payload struct {
	Ref     string   `json:"ref"`
	Commits []Commit `json:"commits"`
}

type Commit struct {
	Message string `json:"message"`
	SHA     string `json:"sha"`
}

type LastPushResponse struct {
	Repo     string    `json:"repo"`
	Branch   string    `json:"branch"`
	Message  string    `json:"message"`
	SHA      string    `json:"sha"`
	Author   string    `json:"author"`
	PushedAt time.Time `json:"pushed_at"`
}

func NewGitHubService(pat string) *GitHubService {
	return &GitHubService{
		pat: pat,
		httpClient: &http.Client{
			Timeout: 10 * time.Second,
		},
	}
}

func (s *GitHubService) getAuthenticatedUser() (string, error) {
	req, err := http.NewRequest("GET", "https://api.github.com/user", nil)
	if err != nil {
		return "", err
	}
	req.Header.Set("Authorization", "Bearer "+s.pat)
	req.Header.Set("Accept", "application/vnd.github.v3+json")
	req.Header.Set("User-Agent", "Portfolio-Backend")

	resp, err := s.httpClient.Do(req)
	if err != nil {
		return "", fmt.Errorf("github user request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("github user API returned status %d", resp.StatusCode)
	}

	var user GitHubUser
	if err := json.NewDecoder(resp.Body).Decode(&user); err != nil {
		return "", fmt.Errorf("failed to decode user response: %w", err)
	}

	return user.Login, nil
}

func (s *GitHubService) GetLastPush() (*LastPushResponse, error) {
	if s.pat == "" {
		return nil, fmt.Errorf("GITHUB_PAT not configured")
	}

	// Get username dynamically if not cached
	if s.login == "" {
		login, err := s.getAuthenticatedUser()
		if err != nil {
			return nil, err
		}
		s.login = login
	}

	// Try 1: Get authenticated user's public events
	result, err := s.getLastPushFromEvents()
	if err == nil && result != nil {
		return result, nil
	}

	// Try 2: Fallback to repos API (includes private repos)
	return s.getLastPushFromRepos()
}

func (s *GitHubService) getLastPushFromEvents() (*LastPushResponse, error) {
	url := fmt.Sprintf("https://api.github.com/users/%s/events/public", s.login)

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, err
	}

	req.Header.Set("Authorization", "Bearer "+s.pat)
	req.Header.Set("Accept", "application/vnd.github.v3+json")
	req.Header.Set("User-Agent", "Portfolio-Backend")

	resp, err := s.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("github events API request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("github events API returned status %d", resp.StatusCode)
	}

	var events []PushEvent
	if err := json.NewDecoder(resp.Body).Decode(&events); err != nil {
		return nil, fmt.Errorf("failed to decode events response: %w", err)
	}

	// Find the most recent push event from any repo
	for _, event := range events {
		if event.Type == "PushEvent" {
			if len(event.Payload.Commits) > 0 {
				lastCommit := event.Payload.Commits[len(event.Payload.Commits)-1]
				return &LastPushResponse{
					Repo:     event.Repo.Name,
					Branch:   event.Payload.Ref,
					Message:  lastCommit.Message,
					SHA:      lastCommit.SHA[:7],
					Author:   event.Actor.Login,
					PushedAt: event.CreatedAt,
				}, nil
			}
		}
	}

	return nil, nil
}

// GitHubRepo represents a repository from the GitHub API
type GitHubRepo struct {
	FullName  string    `json:"full_name"`
	PushedAt  time.Time `json:"pushed_at"`
	Private   bool      `json:"private"`
	DefaultBranch string `json:"default_branch"`
}

func (s *GitHubService) getLastPushFromRepos() (*LastPushResponse, error) {
	// Get user's repos sorted by most recently pushed
	url := "https://api.github.com/user/repos?sort=pushed&direction=desc&per_page=1"

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, err
	}

	req.Header.Set("Authorization", "Bearer "+s.pat)
	req.Header.Set("Accept", "application/vnd.github.v3+json")
	req.Header.Set("User-Agent", "Portfolio-Backend")

	resp, err := s.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("github repos API request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("github repos API returned status %d", resp.StatusCode)
	}

	var repos []GitHubRepo
	if err := json.NewDecoder(resp.Body).Decode(&repos); err != nil {
		return nil, fmt.Errorf("failed to decode repos response: %w", err)
	}

	if len(repos) == 0 {
		return nil, nil
	}

	repo := repos[0]
	return &LastPushResponse{
		Repo:     repo.FullName,
		Branch:   repo.DefaultBranch,
		Message:  "Último push detectado",
		SHA:      "",
		Author:   s.login,
		PushedAt: repo.PushedAt,
	}, nil
}
