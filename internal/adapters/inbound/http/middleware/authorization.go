package middleware

import (
	"context"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/vlone310/cfguardian/internal/usecases/role"
)

// PermissionChecker defines the interface for checking permissions
type PermissionChecker interface {
	Execute(ctx context.Context, req role.CheckPermissionRequest) (*role.CheckPermissionResponse, error)
}

// AuthorizationConfig holds authorization configuration
type AuthorizationConfig struct {
	CheckPermission PermissionChecker
}

// RequireRole middleware requires a specific role level for the project
func RequireRole(cfg AuthorizationConfig, requiredRole string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// Get user ID from context (set by Auth middleware)
			userID := GetUserID(r.Context())
			if userID == "" {
				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(http.StatusUnauthorized)
				w.Write([]byte(`{"error":"User not authenticated","code":"UNAUTHORIZED"}`))
				return
			}

			// Get project ID from URL params
			projectID := chi.URLParam(r, "projectId")
			if projectID == "" {
				// No project context - allow (global admin check would go here)
				next.ServeHTTP(w, r)
				return
			}

			// Check permission
			resp, err := cfg.CheckPermission.Execute(r.Context(), role.CheckPermissionRequest{
				UserID:            userID,
				ProjectID:         projectID,
				RequiredRoleLevel: requiredRole,
			})

			if err != nil || !resp.Allowed {
				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(http.StatusForbidden)
				w.Write([]byte(`{"error":"Insufficient permissions","code":"FORBIDDEN"}`))
				return
			}

			next.ServeHTTP(w, r)
		})
	}
}

// RequireAdmin middleware requires admin role
func RequireAdmin(cfg AuthorizationConfig) func(http.Handler) http.Handler {
	return RequireRole(cfg, "admin")
}

// RequireEditor middleware requires at least editor role
func RequireEditor(cfg AuthorizationConfig) func(http.Handler) http.Handler {
	return RequireRole(cfg, "editor")
}

// RequireViewer middleware requires at least viewer role
func RequireViewer(cfg AuthorizationConfig) func(http.Handler) http.Handler {
	return RequireRole(cfg, "viewer")
}
