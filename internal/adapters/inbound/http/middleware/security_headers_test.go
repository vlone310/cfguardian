package middleware

import (
	"crypto/tls"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestSecurityHeaders(t *testing.T) {
	t.Run("sets all security headers for HTTP request", func(t *testing.T) {
		// Arrange
		finalHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusOK)
		})

		handler := SecurityHeaders(finalHandler)
		req := httptest.NewRequest(http.MethodGet, "/test", nil)
		rec := httptest.NewRecorder()

		// Act
		handler.ServeHTTP(rec, req)

		// Assert
		assert.Equal(t, http.StatusOK, rec.Code)

		// Check all security headers
		assert.Equal(t, "nosniff", rec.Header().Get("X-Content-Type-Options"),
			"Should set X-Content-Type-Options")
		assert.Equal(t, "1; mode=block", rec.Header().Get("X-XSS-Protection"),
			"Should set X-XSS-Protection")
		assert.Equal(t, "DENY", rec.Header().Get("X-Frame-Options"),
			"Should set X-Frame-Options")
		assert.Equal(t, "default-src 'self'", rec.Header().Get("Content-Security-Policy"),
			"Should set Content-Security-Policy")
		assert.Equal(t, "strict-origin-when-cross-origin", rec.Header().Get("Referrer-Policy"),
			"Should set Referrer-Policy")
		assert.Equal(t, "geolocation=(), microphone=(), camera=()", rec.Header().Get("Permissions-Policy"),
			"Should set Permissions-Policy")
	})

	t.Run("sets HSTS header for HTTPS request", func(t *testing.T) {
		// Arrange
		finalHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusOK)
		})

		handler := SecurityHeaders(finalHandler)
		req := httptest.NewRequest(http.MethodGet, "/test", nil)

		// Simulate HTTPS by setting TLS
		req.TLS = &tls.ConnectionState{
			Version: tls.VersionTLS13,
		}

		rec := httptest.NewRecorder()

		// Act
		handler.ServeHTTP(rec, req)

		// Assert
		assert.Equal(t, "max-age=31536000; includeSubDomains",
			rec.Header().Get("Strict-Transport-Security"),
			"Should set HSTS for HTTPS requests")
	})

	t.Run("does not set HSTS header for HTTP request", func(t *testing.T) {
		// Arrange
		finalHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusOK)
		})

		handler := SecurityHeaders(finalHandler)
		req := httptest.NewRequest(http.MethodGet, "/test", nil)
		// No TLS set - HTTP request
		rec := httptest.NewRecorder()

		// Act
		handler.ServeHTTP(rec, req)

		// Assert
		assert.Empty(t, rec.Header().Get("Strict-Transport-Security"),
			"Should not set HSTS for HTTP requests")
	})

	t.Run("security headers are set for different HTTP methods", func(t *testing.T) {
		methods := []string{
			http.MethodGet,
			http.MethodPost,
			http.MethodPut,
			http.MethodPatch,
			http.MethodDelete,
		}

		for _, method := range methods {
			t.Run(method, func(t *testing.T) {
				// Arrange
				finalHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
					w.WriteHeader(http.StatusOK)
				})

				handler := SecurityHeaders(finalHandler)
				req := httptest.NewRequest(method, "/test", nil)
				rec := httptest.NewRecorder()

				// Act
				handler.ServeHTTP(rec, req)

				// Assert - all security headers should be set
				assert.NotEmpty(t, rec.Header().Get("X-Content-Type-Options"))
				assert.NotEmpty(t, rec.Header().Get("X-XSS-Protection"))
				assert.NotEmpty(t, rec.Header().Get("X-Frame-Options"))
				assert.NotEmpty(t, rec.Header().Get("Content-Security-Policy"))
				assert.NotEmpty(t, rec.Header().Get("Referrer-Policy"))
				assert.NotEmpty(t, rec.Header().Get("Permissions-Policy"))
			})
		}
	})

	t.Run("security headers are set for different paths", func(t *testing.T) {
		paths := []string{
			"/",
			"/api/v1/users",
			"/api/v1/configs",
			"/health",
			"/metrics",
		}

		for _, path := range paths {
			t.Run(path, func(t *testing.T) {
				// Arrange
				finalHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
					w.WriteHeader(http.StatusOK)
				})

				handler := SecurityHeaders(finalHandler)
				req := httptest.NewRequest(http.MethodGet, path, nil)
				rec := httptest.NewRecorder()

				// Act
				handler.ServeHTTP(rec, req)

				// Assert
				assert.Equal(t, "nosniff", rec.Header().Get("X-Content-Type-Options"))
			})
		}
	})

	t.Run("does not interfere with handler execution", func(t *testing.T) {
		// Arrange
		handlerCalled := false

		finalHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			handlerCalled = true
			w.WriteHeader(http.StatusCreated)
			w.Write([]byte(`{"status":"ok"}`))
		})

		handler := SecurityHeaders(finalHandler)
		req := httptest.NewRequest(http.MethodPost, "/test", nil)
		rec := httptest.NewRecorder()

		// Act
		handler.ServeHTTP(rec, req)

		// Assert
		assert.True(t, handlerCalled, "Final handler should be called")
		assert.Equal(t, http.StatusCreated, rec.Code)
		assert.Equal(t, `{"status":"ok"}`, rec.Body.String())
	})
}

