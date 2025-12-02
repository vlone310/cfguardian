package http

import (
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"
	chiMiddleware "github.com/go-chi/chi/v5/middleware"
	"github.com/vlone310/cfguardian/internal/adapters/inbound/http/common"
	"github.com/vlone310/cfguardian/internal/adapters/inbound/http/handlers"
	"github.com/vlone310/cfguardian/internal/adapters/inbound/http/middleware"
)

// RouterConfig holds router configuration
type RouterConfig struct {
	JWTSecret          string
	RateLimitRPS       float64
	RateLimitBurst     int
	AuthHandler        *handlers.AuthHandler
	UserHandler        *handlers.UserHandler
	ProjectHandler     *handlers.ProjectHandler
	RoleHandler        *handlers.RoleHandler
	SchemaHandler      *handlers.SchemaHandler
	ConfigHandler      *handlers.ConfigHandler
	ReadHandler        *handlers.ReadHandler
	AuthorizationConfig middleware.AuthorizationConfig
}

// NewRouter creates a new HTTP router with all routes and middleware
func NewRouter(cfg RouterConfig) *chi.Mux {
	r := chi.NewRouter()
	
	// Global middleware (applied to all routes)
	r.Use(middleware.RequestID)
	r.Use(middleware.Recovery)
	r.Use(middleware.Logging)
	r.Use(middleware.CORS())
	r.Use(chiMiddleware.Compress(5))
	r.Use(chiMiddleware.Timeout(60 * time.Second))
	
	// Rate limiting (if configured)
	if cfg.RateLimitRPS > 0 {
		rateLimiter := middleware.NewRateLimiter(cfg.RateLimitRPS, cfg.RateLimitBurst)
		r.Use(middleware.RateLimit(rateLimiter))
	}
	
	// Health check (no auth required)
	r.Get("/health", func(w http.ResponseWriter, r *http.Request) {
		common.OK(w, map[string]string{"status": "healthy"})
	})
	
	r.Get("/", func(w http.ResponseWriter, r *http.Request) {
		common.OK(w, map[string]string{
			"service": "GoConfig Guardian",
			"version": "1.0.0",
			"status":  "running",
		})
	})
	
	// API v1 routes
	r.Route("/api/v1", func(r chi.Router) {
		// Public routes (no authentication required)
		r.Group(func(r chi.Router) {
			// Authentication
			r.Post("/auth/register", cfg.AuthHandler.Register)
			r.Post("/auth/login", cfg.AuthHandler.Login)
			
			// Public read API (API key in URL path)
			r.Get("/read/{apiKey}/{configKey}", cfg.ReadHandler.Read)
		})
		
		// Protected routes (JWT authentication required)
		r.Group(func(r chi.Router) {
			// Apply JWT authentication
			r.Use(middleware.Auth(middleware.AuthConfig{
				JWTSecret: cfg.JWTSecret,
			}))
			
			// Users
			r.Route("/users", func(r chi.Router) {
				r.Get("/", cfg.UserHandler.List)
				r.Post("/", cfg.UserHandler.Create)
				r.Get("/{userId}", cfg.UserHandler.Get)
				r.Delete("/{userId}", cfg.UserHandler.Delete)
			})
			
			// Projects
			r.Route("/projects", func(r chi.Router) {
				r.Get("/", cfg.ProjectHandler.List)
				r.Post("/", cfg.ProjectHandler.Create)
				
				// Project-specific routes
				r.Route("/{projectId}", func(r chi.Router) {
					r.Get("/", cfg.ProjectHandler.Get)
					r.Delete("/", cfg.ProjectHandler.Delete)
					
					// Roles (require at least viewer to list, admin to modify)
					r.Route("/roles", func(r chi.Router) {
						r.With(middleware.RequireViewer(cfg.AuthorizationConfig)).Get("/", func(w http.ResponseWriter, r *http.Request) {
							// TODO: Implement ListProjectRoles
							common.OK(w, []interface{}{})
						})
						r.With(middleware.RequireAdmin(cfg.AuthorizationConfig)).Post("/", cfg.RoleHandler.Assign)
						r.With(middleware.RequireAdmin(cfg.AuthorizationConfig)).Delete("/{userId}", cfg.RoleHandler.Revoke)
					})
					
					// Configs (require appropriate roles)
					r.Route("/configs", func(r chi.Router) {
						// List and create
						r.With(middleware.RequireViewer(cfg.AuthorizationConfig)).Get("/", func(w http.ResponseWriter, r *http.Request) {
							// TODO: Implement ListConfigs
							common.OK(w, []interface{}{})
						})
						r.With(middleware.RequireEditor(cfg.AuthorizationConfig)).Post("/", cfg.ConfigHandler.Create)
						
						// Individual config operations
						r.Route("/{configKey}", func(r chi.Router) {
							r.With(middleware.RequireViewer(cfg.AuthorizationConfig)).Get("/", cfg.ConfigHandler.Get)
							r.With(middleware.RequireEditor(cfg.AuthorizationConfig)).Put("/", cfg.ConfigHandler.Update)
							r.With(middleware.RequireAdmin(cfg.AuthorizationConfig)).Delete("/", cfg.ConfigHandler.Delete)
							
							// Rollback (admin only)
							r.With(middleware.RequireAdmin(cfg.AuthorizationConfig)).Post("/rollback", cfg.ConfigHandler.Rollback)
						})
					})
				})
			})
			
			// Schemas (global resource)
			r.Route("/schemas", func(r chi.Router) {
				r.Get("/", cfg.SchemaHandler.List)
				r.Post("/", cfg.SchemaHandler.Create)
				r.Put("/{schemaId}", cfg.SchemaHandler.Update)
				r.Delete("/{schemaId}", cfg.SchemaHandler.Delete)
			})
		})
	})
	
	return r
}

