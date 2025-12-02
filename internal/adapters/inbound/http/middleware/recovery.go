package middleware

import (
	"fmt"
	"log/slog"
	"net/http"
	"runtime/debug"
)

// Recovery middleware recovers from panics and returns a 500 error
func Recovery(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		defer func() {
			if err := recover(); err != nil {
				// Get request ID for logging
				requestID := GetRequestID(r.Context())
				
				// Log panic with stack trace
				slog.Error("panic recovered",
					slog.String("request_id", requestID),
					slog.String("method", r.Method),
					slog.String("path", r.URL.Path),
					slog.Any("error", err),
					slog.String("stack", string(debug.Stack())),
				)
				
				// Return 500 error
				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(http.StatusInternalServerError)
				fmt.Fprintf(w, `{"error":"Internal server error","code":"INTERNAL_SERVER_ERROR","request_id":"%s"}`, requestID)
			}
		}()
		
		next.ServeHTTP(w, r)
	})
}

