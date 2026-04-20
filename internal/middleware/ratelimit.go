package middleware

import (
	"log"
	"net"
	"net/http"
	"sync"
	"time"
)

type client struct {
	tokens int
	last   time.Time
}

// RateLimiter implementing a simple token bucket algorithm per IP.
type RateLimiter struct {
	clients map[string]*client
	mu      sync.Mutex
	limit   int
	burst   int
}

// NewRateLimiter creates a new instance.
// 'limit' is tokens per minute.
func NewRateLimiter(limit, burst int) *RateLimiter {
	rl := &RateLimiter{
		clients: make(map[string]*client),
		limit:   limit,
		burst:   burst,
	}

	// Cleanup stale IPs every hour
	go func() {
		for {
			time.Sleep(time.Hour)
			rl.mu.Lock()
			for ip, c := range rl.clients {
				if time.Since(c.last) > time.Hour {
					delete(rl.clients, ip)
				}
			}
			rl.mu.Unlock()
		}
	}()

	return rl
}

func (rl *RateLimiter) allow(ip string) bool {
	rl.mu.Lock()
	defer rl.mu.Unlock()

	c, exists := rl.clients[ip]
	now := time.Now()

	if !exists {
		rl.clients[ip] = &client{tokens: rl.burst - 1, last: now}
		return true
	}

	// Refill based on time passed
	elapsedMin := now.Sub(c.last).Minutes()
	c.tokens += int(elapsedMin * float64(rl.limit))

	if c.tokens > rl.burst {
		c.tokens = rl.burst
	}
	c.last = now

	if c.tokens > 0 {
		c.tokens--
		return true
	}

	return false
}

// Limit is the HTTP middleware
func (rl *RateLimiter) Limit(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ip := getIP(r)
		
		if !rl.allow(ip) {
			log.Printf("Rate limit exceeded for IP: %s\n", ip)
			http.Error(w, `{"success":false,"error":"Too Many Requests"}`, http.StatusTooManyRequests)
			return
		}
		
		next.ServeHTTP(w, r)
	}
}

// Helper to extract IP
func getIP(r *http.Request) string {
	ip := r.Header.Get("X-Forwarded-For")
	if ip == "" {
		ip, _, _ = net.SplitHostPort(r.RemoteAddr)
	}
	return ip
}
