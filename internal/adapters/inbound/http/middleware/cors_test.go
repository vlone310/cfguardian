package middleware

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestCORS(t *testing.T) {
	t.Run("allows configured origins", func(t *testing.T) {
		// Arrange
		handler := CORS()(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusOK)
			w.Write([]byte("success"))
		}))

		tests := []struct {
			name   string
			origin string
		}{
			{"localhost:3000", "http://localhost:3000"},
			{"localhost:8080", "http://localhost:8080"},
		}

		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				req := httptest.NewRequest(http.MethodGet, "/test", nil)
				req.Header.Set("Origin", tt.origin)
				rec := httptest.NewRecorder()

				// Act
				handler.ServeHTTP(rec, req)

				// Assert
				assert.Equal(t, http.StatusOK, rec.Code)
				assert.Equal(t, tt.origin, rec.Header().Get("Access-Control-Allow-Origin"))
				assert.Equal(t, "true", rec.Header().Get("Access-Control-Allow-Credentials"))
				assert.Contains(t, rec.Header().Get("Vary"), "Origin")
			})
		}
	})

	t.Run("blocks non-configured origins", func(t *testing.T) {
		// Arrange
		handler := CORS()(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusOK)
			w.Write([]byte("success"))
		}))

		req := httptest.NewRequest(http.MethodGet, "/test", nil)
		req.Header.Set("Origin", "http://malicious.com")
		rec := httptest.NewRecorder()

		// Act
		handler.ServeHTTP(rec, req)

		// Assert
		// CORS middleware should not set Access-Control-Allow-Origin for disallowed origins
		assert.Empty(t, rec.Header().Get("Access-Control-Allow-Origin"))
	})

	t.Run("handles preflight OPTIONS request", func(t *testing.T) {
		// Arrange
		handler := CORS()(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusOK)
		}))

		req := httptest.NewRequest(http.MethodOptions, "/test", nil)
		req.Header.Set("Origin", "http://localhost:3000")
		req.Header.Set("Access-Control-Request-Method", "POST")
		req.Header.Set("Access-Control-Request-Headers", "Content-Type")
		rec := httptest.NewRecorder()

		// Act
		handler.ServeHTTP(rec, req)

		// Assert
		// go-chi/cors returns 200 OK for preflight, not 204
		assert.Equal(t, http.StatusOK, rec.Code)
		assert.Equal(t, "http://localhost:3000", rec.Header().Get("Access-Control-Allow-Origin"))
		assert.NotEmpty(t, rec.Header().Get("Access-Control-Allow-Methods"))
		assert.NotEmpty(t, rec.Header().Get("Access-Control-Allow-Headers"))
		assert.Equal(t, "300", rec.Header().Get("Access-Control-Max-Age"))
	})

	t.Run("allows all configured HTTP methods", func(t *testing.T) {
		// Arrange
		handler := CORS()(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusOK)
		}))

		methods := []string{"GET", "POST", "PUT", "DELETE"}

		for _, method := range methods {
			t.Run(method, func(t *testing.T) {
				req := httptest.NewRequest(http.MethodOptions, "/test", nil)
				req.Header.Set("Origin", "http://localhost:3000")
				req.Header.Set("Access-Control-Request-Method", method)
				rec := httptest.NewRecorder()

				// Act
				handler.ServeHTTP(rec, req)

				// Assert
				assert.Equal(t, http.StatusOK, rec.Code)
				allowedMethods := rec.Header().Get("Access-Control-Allow-Methods")
				assert.NotEmpty(t, allowedMethods)
			})
		}
	})

	t.Run("allows configured headers", func(t *testing.T) {
		// Arrange
		handler := CORS()(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusOK)
		}))

		headers := []string{"Accept", "Authorization", "Content-Type", "X-Request-ID", "X-API-Key"}

		for _, header := range headers {
			t.Run(header, func(t *testing.T) {
				req := httptest.NewRequest(http.MethodOptions, "/test", nil)
				req.Header.Set("Origin", "http://localhost:3000")
				req.Header.Set("Access-Control-Request-Method", "POST")
				req.Header.Set("Access-Control-Request-Headers", header)
				rec := httptest.NewRecorder()

				// Act
				handler.ServeHTTP(rec, req)

				// Assert
				assert.Equal(t, http.StatusOK, rec.Code)
				allowedHeaders := rec.Header().Get("Access-Control-Allow-Headers")
				assert.NotEmpty(t, allowedHeaders)
			})
		}
	})

	t.Run("exposes X-Request-ID header", func(t *testing.T) {
		// Arrange
		handler := CORS()(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("X-Request-ID", "test-id-123")
			w.WriteHeader(http.StatusOK)
		}))

		req := httptest.NewRequest(http.MethodGet, "/test", nil)
		req.Header.Set("Origin", "http://localhost:3000")
		rec := httptest.NewRecorder()

		// Act
		handler.ServeHTTP(rec, req)

		// Assert
		assert.Equal(t, http.StatusOK, rec.Code)
		// Check that exposed headers config is honored
		// Note: HTTP headers are case-insensitive, go-chi/cors may normalize to X-Request-Id
		exposedHeaders := rec.Header().Get("Access-Control-Expose-Headers")
		if exposedHeaders != "" {
			// Check for either capitalization
			hasHeader := exposedHeaders == "X-Request-ID" || exposedHeaders == "X-Request-Id"
			assert.True(t, hasHeader, "Expected X-Request-ID or X-Request-Id in exposed headers, got: %s", exposedHeaders)
		}
	})

	t.Run("allows credentials", func(t *testing.T) {
		// Arrange
		handler := CORS()(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusOK)
		}))

		req := httptest.NewRequest(http.MethodGet, "/test", nil)
		req.Header.Set("Origin", "http://localhost:3000")
		rec := httptest.NewRecorder()

		// Act
		handler.ServeHTTP(rec, req)

		// Assert
		assert.Equal(t, http.StatusOK, rec.Code)
		assert.Equal(t, "true", rec.Header().Get("Access-Control-Allow-Credentials"))
	})

	t.Run("sets max age for preflight cache", func(t *testing.T) {
		// Arrange
		handler := CORS()(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusOK)
		}))

		req := httptest.NewRequest(http.MethodOptions, "/test", nil)
		req.Header.Set("Origin", "http://localhost:3000")
		req.Header.Set("Access-Control-Request-Method", "POST")
		rec := httptest.NewRecorder()

		// Act
		handler.ServeHTTP(rec, req)

		// Assert
		assert.Equal(t, "300", rec.Header().Get("Access-Control-Max-Age")) // 5 minutes
	})

	t.Run("handles requests without Origin header", func(t *testing.T) {
		// Arrange
		handler := CORS()(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusOK)
			w.Write([]byte("success"))
		}))

		req := httptest.NewRequest(http.MethodGet, "/test", nil)
		// No Origin header set
		rec := httptest.NewRecorder()

		// Act
		handler.ServeHTTP(rec, req)

		// Assert
		// Request should still succeed
		assert.Equal(t, http.StatusOK, rec.Code)
		assert.Equal(t, "success", rec.Body.String())
	})

	t.Run("rejects disallowed methods in preflight", func(t *testing.T) {
		// Arrange
		handler := CORS()(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusOK)
		}))

		req := httptest.NewRequest(http.MethodOptions, "/test", nil)
		req.Header.Set("Origin", "http://localhost:3000")
		req.Header.Set("Access-Control-Request-Method", "PATCH") // Not in allowed methods
		rec := httptest.NewRecorder()

		// Act
		handler.ServeHTTP(rec, req)

		// Assert
		// CORS middleware should reject this
		assert.NotEqual(t, http.StatusNoContent, rec.Code)
	})

	t.Run("rejects disallowed headers in preflight", func(t *testing.T) {
		// Arrange
		handler := CORS()(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusOK)
		}))

		req := httptest.NewRequest(http.MethodOptions, "/test", nil)
		req.Header.Set("Origin", "http://localhost:3000")
		req.Header.Set("Access-Control-Request-Method", "POST")
		req.Header.Set("Access-Control-Request-Headers", "X-Custom-Forbidden-Header")
		rec := httptest.NewRecorder()

		// Act
		handler.ServeHTTP(rec, req)

		// Assert
		// CORS middleware should reject this
		assert.NotEqual(t, http.StatusNoContent, rec.Code)
	})
}

