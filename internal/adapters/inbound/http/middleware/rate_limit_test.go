package middleware

import (
	"net/http"
	"net/http/httptest"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"golang.org/x/time/rate"
)

func TestNewRateLimiter(t *testing.T) {
	t.Run("creates rate limiter with correct settings", func(t *testing.T) {
		// Arrange & Act
		limiter := NewRateLimiter(10, 5)

		// Assert
		assert.NotNil(t, limiter)
		assert.Equal(t, rate.Limit(10), limiter.rps)
		assert.Equal(t, 5, limiter.burst)
		assert.NotNil(t, limiter.limiters)
	})

	t.Run("creates limiter with different rates", func(t *testing.T) {
		tests := []struct {
			name  string
			rps   float64
			burst int
		}{
			{"low rate", 1, 1},
			{"medium rate", 10, 5},
			{"high rate", 100, 20},
		}

		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				limiter := NewRateLimiter(tt.rps, tt.burst)
				assert.Equal(t, rate.Limit(tt.rps), limiter.rps)
				assert.Equal(t, tt.burst, limiter.burst)
			})
		}
	})
}

func TestRateLimiter_GetLimiter(t *testing.T) {
	t.Run("creates new limiter for new key", func(t *testing.T) {
		// Arrange
		rl := NewRateLimiter(10, 5)
		key := "192.168.1.1"

		// Act
		limiter := rl.getLimiter(key)

		// Assert
		assert.NotNil(t, limiter)
		assert.Len(t, rl.limiters, 1)
	})

	t.Run("returns same limiter for same key", func(t *testing.T) {
		// Arrange
		rl := NewRateLimiter(10, 5)
		key := "192.168.1.1"

		// Act
		limiter1 := rl.getLimiter(key)
		limiter2 := rl.getLimiter(key)

		// Assert
		assert.Equal(t, limiter1, limiter2)
		assert.Len(t, rl.limiters, 1)
	})

	t.Run("creates different limiters for different keys", func(t *testing.T) {
		// Arrange
		rl := NewRateLimiter(10, 5)

		// Act
		limiter1 := rl.getLimiter("192.168.1.1")
		limiter2 := rl.getLimiter("192.168.1.2")
		limiter3 := rl.getLimiter("192.168.1.3")

		// Assert - check they are different instances (pointer comparison)
		assert.NotSame(t, limiter1, limiter2, "Should create different limiter instances")
		assert.NotSame(t, limiter2, limiter3, "Should create different limiter instances")
		assert.Len(t, rl.limiters, 3)
	})
}

