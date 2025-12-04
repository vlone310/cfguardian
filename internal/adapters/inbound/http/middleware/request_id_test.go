package middleware

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestRequestID(t *testing.T) {
	t.Run("generates new request ID when not provided", func(t *testing.T) {
		// Arrange
		finalHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// Get request ID from context
			requestID := GetRequestID(r.Context())
			assert.NotEmpty(t, requestID, "Request ID should be set in context")

			// Validate it's a valid UUID
			_, err := uuid.Parse(requestID)
			assert.NoError(t, err, "Request ID should be a valid UUID")

			w.WriteHeader(http.StatusOK)
		})

		handler := RequestID(finalHandler)
		req := httptest.NewRequest(http.MethodGet, "/test", nil)
		rec := httptest.NewRecorder()

		// Act
		handler.ServeHTTP(rec, req)

		// Assert
		assert.Equal(t, http.StatusOK, rec.Code)

		// Check response header
		responseRequestID := rec.Header().Get("X-Request-ID")
		assert.NotEmpty(t, responseRequestID, "Response should have X-Request-ID header")

		// Validate it's a valid UUID
		_, err := uuid.Parse(responseRequestID)
		assert.NoError(t, err, "Response request ID should be a valid UUID")
	})

	t.Run("uses existing request ID from header", func(t *testing.T) {
		// Arrange
		existingID := "test-request-id-12345"

		finalHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// Get request ID from context
			requestID := GetRequestID(r.Context())
			assert.Equal(t, existingID, requestID, "Should use existing request ID")

			w.WriteHeader(http.StatusOK)
		})

		handler := RequestID(finalHandler)
		req := httptest.NewRequest(http.MethodGet, "/test", nil)
		req.Header.Set("X-Request-ID", existingID)
		rec := httptest.NewRecorder()

		// Act
		handler.ServeHTTP(rec, req)

		// Assert
		assert.Equal(t, http.StatusOK, rec.Code)

		// Check response header has same ID
		responseRequestID := rec.Header().Get("X-Request-ID")
		assert.Equal(t, existingID, responseRequestID)
	})

	t.Run("generates unique IDs for different requests", func(t *testing.T) {
		// Arrange
		requestIDs := make(map[string]bool)

		finalHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			requestID := GetRequestID(r.Context())
			requestIDs[requestID] = true
			w.WriteHeader(http.StatusOK)
		})

		handler := RequestID(finalHandler)

		// Act - make multiple requests
		for i := 0; i < 10; i++ {
			req := httptest.NewRequest(http.MethodGet, "/test", nil)
			rec := httptest.NewRecorder()
			handler.ServeHTTP(rec, req)
		}

		// Assert - all IDs should be unique
		assert.Equal(t, 10, len(requestIDs), "All request IDs should be unique")
	})

	t.Run("adds request ID to context", func(t *testing.T) {
		// Arrange
		var capturedRequestID string

		finalHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			capturedRequestID = GetRequestID(r.Context())
			w.WriteHeader(http.StatusOK)
		})

		handler := RequestID(finalHandler)
		req := httptest.NewRequest(http.MethodGet, "/test", nil)
		rec := httptest.NewRecorder()

		// Act
		handler.ServeHTTP(rec, req)

		// Assert
		assert.NotEmpty(t, capturedRequestID)
		assert.Equal(t, rec.Header().Get("X-Request-ID"), capturedRequestID)
	})

	t.Run("preserves UUID format", func(t *testing.T) {
		// Arrange
		finalHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusOK)
		})

		handler := RequestID(finalHandler)
		req := httptest.NewRequest(http.MethodGet, "/test", nil)
		rec := httptest.NewRecorder()

		// Act
		handler.ServeHTTP(rec, req)

		// Assert
		requestID := rec.Header().Get("X-Request-ID")
		parsedUUID, err := uuid.Parse(requestID)
		require.NoError(t, err)
		assert.NotEqual(t, uuid.Nil, parsedUUID)
	})
}

func TestGetRequestID(t *testing.T) {
	t.Run("returns request ID from context", func(t *testing.T) {
		// Arrange
		expectedID := "test-id-123"
		ctx := context.WithValue(context.Background(), RequestIDKey, expectedID)

		// Act
		requestID := GetRequestID(ctx)

		// Assert
		assert.Equal(t, expectedID, requestID)
	})

	t.Run("returns empty string when no request ID in context", func(t *testing.T) {
		// Arrange
		ctx := context.Background()

		// Act
		requestID := GetRequestID(ctx)

		// Assert
		assert.Empty(t, requestID)
	})

	t.Run("returns empty string for wrong type in context", func(t *testing.T) {
		// Arrange
		ctx := context.WithValue(context.Background(), RequestIDKey, 12345) // wrong type

		// Act
		requestID := GetRequestID(ctx)

		// Assert
		assert.Empty(t, requestID)
	})
}

func TestRequestID_Integration(t *testing.T) {
	t.Run("request ID flows through middleware chain", func(t *testing.T) {
		// Arrange
		var requestIDInHandler string

		finalHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			requestIDInHandler = GetRequestID(r.Context())
			w.WriteHeader(http.StatusOK)
		})

		// Stack middleware
		handler := RequestID(finalHandler)

		req := httptest.NewRequest(http.MethodGet, "/test", nil)
		rec := httptest.NewRecorder()

		// Act
		handler.ServeHTTP(rec, req)

		// Assert
		headerRequestID := rec.Header().Get("X-Request-ID")
		assert.NotEmpty(t, headerRequestID)
		assert.NotEmpty(t, requestIDInHandler)
		assert.Equal(t, headerRequestID, requestIDInHandler,
			"Request ID in header should match ID in context")
	})

	t.Run("multiple requests have different IDs", func(t *testing.T) {
		// Arrange
		seenIDs := make(map[string]bool)

		finalHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusOK)
		})

		handler := RequestID(finalHandler)

		// Act - simulate multiple concurrent requests
		for i := 0; i < 100; i++ {
			req := httptest.NewRequest(http.MethodPost, "/api/test", nil)
			rec := httptest.NewRecorder()

			handler.ServeHTTP(rec, req)

			requestID := rec.Header().Get("X-Request-ID")
			require.NotEmpty(t, requestID)

			// Check for uniqueness
			assert.False(t, seenIDs[requestID], "Request ID should be unique")
			seenIDs[requestID] = true
		}

		// Assert - all 100 requests should have unique IDs
		assert.Equal(t, 100, len(seenIDs))
	})
}
