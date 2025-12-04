package middleware

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"github.com/vlone310/cfguardian/internal/usecases/role"
)

// MockPermissionChecker is a mock for PermissionChecker
type MockPermissionChecker struct {
	mock.Mock
}

func (m *MockPermissionChecker) Execute(ctx context.Context, req role.CheckPermissionRequest) (*role.CheckPermissionResponse, error) {
	args := m.Called(ctx, req)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*role.CheckPermissionResponse), args.Error(1)
}

func TestRequireRole(t *testing.T) {
	t.Run("missing user ID in context", func(t *testing.T) {
		// Arrange
		mockUseCase := new(MockPermissionChecker)
		cfg := AuthorizationConfig{CheckPermission: mockUseCase}

		handler := RequireRole(cfg, "admin")(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusOK)
		}))

		req := httptest.NewRequest(http.MethodGet, "/test", nil)
		rec := httptest.NewRecorder()

		// Act
		handler.ServeHTTP(rec, req)

		// Assert
		assert.Equal(t, http.StatusUnauthorized, rec.Code)
		assert.Contains(t, rec.Body.String(), "User not authenticated")
		assert.Contains(t, rec.Body.String(), "UNAUTHORIZED")
		mockUseCase.AssertNotCalled(t, "Execute")
	})

	t.Run("no project context", func(t *testing.T) {
		// Arrange
		mockUseCase := new(MockPermissionChecker)
		cfg := AuthorizationConfig{CheckPermission: mockUseCase}

		handler := RequireRole(cfg, "admin")(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusOK)
		}))

		// Add user ID to context
		ctx := context.WithValue(context.Background(), UserIDKey, "user123")
		req := httptest.NewRequest(http.MethodGet, "/test", nil).WithContext(ctx)
		rec := httptest.NewRecorder()

		// Act
		handler.ServeHTTP(rec, req)

		// Assert
		assert.Equal(t, http.StatusOK, rec.Code)
		mockUseCase.AssertNotCalled(t, "Execute")
	})

	t.Run("user has permission", func(t *testing.T) {
		// Arrange
		mockUseCase := new(MockPermissionChecker)
		mockUseCase.On("Execute", mock.Anything, role.CheckPermissionRequest{
			UserID:            "user123",
			ProjectID:         "project456",
			RequiredRoleLevel: "admin",
		}).Return(&role.CheckPermissionResponse{
			Allowed:       true,
			UserRoleLevel: "admin",
		}, nil)

		cfg := AuthorizationConfig{CheckPermission: mockUseCase}

		// Create chi router to simulate URL params
		router := chi.NewRouter()
		router.With(func(next http.Handler) http.Handler {
			return RequireRole(cfg, "admin")(next)
		}).Get("/projects/{projectId}/configs", func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusOK)
			w.Write([]byte("success"))
		})

		// Add user ID to context
		ctx := context.WithValue(context.Background(), UserIDKey, "user123")
		req := httptest.NewRequest(http.MethodGet, "/projects/project456/configs", nil).WithContext(ctx)
		rec := httptest.NewRecorder()

		// Act
		router.ServeHTTP(rec, req)

		// Assert
		assert.Equal(t, http.StatusOK, rec.Code)
		assert.Equal(t, "success", rec.Body.String())
		mockUseCase.AssertExpectations(t)
	})

	t.Run("user lacks permission", func(t *testing.T) {
		// Arrange
		mockUseCase := new(MockPermissionChecker)
		mockUseCase.On("Execute", mock.Anything, role.CheckPermissionRequest{
			UserID:            "user123",
			ProjectID:         "project456",
			RequiredRoleLevel: "admin",
		}).Return(&role.CheckPermissionResponse{
			Allowed:       false,
			UserRoleLevel: "viewer",
		}, nil)

		cfg := AuthorizationConfig{CheckPermission: mockUseCase}

		// Create chi router
		router := chi.NewRouter()
		router.With(func(next http.Handler) http.Handler {
			return RequireRole(cfg, "admin")(next)
		}).Get("/projects/{projectId}/configs", func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusOK)
		})

		// Add user ID to context
		ctx := context.WithValue(context.Background(), UserIDKey, "user123")
		req := httptest.NewRequest(http.MethodGet, "/projects/project456/configs", nil).WithContext(ctx)
		rec := httptest.NewRecorder()

		// Act
		router.ServeHTTP(rec, req)

		// Assert
		assert.Equal(t, http.StatusForbidden, rec.Code)
		assert.Contains(t, rec.Body.String(), "Insufficient permissions")
		assert.Contains(t, rec.Body.String(), "FORBIDDEN")
		mockUseCase.AssertExpectations(t)
	})

	t.Run("permission check returns error", func(t *testing.T) {
		// Arrange
		mockUseCase := new(MockPermissionChecker)
		mockUseCase.On("Execute", mock.Anything, mock.Anything).Return(
			nil, errors.New("database error"),
		)

		cfg := AuthorizationConfig{CheckPermission: mockUseCase}

		// Create chi router
		router := chi.NewRouter()
		router.With(func(next http.Handler) http.Handler {
			return RequireRole(cfg, "admin")(next)
		}).Get("/projects/{projectId}/configs", func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusOK)
		})

		// Add user ID to context
		ctx := context.WithValue(context.Background(), UserIDKey, "user123")
		req := httptest.NewRequest(http.MethodGet, "/projects/project456/configs", nil).WithContext(ctx)
		rec := httptest.NewRecorder()

		// Act
		router.ServeHTTP(rec, req)

		// Assert
		assert.Equal(t, http.StatusForbidden, rec.Code)
		mockUseCase.AssertExpectations(t)
	})
}

