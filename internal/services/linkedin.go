package services

import (
	"context"
	"log"
	"sync"
	"time"

	"github.com/eduardobachmann/back-portfolio/internal/models"
)

type LinkedInService struct {
	accessToken string
	cachedPost  *models.LinkedInPost
	lastFetch   time.Time
	mu          sync.RWMutex
}

func NewLinkedInService(token string) *LinkedInService {
	svc := &LinkedInService{
		accessToken: token,
	}
	// Initial population on startup
	svc.refreshCache()
	
	// Start background refresh routine (every 1 hour)
	go svc.startBackgroundRefresh()

	return svc
}

func (s *LinkedInService) startBackgroundRefresh() {
	ticker := time.NewTicker(1 * time.Hour)
	for range ticker.C {
		s.refreshCache()
	}
}

// GetTopPost returns the top post from memory cache
func (s *LinkedInService) GetTopPost(ctx context.Context) *models.LinkedInPost {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.cachedPost
}

func (s *LinkedInService) refreshCache() {
	// If no real token is provided, use static fallback data.
	// Implementing a full OAuth2 LinkedIn flow requires a registered App on the developer portal.
	// For this portfolio backend, a fallback is acceptable.
	if s.accessToken == "" {
		log.Println("LinkedIn access token not found. Using fallback static data for /api/linkedin/top-post")
		s.setFallbackData()
		return
	}

	// TODO: If a real token is provided, implement HTTP call to LinkedIn API
	// 1. GET https://api.linkedin.com/v2/me to get author ID (URN)
	// 2. GET https://api.linkedin.com/v2/ugcPosts?q=authors&authors={URN}
	// 3. Find post with highest engagement
	// (For now, since we usually don't have this token ready, fallback is invoked)
	
	s.setFallbackData()
}

func (s *LinkedInService) setFallbackData() {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.cachedPost = &models.LinkedInPost{
		Text:     "La arquitectura de las aplicaciones web modernas requiere un balance entre rendimiento bruto y una estructura de código mantenible. Migrar a Golang para microservicios ha reducido nuestra latencia p99 en un 60% sin sacrificar el tiempo de desarrollo. #Arquitectura #Golang #Microservicios",
		URL:      "https://linkedin.com/in/eduardobachmann", // Link to profile if post specific link not available
		Likes:    1240,
		Comments: 84,
		Date:     "1 Sem",
	}
	s.lastFetch = time.Now()
}
