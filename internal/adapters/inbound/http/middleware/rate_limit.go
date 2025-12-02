package middleware

import (
	"net/http"
	"sync"
	"time"

	"golang.org/x/time/rate"
)

// RateLimiter provides rate limiting per IP address
type RateLimiter struct {
	limiters map[string]*rate.Limiter
	mu       sync.RWMutex
	rps      rate.Limit
	burst    int
}

// NewRateLimiter creates a new rate limiter
func NewRateLimiter(requestsPerSecond float64, burst int) *RateLimiter {
	rl := &RateLimiter{
		limiters: make(map[string]*rate.Limiter),
		rps:      rate.Limit(requestsPerSecond),
		burst:    burst,
	}
	
	// Start cleanup goroutine
	go rl.cleanupStaleEntries()
	
	return rl
}

// getLimiter returns the rate limiter for a specific key (e.g., IP address)
func (rl *RateLimiter) getLimiter(key string) *rate.Limiter {
	rl.mu.RLock()
	limiter, exists := rl.limiters[key]
	rl.mu.RUnlock()
	
	if exists {
		return limiter
	}
	
	// Create new limiter
	rl.mu.Lock()
	defer rl.mu.Unlock()
	
	// Double-check after acquiring write lock
	if limiter, exists = rl.limiters[key]; exists {
		return limiter
	}
	
	limiter = rate.NewLimiter(rl.rps, rl.burst)
	rl.limiters[key] = limiter
	
	return limiter
}

// cleanupStaleEntries removes inactive limiters periodically
func (rl *RateLimiter) cleanupStaleEntries() {
	ticker := time.NewTicker(1 * time.Minute)
	defer ticker.Stop()
	
	for range ticker.C {
		rl.mu.Lock()
		for key, limiter := range rl.limiters {
			// Remove if no tokens have been consumed recently
			if limiter.Tokens() == float64(rl.burst) {
				delete(rl.limiters, key)
			}
		}
		rl.mu.Unlock()
	}
}

// RateLimit middleware limits requests per IP address
func RateLimit(limiter *RateLimiter) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// Use IP address as key
			ip := r.RemoteAddr
			
			// Get limiter for this IP
			ipLimiter := limiter.getLimiter(ip)
			
			// Check if request is allowed
			if !ipLimiter.Allow() {
				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(http.StatusTooManyRequests)
				w.Write([]byte(`{"error":"Rate limit exceeded","code":"RATE_LIMIT_EXCEEDED"}`))
				return
			}
			
			next.ServeHTTP(w, r)
		})
	}
}

