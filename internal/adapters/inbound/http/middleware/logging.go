package middleware

import (
	"fmt"
	"log/slog"
	"net/http"
	"time"
)

// responseWriter wraps http.ResponseWriter to capture status code
type responseWriter struct {
	http.ResponseWriter
	status int
	size   int
}

func (rw *responseWriter) WriteHeader(status int) {
	rw.status = status
	rw.ResponseWriter.WriteHeader(status)
}

func (rw *responseWriter) Write(b []byte) (int, error) {
	size, err := rw.ResponseWriter.Write(b)
	rw.size += size
	return size, err
}

// Logging middleware logs HTTP requests with structured logging
func Logging(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		
		// Wrap response writer to capture status
		wrapped := &responseWriter{
			ResponseWriter: w,
			status:         http.StatusOK,
		}
		
		// Get request ID from context
		requestID := GetRequestID(r.Context())
		
		// Log request start
		slog.Info("http request started",
			slog.String("request_id", requestID),
			slog.String("method", r.Method),
			slog.String("path", r.URL.Path),
			slog.String("remote_addr", r.RemoteAddr),
			slog.String("user_agent", r.UserAgent()),
		)
		
		// Call next handler
		next.ServeHTTP(wrapped, r)
		
		// Log request completion
		duration := time.Since(start)
		
		slog.Info("http request completed",
			slog.String("request_id", requestID),
			slog.String("method", r.Method),
			slog.String("path", r.URL.Path),
			slog.Int("status", wrapped.status),
			slog.Int("size", wrapped.size),
			slog.Duration("duration", duration),
			slog.String("duration_ms", fmt.Sprintf("%.2f", duration.Seconds()*1000)),
		)
	})
}