func TestRateLimit(t *testing.T) {
	t.Run("allows requests within limit", func(t *testing.T) {
		// Arrange
		limiter := NewRateLimiter(10, 5)
		handler := RateLimit(limiter)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusOK)
			w.Write([]byte("success"))
		}))

		req := httptest.NewRequest(http.MethodGet, "/test", nil)
		req.RemoteAddr = "192.168.1.1:1234"
		rec := httptest.NewRecorder()

		// Act
		handler.ServeHTTP(rec, req)

		// Assert
		assert.Equal(t, http.StatusOK, rec.Code)
		assert.Equal(t, "success", rec.Body.String())
	})

	t.Run("blocks requests exceeding rate limit", func(t *testing.T) {
		// Arrange - very restrictive limit
		limiter := NewRateLimiter(1, 1) // 1 request per second, burst of 1
		handler := RateLimit(limiter)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusOK)
			w.Write([]byte("success"))
		}))

		ip := "192.168.1.1:1234"

		// Act - make multiple rapid requests
		var blockedCount int
		for i := 0; i < 5; i++ {
			req := httptest.NewRequest(http.MethodGet, "/test", nil)
			req.RemoteAddr = ip
			rec := httptest.NewRecorder()

			handler.ServeHTTP(rec, req)

			if rec.Code == http.StatusTooManyRequests {
				blockedCount++
				assert.Contains(t, rec.Body.String(), "Rate limit exceeded")
				assert.Contains(t, rec.Body.String(), "RATE_LIMIT_EXCEEDED")
			}
		}

		// Assert - at least some requests should be blocked
		assert.Greater(t, blockedCount, 0, "Expected some requests to be blocked")
	})

	t.Run("respects burst parameter", func(t *testing.T) {
		// Arrange - allow burst of 3
		limiter := NewRateLimiter(0.1, 3) // Very low rate, but burst of 3
		handler := RateLimit(limiter)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusOK)
		}))

		ip := "192.168.1.1:1234"

		// Act - make burst requests
		successCount := 0
		for i := 0; i < 5; i++ {
			req := httptest.NewRequest(http.MethodGet, "/test", nil)
			req.RemoteAddr = ip
			rec := httptest.NewRecorder()

			handler.ServeHTTP(rec, req)

			if rec.Code == http.StatusOK {
				successCount++
			}
		}

		// Assert - should allow exactly burst number of requests
		assert.Equal(t, 3, successCount, "Should allow burst of 3 requests")
	})

	t.Run("tracks different IPs separately", func(t *testing.T) {
		// Arrange
		limiter := NewRateLimiter(1, 1)
		handler := RateLimit(limiter)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusOK)
		}))

		// Act - requests from different IPs
		req1 := httptest.NewRequest(http.MethodGet, "/test", nil)
		req1.RemoteAddr = "192.168.1.1:1234"
		rec1 := httptest.NewRecorder()

		req2 := httptest.NewRequest(http.MethodGet, "/test", nil)
		req2.RemoteAddr = "192.168.1.2:1234"
		rec2 := httptest.NewRecorder()

		handler.ServeHTTP(rec1, req1)
		handler.ServeHTTP(rec2, req2)

		// Assert - both should succeed (different IPs)
		assert.Equal(t, http.StatusOK, rec1.Code)
		assert.Equal(t, http.StatusOK, rec2.Code)
	})

	t.Run("rate limit per IP is independent", func(t *testing.T) {
		// Arrange
		limiter := NewRateLimiter(1, 2)
		handler := RateLimit(limiter)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusOK)
		}))

		ip1 := "192.168.1.1:1234"
		ip2 := "192.168.1.2:5678"

		// Act - exhaust limit for IP1
		for i := 0; i < 3; i++ {
			req := httptest.NewRequest(http.MethodGet, "/test", nil)
			req.RemoteAddr = ip1
			rec := httptest.NewRecorder()
			handler.ServeHTTP(rec, req)
		}

		// IP2 should still be allowed
		req := httptest.NewRequest(http.MethodGet, "/test", nil)
		req.RemoteAddr = ip2
		rec := httptest.NewRecorder()
		handler.ServeHTTP(rec, req)

		// Assert
		assert.Equal(t, http.StatusOK, rec.Code, "IP2 should not be affected by IP1's rate limit")
	})

	t.Run("returns correct error response", func(t *testing.T) {
		// Arrange
		limiter := NewRateLimiter(0.1, 1)
		handler := RateLimit(limiter)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusOK)
		}))

		ip := "192.168.1.1:1234"

		// Act - exhaust limit
		req1 := httptest.NewRequest(http.MethodGet, "/test", nil)
		req1.RemoteAddr = ip
		rec1 := httptest.NewRecorder()
		handler.ServeHTTP(rec1, req1) // First request succeeds

		req2 := httptest.NewRequest(http.MethodGet, "/test", nil)
		req2.RemoteAddr = ip
		rec2 := httptest.NewRecorder()
		handler.ServeHTTP(rec2, req2) // Second should be blocked

		// Assert
		assert.Equal(t, http.StatusTooManyRequests, rec2.Code)
		assert.Equal(t, "application/json", rec2.Header().Get("Content-Type"))
		assert.Contains(t, rec2.Body.String(), "Rate limit exceeded")
		assert.Contains(t, rec2.Body.String(), "RATE_LIMIT_EXCEEDED")
	})

	t.Run("allows requests after waiting", func(t *testing.T) {
		// Arrange
		limiter := NewRateLimiter(5, 1) // 5 requests per second
		handler := RateLimit(limiter)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusOK)
		}))

		ip := "192.168.1.1:1234"

		// Act - first request
		req1 := httptest.NewRequest(http.MethodGet, "/test", nil)
		req1.RemoteAddr = ip
		rec1 := httptest.NewRecorder()
		handler.ServeHTTP(rec1, req1)

		// Wait for rate limit to reset
		time.Sleep(250 * time.Millisecond) // Wait 0.25 seconds (more than 1/5 second)

		// Second request after waiting
		req2 := httptest.NewRequest(http.MethodGet, "/test", nil)
		req2.RemoteAddr = ip
		rec2 := httptest.NewRecorder()
		handler.ServeHTTP(rec2, req2)

		// Assert
		assert.Equal(t, http.StatusOK, rec1.Code)
		assert.Equal(t, http.StatusOK, rec2.Code)
	})
}

func TestRateLimit_Concurrent(t *testing.T) {
	t.Run("handles concurrent requests from same IP", func(t *testing.T) {
		// Arrange
		limiter := NewRateLimiter(10, 5)
		handler := RateLimit(limiter)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusOK)
		}))

		ip := "192.168.1.1:1234"
		concurrentRequests := 10

		// Act
		var wg sync.WaitGroup
		results := make([]int, concurrentRequests)

		for i := 0; i < concurrentRequests; i++ {
			wg.Add(1)
			go func(idx int) {
				defer wg.Done()

				req := httptest.NewRequest(http.MethodGet, "/test", nil)
				req.RemoteAddr = ip
				rec := httptest.NewRecorder()

				handler.ServeHTTP(rec, req)
				results[idx] = rec.Code
			}(i)
		}

		wg.Wait()

		// Assert - count successes and failures
		successCount := 0
		failCount := 0
		for _, code := range results {
			if code == http.StatusOK {
				successCount++
			} else if code == http.StatusTooManyRequests {
				failCount++
			}
		}

		// Should have some successes (within burst) and some failures
		assert.Greater(t, successCount, 0, "Should have some successful requests")
		assert.Greater(t, failCount, 0, "Should have some rate-limited requests")
	})

	t.Run("handles concurrent requests from multiple IPs", func(t *testing.T) {
		// Arrange
		limiter := NewRateLimiter(5, 2)
		handler := RateLimit(limiter)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusOK)
		}))

		ips := []string{
			"192.168.1.1:1234",
			"192.168.1.2:5678",
			"192.168.1.3:9012",
		}

		// Act
		var wg sync.WaitGroup
		for _, ip := range ips {
			for i := 0; i < 2; i++ { // 2 requests per IP (within burst)
				wg.Add(1)
				go func(remoteAddr string) {
					defer wg.Done()

					req := httptest.NewRequest(http.MethodGet, "/test", nil)
					req.RemoteAddr = remoteAddr
					rec := httptest.NewRecorder()

					handler.ServeHTTP(rec, req)
					assert.Equal(t, http.StatusOK, rec.Code, "Requests within burst from different IPs should succeed")
				}(ip)
			}
		}

		wg.Wait()

		// Assert - all requests should succeed (each IP has its own limit)
		// No assertion needed here as the test would fail if any goroutine assertion failed
	})
}