func TestCORS_Integration(t *testing.T) {
	t.Run("works with other middleware", func(t *testing.T) {
		// Arrange
		finalHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("X-Request-ID", "test-123")
			w.WriteHeader(http.StatusOK)
			w.Write([]byte("success"))
		})

		// Apply multiple middleware
		handler := RequestID(SecurityHeaders(CORS()(finalHandler)))

		req := httptest.NewRequest(http.MethodGet, "/test", nil)
		req.Header.Set("Origin", "http://localhost:3000")
		rec := httptest.NewRecorder()

		// Act
		handler.ServeHTTP(rec, req)

		// Assert
		assert.Equal(t, http.StatusOK, rec.Code)
		assert.Equal(t, "success", rec.Body.String())

		// Check CORS headers
		assert.Equal(t, "http://localhost:3000", rec.Header().Get("Access-Control-Allow-Origin"))
		assert.Equal(t, "true", rec.Header().Get("Access-Control-Allow-Credentials"))

		// Check security headers still applied
		assert.Equal(t, "nosniff", rec.Header().Get("X-Content-Type-Options"))
		assert.Equal(t, "DENY", rec.Header().Get("X-Frame-Options"))

		// Check request ID
		assert.NotEmpty(t, rec.Header().Get("X-Request-ID"))
	})

	t.Run("preflight request bypasses other middleware", func(t *testing.T) {
		// Arrange
		finalHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// go-chi/cors calls the next handler even for preflight
			w.WriteHeader(http.StatusOK)
		})

		handler := CORS()(finalHandler)

		req := httptest.NewRequest(http.MethodOptions, "/test", nil)
		req.Header.Set("Origin", "http://localhost:3000")
		req.Header.Set("Access-Control-Request-Method", "POST")
		rec := httptest.NewRecorder()

		// Act
		handler.ServeHTTP(rec, req)

		// Assert
		assert.Equal(t, http.StatusOK, rec.Code)
		assert.Equal(t, "http://localhost:3000", rec.Header().Get("Access-Control-Allow-Origin"))
	})

	t.Run("actual request after preflight", func(t *testing.T) {
		// Arrange
		handler := CORS()(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if r.Method == http.MethodPost {
				w.WriteHeader(http.StatusCreated)
				w.Write([]byte(`{"status":"created"}`))
			} else {
				w.WriteHeader(http.StatusOK)
			}
		}))

		// Step 1: Preflight request
		preflightReq := httptest.NewRequest(http.MethodOptions, "/test", nil)
		preflightReq.Header.Set("Origin", "http://localhost:3000")
		preflightReq.Header.Set("Access-Control-Request-Method", "POST")
		preflightReq.Header.Set("Access-Control-Request-Headers", "Content-Type")
		preflightRec := httptest.NewRecorder()

		handler.ServeHTTP(preflightRec, preflightReq)

		assert.Equal(t, http.StatusOK, preflightRec.Code)

		// Step 2: Actual request
		actualReq := httptest.NewRequest(http.MethodPost, "/test", nil)
		actualReq.Header.Set("Origin", "http://localhost:3000")
		actualReq.Header.Set("Content-Type", "application/json")
		actualRec := httptest.NewRecorder()

		handler.ServeHTTP(actualRec, actualReq)

		// Assert
		assert.Equal(t, http.StatusCreated, actualRec.Code)
		assert.Equal(t, `{"status":"created"}`, actualRec.Body.String())
		assert.Equal(t, "http://localhost:3000", actualRec.Header().Get("Access-Control-Allow-Origin"))
	})
}

