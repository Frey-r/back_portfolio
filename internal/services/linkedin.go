package services

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"net/url"
	"sync"
	"time"

	"github.com/eduardobachmann/back-portfolio/internal/config"
	"github.com/eduardobachmann/back-portfolio/internal/models"
)

type LinkedInService struct {
	cfg     *config.Config
	authSvc *AuthService
	cache   *models.LinkedInPost
	mu      sync.RWMutex
}

func NewLinkedInService(cfg *config.Config, authSvc *AuthService) *LinkedInService {
	s := &LinkedInService{
		cfg:     cfg,
		authSvc: authSvc,
	}
	// Start background refresh
	go s.startBackgroundRefresh()
	return s
}

func (s *LinkedInService) GetTopPost(ctx context.Context) *models.LinkedInPost {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.cache
}

func (s *LinkedInService) startBackgroundRefresh() {
	// Refresh immediately on start
	s.refreshCache()

	// Then every hour
	ticker := time.NewTicker(1 * time.Hour)
	for range ticker.C {
		s.refreshCache()
	}
}

func (s *LinkedInService) refreshCache() {
	ctx := context.Background()
	
	// 1. Try to get token from DB
	token, err := s.authSvc.GetLinkedInToken(ctx)
	
	accessToken := ""
	userURN := ""
	
	if err == nil && token != nil {
		accessToken = token.AccessToken
		userURN = token.UserURN
	} else if s.cfg.LinkedInAccessToken != "" {
		// 2. Fallback to manual token from .env
		accessToken = s.cfg.LinkedInAccessToken
		log.Println("Using manual LinkedIn access token from .env")
	}

	if accessToken == "" {
		log.Println("LinkedIn access token not found. Using static data.")
		s.setFallbackData()
		return
	}

	// 3. Fetch real data
	post, err := s.fetchTopPostFromAPI(accessToken, userURN)
	if err != nil {
		log.Printf("Failed to fetch LinkedIn data: %v. Using static data.", err)
		s.setFallbackData()
		return
	}

	s.mu.Lock()
	s.cache = post
	s.mu.Unlock()
	log.Printf("LinkedIn cache refreshed: Found top post with %d likes", post.Likes)
}

func (s *LinkedInService) fetchTopPostFromAPI(token, urn string) (*models.LinkedInPost, error) {
	// If we don't have a URN, get it first via /v2/userinfo (OpenID Connect)
	if urn == "" {
		req, _ := http.NewRequest("GET", "https://api.linkedin.com/v2/userinfo", nil)
		req.Header.Set("Authorization", "Bearer "+token)
		client := &http.Client{}
		resp, err := client.Do(req)
		if err == nil && resp.StatusCode == 200 {
			var user models.LinkedInUser
			json.NewDecoder(resp.Body).Decode(&user)
			urn = user.Sub
			resp.Body.Close()
		}
	}

	// Fetch real posts
	// NOTE: LinkedIn API versioning and scopes change often. 
	// We use /v2/posts which is the modern standard for personal posts.
	apiURL := fmt.Sprintf("https://api.linkedin.com/v2/posts?author=%s&q=author&count=5", url.QueryEscape(urn))
	req, _ := http.NewRequest("GET", apiURL, nil)
	req.Header.Set("Authorization", "Bearer "+token)
	req.Header.Set("Linkedin-Version", "202404")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		return nil, fmt.Errorf("api returned status %d", resp.StatusCode)
	}

	var result struct {
		Elements []struct {
			ID         string `json:"id"`
			Commentary string `json:"commentary"`
			CreatedAt  int64  `json:"createdAt"`
		} `json:"elements"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, err
	}

	if len(result.Elements) == 0 {
		return nil, fmt.Errorf("no posts found")
	}

	// For MVP, select the first one (most recent)
	// REAL engagement metrics would require calls to /v2/socialActions/{urn}
	bestPost := result.Elements[0]
	
	return &models.LinkedInPost{
		ID:       bestPost.ID,
		Text:     bestPost.Commentary,
		URL:      fmt.Sprintf("https://www.linkedin.com/feed/update/%s", bestPost.ID),
		Likes:    124, // Static for now as extra permissions are needed for SocialActions
		Comments: 12,
		Date:     time.Unix(bestPost.CreatedAt/1000, 0).Format("2006-01-02"),
	}, nil
}

func (s *LinkedInService) setFallbackData() {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.cache = &models.LinkedInPost{
		ID:       "fallback",
		Text:     "La arquitectura de las aplicaciones web modernas requiere un balance entre rendimiento bruto y una estructura de código mantenible. Migrar a Golang para microservicios ha reducido nuestra latencia p99 en un 60% sin sacrificar el tiempo de desarrollo. #Arquitectura #Golang #Microservicios",
		URL:      "https://linkedin.com/in/eduardobachmann",
		Likes:    1240,
		Comments: 84,
		Date:     "1 Sem",
	}
}
