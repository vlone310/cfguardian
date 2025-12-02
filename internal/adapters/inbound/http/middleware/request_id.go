package middleware

import (
	"context"
	"net/http"

	"github.com/google/uuid"
)

// contextKey is a custom type for context keys to avoid collisions
type contextKey string

const (
	// RequestIDKey is the context key for request ID
	RequestIDKey contextKey = "request_id"
)

// RequestID middleware generates a unique ID for each request
func RequestID(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Check if request ID already exists in header
		requestID := r.Header.Get("X-Request-ID")
		
		// Generate new ID if not present
		if requestID == "" {
			requestID = uuid.New().String()
		}
		
		// Add to context
		ctx := context.WithValue(r.Context(), RequestIDKey, requestID)
		
		// Add to response header
		w.Header().Set("X-Request-ID", requestID)
		
		// Call next handler
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

// GetRequestID retrieves the request ID from context
func GetRequestID(ctx context.Context) string {
	if requestID, ok := ctx.Value(RequestIDKey).(string); ok {
		return requestID
	}
	return ""
}