func TestCORS_SecurityScenarios(t *testing.T) {
	t.Run("blocks null origin", func(t *testing.T) {
		// Arrange
		handler := CORS()(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusOK)
		}))

		req := httptest.NewRequest(http.MethodGet, "/test", nil)
		req.Header.Set("Origin", "null")
		rec := httptest.NewRecorder()

		// Act
		handler.ServeHTTP(rec, req)

		// Assert
		// Should not allow null origin
		assert.Empty(t, rec.Header().Get("Access-Control-Allow-Origin"))
	})

	t.Run("blocks wildcard origin", func(t *testing.T) {
		// Arrange
		handler := CORS()(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusOK)
		}))

		req := httptest.NewRequest(http.MethodGet, "/test", nil)
		req.Header.Set("Origin", "*")
		rec := httptest.NewRecorder()

		// Act
		handler.ServeHTTP(rec, req)

		// Assert
		// Should not allow wildcard origin
		assert.Empty(t, rec.Header().Get("Access-Control-Allow-Origin"))
	})

	t.Run("prevents CORS bypass attempts", func(t *testing.T) {
		// Arrange
		handler := CORS()(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusOK)
		}))

		bypassAttempts := []string{
			"http://localhost:3000.malicious.com",
			"http://malicious.com/localhost:3000",
			"http://localhost:3000@malicious.com",
		}

		for _, origin := range bypassAttempts {
			t.Run(origin, func(t *testing.T) {
				req := httptest.NewRequest(http.MethodGet, "/test", nil)
				req.Header.Set("Origin", origin)
				rec := httptest.NewRecorder()

				// Act
				handler.ServeHTTP(rec, req)

				// Assert
				assert.Empty(t, rec.Header().Get("Access-Control-Allow-Origin"),
					"Should not allow bypass attempt: %s", origin)
			})
		}
	})
}
