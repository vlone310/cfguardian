package middleware

import (
	"net/http"
)

// HTTPSConfig holds HTTPS enforcement configuration
type HTTPSConfig struct {
	Enabled bool
}

// EnforceHTTPS redirects HTTP requests to HTTPS
func EnforceHTTPS(cfg HTTPSConfig) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// Skip enforcement if disabled
			if !cfg.Enabled {
				next.ServeHTTP(w, r)
				return
			}
			
			// Check if request is over HTTPS
			if r.TLS == nil && r.Header.Get("X-Forwarded-Proto") != "https" {
				// Redirect to HTTPS
				target := "https://" + r.Host + r.URL.Path
				if r.URL.RawQuery != "" {
					target += "?" + r.URL.RawQuery
				}
				
				http.Redirect(w, r, target, http.StatusMovedPermanently)
				return
			}
			
			next.ServeHTTP(w, r)
		})
	}
}