func TestRequireAdmin(t *testing.T) {
	t.Run("requires admin role", func(t *testing.T) {
		// Arrange
		mockUseCase := new(MockPermissionChecker)
		mockUseCase.On("Execute", mock.Anything, role.CheckPermissionRequest{
			UserID:            "user123",
			ProjectID:         "project456",
			RequiredRoleLevel: "admin",
		}).Return(&role.CheckPermissionResponse{
			Allowed:       true,
			UserRoleLevel: "admin",
		}, nil)

		cfg := AuthorizationConfig{CheckPermission: mockUseCase}

		// Create chi router
		router := chi.NewRouter()
		router.With(func(next http.Handler) http.Handler {
			return RequireAdmin(cfg)(next)
		}).Delete("/projects/{projectId}", func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusNoContent)
		})

		// Add user ID to context
		ctx := context.WithValue(context.Background(), UserIDKey, "user123")
		req := httptest.NewRequest(http.MethodDelete, "/projects/project456", nil).WithContext(ctx)
		rec := httptest.NewRecorder()

		// Act
		router.ServeHTTP(rec, req)

		// Assert
		assert.Equal(t, http.StatusNoContent, rec.Code)
		mockUseCase.AssertExpectations(t)
	})

	t.Run("rejects non-admin", func(t *testing.T) {
		// Arrange
		mockUseCase := new(MockPermissionChecker)
		mockUseCase.On("Execute", mock.Anything, mock.Anything).Return(&role.CheckPermissionResponse{
			Allowed:       false,
			UserRoleLevel: "editor",
		}, nil)

		cfg := AuthorizationConfig{CheckPermission: mockUseCase}

		router := chi.NewRouter()
		router.With(func(next http.Handler) http.Handler {
			return RequireAdmin(cfg)(next)
		}).Delete("/projects/{projectId}", func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusNoContent)
		})

		ctx := context.WithValue(context.Background(), UserIDKey, "user123")
		req := httptest.NewRequest(http.MethodDelete, "/projects/project456", nil).WithContext(ctx)
		rec := httptest.NewRecorder()

		// Act
		router.ServeHTTP(rec, req)

		// Assert
		assert.Equal(t, http.StatusForbidden, rec.Code)
	})
}

func TestRequireEditor(t *testing.T) {
	t.Run("allows editor", func(t *testing.T) {
		// Arrange
		mockUseCase := new(MockPermissionChecker)
		mockUseCase.On("Execute", mock.Anything, role.CheckPermissionRequest{
			UserID:            "user123",
			ProjectID:         "project456",
			RequiredRoleLevel: "editor",
		}).Return(&role.CheckPermissionResponse{
			Allowed:       true,
			UserRoleLevel: "editor",
		}, nil)

		cfg := AuthorizationConfig{CheckPermission: mockUseCase}

		router := chi.NewRouter()
		router.With(func(next http.Handler) http.Handler {
			return RequireEditor(cfg)(next)
		}).Put("/projects/{projectId}/configs/{key}", func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusOK)
		})

		ctx := context.WithValue(context.Background(), UserIDKey, "user123")
		req := httptest.NewRequest(http.MethodPut, "/projects/project456/configs/app-config", nil).WithContext(ctx)
		rec := httptest.NewRecorder()

		// Act
		router.ServeHTTP(rec, req)

		// Assert
		assert.Equal(t, http.StatusOK, rec.Code)
		mockUseCase.AssertExpectations(t)
	})

	t.Run("rejects viewer", func(t *testing.T) {
		// Arrange
		mockUseCase := new(MockPermissionChecker)
		mockUseCase.On("Execute", mock.Anything, mock.Anything).Return(&role.CheckPermissionResponse{
			Allowed:       false,
			UserRoleLevel: "viewer",
		}, nil)

		cfg := AuthorizationConfig{CheckPermission: mockUseCase}

		router := chi.NewRouter()
		router.With(func(next http.Handler) http.Handler {
			return RequireEditor(cfg)(next)
		}).Put("/projects/{projectId}/configs/{key}", func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusOK)
		})

		ctx := context.WithValue(context.Background(), UserIDKey, "user123")
		req := httptest.NewRequest(http.MethodPut, "/projects/project456/configs/app-config", nil).WithContext(ctx)
		rec := httptest.NewRecorder()

		// Act
		router.ServeHTTP(rec, req)

		// Assert
		assert.Equal(t, http.StatusForbidden, rec.Code)
	})
}