func TestSecurityHeaders_XSSProtection(t *testing.T) {
	t.Run("prevents XSS attacks with mode=block", func(t *testing.T) {
		// Arrange
		finalHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusOK)
		})

		handler := SecurityHeaders(finalHandler)
		req := httptest.NewRequest(http.MethodGet, "/test", nil)
		rec := httptest.NewRecorder()

		// Act
		handler.ServeHTTP(rec, req)

		// Assert
		xssHeader := rec.Header().Get("X-XSS-Protection")
		assert.Contains(t, xssHeader, "1")
		assert.Contains(t, xssHeader, "mode=block")
	})
}

func TestSecurityHeaders_Clickjacking(t *testing.T) {
	t.Run("prevents clickjacking with DENY", func(t *testing.T) {
		// Arrange
		finalHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusOK)
		})

		handler := SecurityHeaders(finalHandler)
		req := httptest.NewRequest(http.MethodGet, "/test", nil)
		rec := httptest.NewRecorder()

		// Act
		handler.ServeHTTP(rec, req)

		// Assert
		assert.Equal(t, "DENY", rec.Header().Get("X-Frame-Options"),
			"Should prevent all framing")
	})
}

func TestSecurityHeaders_CSP(t *testing.T) {
	t.Run("sets restrictive Content Security Policy", func(t *testing.T) {
		// Arrange
		finalHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusOK)
		})

		handler := SecurityHeaders(finalHandler)
		req := httptest.NewRequest(http.MethodGet, "/test", nil)
		rec := httptest.NewRecorder()

		// Act
		handler.ServeHTTP(rec, req)

		// Assert
		csp := rec.Header().Get("Content-Security-Policy")
		assert.Equal(t, "default-src 'self'", csp,
			"Should only allow resources from same origin")
	})
}

func TestSecurityHeaders_HSTS(t *testing.T) {
	t.Run("sets HSTS with includeSubDomains for HTTPS", func(t *testing.T) {
		// Arrange
		finalHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusOK)
		})

		handler := SecurityHeaders(finalHandler)
		req := httptest.NewRequest(http.MethodGet, "/test", nil)
		req.TLS = &tls.ConnectionState{Version: tls.VersionTLS13}
		rec := httptest.NewRecorder()

		// Act
		handler.ServeHTTP(rec, req)

		// Assert
		hsts := rec.Header().Get("Strict-Transport-Security")
		assert.Contains(t, hsts, "max-age=31536000")
		assert.Contains(t, hsts, "includeSubDomains")
	})

	t.Run("HSTS max-age is one year", func(t *testing.T) {
		// Arrange
		finalHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusOK)
		})

		handler := SecurityHeaders(finalHandler)
		req := httptest.NewRequest(http.MethodGet, "/test", nil)
		req.TLS = &tls.ConnectionState{Version: tls.VersionTLS12}
		rec := httptest.NewRecorder()

		// Act
		handler.ServeHTTP(rec, req)

		// Assert
		hsts := rec.Header().Get("Strict-Transport-Security")
		// 31536000 seconds = 365 days
		assert.Contains(t, hsts, "max-age=31536000")
	})
}

func TestSecurityHeaders_PermissionsPolicy(t *testing.T) {
	t.Run("restricts dangerous browser features", func(t *testing.T) {
		// Arrange
		finalHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusOK)
		})

		handler := SecurityHeaders(finalHandler)
		req := httptest.NewRequest(http.MethodGet, "/test", nil)
		rec := httptest.NewRecorder()

		// Act
		handler.ServeHTTP(rec, req)

		// Assert
		permsPolicy := rec.Header().Get("Permissions-Policy")
		assert.Contains(t, permsPolicy, "geolocation=()")
		assert.Contains(t, permsPolicy, "microphone=()")
		assert.Contains(t, permsPolicy, "camera=()")
	})
}