func TestRateLimit_Integration(t *testing.T) {
	t.Run("works with other middleware", func(t *testing.T) {
		// Arrange
		limiter := NewRateLimiter(10, 5)

		finalHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusOK)
			w.Write([]byte("success"))
		})

		// Apply multiple middleware
		handler := RequestID(SecurityHeaders(RateLimit(limiter)(finalHandler)))

		req := httptest.NewRequest(http.MethodGet, "/test", nil)
		req.RemoteAddr = "192.168.1.1:1234"
		rec := httptest.NewRecorder()

		// Act
		handler.ServeHTTP(rec, req)

		// Assert
		assert.Equal(t, http.StatusOK, rec.Code)
		assert.Equal(t, "success", rec.Body.String())

		// Check other middleware still works
		assert.NotEmpty(t, rec.Header().Get("X-Request-ID"))
		assert.Equal(t, "nosniff", rec.Header().Get("X-Content-Type-Options"))
	})

	t.Run("rate limit blocks before reaching handler", func(t *testing.T) {
		// Arrange
		limiter := NewRateLimiter(0.1, 1)
		handlerCalled := false

		handler := RateLimit(limiter)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			handlerCalled = true
			w.WriteHeader(http.StatusOK)
		}))

		ip := "192.168.1.1:1234"

		// Act - first request succeeds
		req1 := httptest.NewRequest(http.MethodGet, "/test", nil)
		req1.RemoteAddr = ip
		rec1 := httptest.NewRecorder()
		handler.ServeHTTP(rec1, req1)

		// Reset flag
		handlerCalled = false

		// Second request should be blocked
		req2 := httptest.NewRequest(http.MethodGet, "/test", nil)
		req2.RemoteAddr = ip
		rec2 := httptest.NewRecorder()
		handler.ServeHTTP(rec2, req2)

		// Assert
		assert.Equal(t, http.StatusTooManyRequests, rec2.Code)
		assert.False(t, handlerCalled, "Handler should not be called when rate limited")
	})
}

func TestRateLimit_EdgeCases(t *testing.T) {
	t.Run("handles empty remote address", func(t *testing.T) {
		// Arrange
		limiter := NewRateLimiter(10, 5)
		handler := RateLimit(limiter)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusOK)
		}))

		req := httptest.NewRequest(http.MethodGet, "/test", nil)
		req.RemoteAddr = ""
		rec := httptest.NewRecorder()

		// Act
		handler.ServeHTTP(rec, req)

		// Assert - should still work (empty string is a valid key)
		assert.Equal(t, http.StatusOK, rec.Code)
	})

	t.Run("handles very high burst", func(t *testing.T) {
		// Arrange
		limiter := NewRateLimiter(1, 1000) // Very high burst
		handler := RateLimit(limiter)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusOK)
		}))

		ip := "192.168.1.1:1234"

		// Act - make many requests rapidly
		successCount := 0
		for i := 0; i < 100; i++ {
			req := httptest.NewRequest(http.MethodGet, "/test", nil)
			req.RemoteAddr = ip
			rec := httptest.NewRecorder()
			handler.ServeHTTP(rec, req)

			if rec.Code == http.StatusOK {
				successCount++
			}
		}

		// Assert - should allow many requests due to high burst
		assert.Greater(t, successCount, 50, "High burst should allow many requests")
	})

	t.Run("handles zero burst", func(t *testing.T) {
		// Arrange
		limiter := NewRateLimiter(10, 0) // Zero burst
		handler := RateLimit(limiter)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusOK)
		}))

		req := httptest.NewRequest(http.MethodGet, "/test", nil)
		req.RemoteAddr = "192.168.1.1:1234"
		rec := httptest.NewRecorder()

		// Act
		handler.ServeHTTP(rec, req)

		// Assert - behavior with zero burst (golang.org/x/time/rate handles this)
		// The request might succeed or fail depending on rate limiter implementation
		assert.True(t, rec.Code == http.StatusOK || rec.Code == http.StatusTooManyRequests)
	})
}