func TestRequireViewer(t *testing.T) {
	t.Run("allows viewer", func(t *testing.T) {
		// Arrange
		mockUseCase := new(MockPermissionChecker)
		mockUseCase.On("Execute", mock.Anything, role.CheckPermissionRequest{
			UserID:            "user123",
			ProjectID:         "project456",
			RequiredRoleLevel: "viewer",
		}).Return(&role.CheckPermissionResponse{
			Allowed:       true,
			UserRoleLevel: "viewer",
		}, nil)

		cfg := AuthorizationConfig{CheckPermission: mockUseCase}

		router := chi.NewRouter()
		router.With(func(next http.Handler) http.Handler {
			return RequireViewer(cfg)(next)
		}).Get("/projects/{projectId}/configs", func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusOK)
		})

		ctx := context.WithValue(context.Background(), UserIDKey, "user123")
		req := httptest.NewRequest(http.MethodGet, "/projects/project456/configs", nil).WithContext(ctx)
		rec := httptest.NewRecorder()

		// Act
		router.ServeHTTP(rec, req)

		// Assert
		assert.Equal(t, http.StatusOK, rec.Code)
		mockUseCase.AssertExpectations(t)
	})
}

func TestAuthorization_Integration(t *testing.T) {
	t.Run("full auth and authz chain", func(t *testing.T) {
		// Arrange
		secret := "test-secret"
		authCfg := AuthConfig{JWTSecret: secret}

		mockUseCase := new(MockPermissionChecker)
		mockUseCase.On("Execute", mock.Anything, mock.Anything).Return(&role.CheckPermissionResponse{
			Allowed:       true,
			UserRoleLevel: "admin",
		}, nil)

		authzCfg := AuthorizationConfig{CheckPermission: mockUseCase}

		// Generate valid token
		token, err := GenerateToken("user123", "user@example.com", secret, 1*time.Hour)
		require.NoError(t, err)

		// Create router with auth chain
		router := chi.NewRouter()
		router.With(
			func(next http.Handler) http.Handler {
				return Auth(authCfg)(next)
			},
			func(next http.Handler) http.Handler {
				return RequireAdmin(authzCfg)(next)
			},
		).Delete("/projects/{projectId}", func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusNoContent)
		})

		req := httptest.NewRequest(http.MethodDelete, "/projects/project456", nil)
		req.Header.Set("Authorization", "Bearer "+token)
		rec := httptest.NewRecorder()

		// Act
		router.ServeHTTP(rec, req)

		// Assert
		assert.Equal(t, http.StatusNoContent, rec.Code)
		mockUseCase.AssertExpectations(t)
	})

	t.Run("blocks unauthenticated before authorization check", func(t *testing.T) {
		// Arrange
		secret := "test-secret"
		authCfg := AuthConfig{JWTSecret: secret}

		mockUseCase := new(MockPermissionChecker)
		authzCfg := AuthorizationConfig{CheckPermission: mockUseCase}

		router := chi.NewRouter()
		router.With(
			func(next http.Handler) http.Handler {
				return Auth(authCfg)(next)
			},
			func(next http.Handler) http.Handler {
				return RequireAdmin(authzCfg)(next)
			},
		).Delete("/projects/{projectId}", func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusNoContent)
		})

		// No auth header
		req := httptest.NewRequest(http.MethodDelete, "/projects/project456", nil)
		rec := httptest.NewRecorder()

		// Act
		router.ServeHTTP(rec, req)

		// Assert
		assert.Equal(t, http.StatusUnauthorized, rec.Code)
		mockUseCase.AssertNotCalled(t, "Execute", "Authorization check should not run without authentication")
	})

	t.Run("blocks unauthorized after authentication", func(t *testing.T) {
		// Arrange
		secret := "test-secret"
		authCfg := AuthConfig{JWTSecret: secret}

		mockUseCase := new(MockPermissionChecker)
		mockUseCase.On("Execute", mock.Anything, mock.Anything).Return(&role.CheckPermissionResponse{
			Allowed:       false,
			UserRoleLevel: "viewer",
		}, nil)

		authzCfg := AuthorizationConfig{CheckPermission: mockUseCase}

		token, err := GenerateToken("user123", "user@example.com", secret, 1*time.Hour)
		require.NoError(t, err)

		router := chi.NewRouter()
		router.With(
			func(next http.Handler) http.Handler {
				return Auth(authCfg)(next)
			},
			func(next http.Handler) http.Handler {
				return RequireAdmin(authzCfg)(next)
			},
		).Delete("/projects/{projectId}", func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusNoContent)
		})

		req := httptest.NewRequest(http.MethodDelete, "/projects/project456", nil)
		req.Header.Set("Authorization", "Bearer "+token)
		rec := httptest.NewRecorder()

		// Act
		router.ServeHTTP(rec, req)

		// Assert
		assert.Equal(t, http.StatusForbidden, rec.Code)
		assert.Contains(t, rec.Body.String(), "Insufficient permissions")
		mockUseCase.AssertExpectations(t)
	})
}

