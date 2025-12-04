package middleware

import (
	"crypto/tls"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestEnforceHTTPS(t *testing.T) {
	t.Run("disabled", func(t *testing.T) {
		t.Run("allows HTTP when disabled", func(t *testing.T) {
			// Arrange
			cfg := HTTPSConfig{Enabled: false}
			handler := EnforceHTTPS(cfg)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(http.StatusOK)
			}))
			req := httptest.NewRequest(http.MethodGet, "http://example.com/test", nil)
			rec := httptest.NewRecorder()

			// Act
			handler.ServeHTTP(rec, req)

			// Assert
			assert.Equal(t, http.StatusOK, rec.Code)
		})
	})

	t.Run("enabled", func(t *testing.T) {
		t.Run("allows HTTPS requests with TLS", func(t *testing.T) {
			// Arrange
			cfg := HTTPSConfig{Enabled: true}
			handler := EnforceHTTPS(cfg)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(http.StatusOK)
			}))
			req := httptest.NewRequest(http.MethodGet, "https://example.com/test", nil)
			req.TLS = &tls.ConnectionState{}
			rec := httptest.NewRecorder()

			// Act
			handler.ServeHTTP(rec, req)

			// Assert
			assert.Equal(t, http.StatusOK, rec.Code)
		})

		t.Run("allows requests with X-Forwarded-Proto: https", func(t *testing.T) {
			// Arrange
			cfg := HTTPSConfig{Enabled: true}
			handler := EnforceHTTPS(cfg)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(http.StatusOK)
			}))
			req := httptest.NewRequest(http.MethodGet, "http://example.com/test", nil)
			req.Header.Set("X-Forwarded-Proto", "https")
			rec := httptest.NewRecorder()

			// Act
			handler.ServeHTTP(rec, req)

			// Assert
			assert.Equal(t, http.StatusOK, rec.Code)
		})

		t.Run("redirects HTTP to HTTPS", func(t *testing.T) {
			// Arrange
			cfg := HTTPSConfig{Enabled: true}
			handler := EnforceHTTPS(cfg)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(http.StatusOK)
			}))
			req := httptest.NewRequest(http.MethodGet, "http://example.com/test", nil)
			rec := httptest.NewRecorder()

			// Act
			handler.ServeHTTP(rec, req)

			// Assert
			assert.Equal(t, http.StatusMovedPermanently, rec.Code)
			assert.Equal(t, "https://example.com/test", rec.Header().Get("Location"))
		})

		t.Run("redirects HTTP to HTTPS with query string", func(t *testing.T) {
			// Arrange
			cfg := HTTPSConfig{Enabled: true}
			handler := EnforceHTTPS(cfg)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(http.StatusOK)
			}))
			req := httptest.NewRequest(http.MethodGet, "http://example.com/test?foo=bar&baz=qux", nil)
			rec := httptest.NewRecorder()

			// Act
			handler.ServeHTTP(rec, req)

			// Assert
			assert.Equal(t, http.StatusMovedPermanently, rec.Code)
			assert.Equal(t, "https://example.com/test?foo=bar&baz=qux", rec.Header().Get("Location"))
		})

		t.Run("redirects HTTP to HTTPS preserving path", func(t *testing.T) {
			// Arrange
			cfg := HTTPSConfig{Enabled: true}
			handler := EnforceHTTPS(cfg)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(http.StatusOK)
			}))
			req := httptest.NewRequest(http.MethodGet, "http://example.com/api/v1/users/123", nil)
			rec := httptest.NewRecorder()

			// Act
			handler.ServeHTTP(rec, req)

			// Assert
			assert.Equal(t, http.StatusMovedPermanently, rec.Code)
			assert.Equal(t, "https://example.com/api/v1/users/123", rec.Header().Get("Location"))
		})
	})

	t.Run("edge cases", func(t *testing.T) {
		t.Run("handles root path", func(t *testing.T) {
			// Arrange
			cfg := HTTPSConfig{Enabled: true}
			handler := EnforceHTTPS(cfg)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(http.StatusOK)
			}))
			req := httptest.NewRequest(http.MethodGet, "http://example.com/", nil)
			rec := httptest.NewRecorder()

			// Act
			handler.ServeHTTP(rec, req)

			// Assert
			assert.Equal(t, http.StatusMovedPermanently, rec.Code)
			assert.Equal(t, "https://example.com/", rec.Header().Get("Location"))
		})

		t.Run("handles missing host", func(t *testing.T) {
			// Arrange
			cfg := HTTPSConfig{Enabled: true}
			handler := EnforceHTTPS(cfg)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(http.StatusOK)
			}))
			req := httptest.NewRequest(http.MethodGet, "/test", nil)
			req.Host = "example.com"
			rec := httptest.NewRecorder()

			// Act
			handler.ServeHTTP(rec, req)

			// Assert
			assert.Equal(t, http.StatusMovedPermanently, rec.Code)
			assert.Equal(t, "https://example.com/test", rec.Header().Get("Location"))
		})
	})
}
