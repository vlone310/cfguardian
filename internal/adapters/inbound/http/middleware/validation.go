package middleware

import (
	"fmt"
	"net/http"

	"github.com/vlone310/cfguardian/internal/adapters/inbound/http/common"
)

// MaxRequestSize sets the maximum allowed request body size (default 10MB)
const MaxRequestSize = 10 * 1024 * 1024 // 10MB

// RequestSizeLimit middleware limits the size of incoming request bodies
func RequestSizeLimit(maxBytes int64) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// Limit the request body size
			r.Body = http.MaxBytesReader(w, r.Body, maxBytes)
			next.ServeHTTP(w, r)
		})
	}
}

// ContentTypeValidation middleware validates Content-Type header for POST/PUT requests
func ContentTypeValidation(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Only check for requests with body
		if r.Method == http.MethodPost || r.Method == http.MethodPut || r.Method == http.MethodPatch {
			contentType := r.Header.Get("Content-Type")
			
			// Allow empty content type for requests with no body
			if r.ContentLength > 0 && contentType == "" {
				common.BadRequest(w, "Content-Type header is required")
				return
			}
			
			// Validate Content-Type is application/json for our API
			if r.ContentLength > 0 && contentType != "application/json" {
				common.BadRequest(w, fmt.Sprintf("Invalid Content-Type: %s (expected application/json)", contentType))
				return
			}
		}
		
		next.ServeHTTP(w, r)
	})
}

// MethodValidation middleware validates HTTP methods
func MethodValidation(allowedMethods []string) func(http.Handler) http.Handler {
	methodMap := make(map[string]bool)
	for _, method := range allowedMethods {
		methodMap[method] = true
	}
	
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if !methodMap[r.Method] {
				w.Header().Set("Allow", joinMethods(allowedMethods))
				common.BadRequest(w, fmt.Sprintf("Method %s not allowed", r.Method))
				return
			}
			next.ServeHTTP(w, r)
		})
	}
}

// joinMethods joins method names with comma
func joinMethods(methods []string) string {
	if len(methods) == 0 {
		return ""
	}
	result := methods[0]
	for i := 1; i < len(methods); i++ {
		result += ", " + methods[i]
	}
	return result
}