func TestAuthorization_RoleHierarchy(t *testing.T) {
	t.Run("admin can access editor endpoints", func(t *testing.T) {
		// Arrange
		mockUseCase := new(MockPermissionChecker)
		mockUseCase.On("Execute", mock.Anything, mock.Anything).Return(&role.CheckPermissionResponse{
			Allowed:       true,
			UserRoleLevel: "admin",
		}, nil)

		cfg := AuthorizationConfig{CheckPermission: mockUseCase}

		router := chi.NewRouter()
		router.With(func(next http.Handler) http.Handler {
			return RequireEditor(cfg)(next)
		}).Put("/projects/{projectId}/configs/{key}", func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusOK)
		})

		ctx := context.WithValue(context.Background(), UserIDKey, "user123")
		req := httptest.NewRequest(http.MethodPut, "/projects/project456/configs/app-config", nil).WithContext(ctx)
		rec := httptest.NewRecorder()

		// Act
		router.ServeHTTP(rec, req)

		// Assert
		assert.Equal(t, http.StatusOK, rec.Code)
	})

	t.Run("editor can access viewer endpoints", func(t *testing.T) {
		// Arrange
		mockUseCase := new(MockPermissionChecker)
		mockUseCase.On("Execute", mock.Anything, mock.Anything).Return(&role.CheckPermissionResponse{
			Allowed:       true,
			UserRoleLevel: "editor",
		}, nil)

		cfg := AuthorizationConfig{CheckPermission: mockUseCase}

		router := chi.NewRouter()
		router.With(func(next http.Handler) http.Handler {
			return RequireViewer(cfg)(next)
		}).Get("/projects/{projectId}/configs", func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusOK)
		})

		ctx := context.WithValue(context.Background(), UserIDKey, "user123")
		req := httptest.NewRequest(http.MethodGet, "/projects/project456/configs", nil).WithContext(ctx)
		rec := httptest.NewRecorder()

		// Act
		router.ServeHTTP(rec, req)

		// Assert
		assert.Equal(t, http.StatusOK, rec.Code)
	})

	t.Run("viewer cannot access editor endpoints", func(t *testing.T) {
		// Arrange
		mockUseCase := new(MockPermissionChecker)
		mockUseCase.On("Execute", mock.Anything, mock.Anything).Return(&role.CheckPermissionResponse{
			Allowed:       false,
			UserRoleLevel: "viewer",
		}, nil)

		cfg := AuthorizationConfig{CheckPermission: mockUseCase}

		router := chi.NewRouter()
		router.With(func(next http.Handler) http.Handler {
			return RequireEditor(cfg)(next)
		}).Put("/projects/{projectId}/configs/{key}", func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusOK)
		})

		ctx := context.WithValue(context.Background(), UserIDKey, "user123")
		req := httptest.NewRequest(http.MethodPut, "/projects/project456/configs/app-config", nil).WithContext(ctx)
		rec := httptest.NewRecorder()

		// Act
		router.ServeHTTP(rec, req)

		// Assert
		assert.Equal(t, http.StatusForbidden, rec.Code)
	})
}
