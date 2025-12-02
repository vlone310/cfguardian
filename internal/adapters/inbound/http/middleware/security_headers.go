package middleware

import (
	"net/http"
)

// SecurityHeaders adds security-related HTTP headers
func SecurityHeaders(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Prevent MIME type sniffing
		w.Header().Set("X-Content-Type-Options", "nosniff")
		
		// Enable XSS protection
		w.Header().Set("X-XSS-Protection", "1; mode=block")
		
		// Prevent clickjacking
		w.Header().Set("X-Frame-Options", "DENY")
		
		// Strict Transport Security (HSTS) - enforce HTTPS
		// Only set if running over HTTPS
		if r.TLS != nil {
			w.Header().Set("Strict-Transport-Security", "max-age=31536000; includeSubDomains")
		}
		
		// Content Security Policy - prevent XSS and data injection
		w.Header().Set("Content-Security-Policy", "default-src 'self'")
		
		// Referrer Policy - control referrer information
		w.Header().Set("Referrer-Policy", "strict-origin-when-cross-origin")
		
		// Permissions Policy - restrict browser features
		w.Header().Set("Permissions-Policy", "geolocation=(), microphone=(), camera=()")
		
		next.ServeHTTP(w, r)
	})
}

