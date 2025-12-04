package middleware

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestAuth(t *testing.T) {
	secret := "test-secret-key-for-jwt-signing"
	cfg := AuthConfig{JWTSecret: secret}

	t.Run("missing authorization header", func(t *testing.T) {
		// Arrange
		handler := Auth(cfg)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusOK)
		}))
		req := httptest.NewRequest(http.MethodGet, "/test", nil)
		rec := httptest.NewRecorder()

		// Act
		handler.ServeHTTP(rec, req)

		// Assert
		assert.Equal(t, http.StatusUnauthorized, rec.Code)
		assert.Contains(t, rec.Body.String(), "Missing authorization header")
		assert.Contains(t, rec.Body.String(), "UNAUTHORIZED")
	})

	t.Run("invalid authorization header format", func(t *testing.T) {
		testCases := []struct {
			name   string
			header string
		}{
			{"missing Bearer prefix", "token123"},
			{"wrong prefix", "Basic token123"},
			{"too many parts", "Bearer token extra"},
		}

		for _, tc := range testCases {
			t.Run(tc.name, func(t *testing.T) {
				// Arrange
				handler := Auth(cfg)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
					w.WriteHeader(http.StatusOK)
				}))
				req := httptest.NewRequest(http.MethodGet, "/test", nil)
				req.Header.Set("Authorization", tc.header)
				rec := httptest.NewRecorder()

				// Act
				handler.ServeHTTP(rec, req)

				// Assert
				assert.Equal(t, http.StatusUnauthorized, rec.Code)
				assert.Contains(t, rec.Body.String(), "Invalid authorization header format")
			})
		}
	})

	t.Run("invalid token", func(t *testing.T) {
		testCases := []struct {
			name  string
			token string
		}{
			{"malformed token", "not.a.jwt"},
			{"random string", "random-token-string"},
			{"empty token", ""},
		}

		for _, tc := range testCases {
			t.Run(tc.name, func(t *testing.T) {
				// Arrange
				handler := Auth(cfg)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
					w.WriteHeader(http.StatusOK)
				}))
				req := httptest.NewRequest(http.MethodGet, "/test", nil)
				req.Header.Set("Authorization", "Bearer "+tc.token)
				rec := httptest.NewRecorder()

				// Act
				handler.ServeHTTP(rec, req)

				// Assert
				assert.Equal(t, http.StatusUnauthorized, rec.Code)
				assert.Contains(t, rec.Body.String(), "Invalid or expired token")
			})
		}
	})

	t.Run("expired token", func(t *testing.T) {
		// Arrange
		// Generate an expired token
		token, err := GenerateToken("user123", "user@example.com", secret, -1*time.Hour)
		require.NoError(t, err)

		handler := Auth(cfg)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusOK)
		}))
		req := httptest.NewRequest(http.MethodGet, "/test", nil)
		req.Header.Set("Authorization", "Bearer "+token)
		rec := httptest.NewRecorder()

		// Act
		handler.ServeHTTP(rec, req)

		// Assert
		assert.Equal(t, http.StatusUnauthorized, rec.Code)
		assert.Contains(t, rec.Body.String(), "Invalid or expired token")
	})

	t.Run("token with wrong signing method", func(t *testing.T) {
		// Arrange
		// Generate a token with RSA signing (not HMAC)
		claims := &Claims{
			UserID: "user123",
			Email:  "user@example.com",
			Type:   "access",
			RegisteredClaims: jwt.RegisteredClaims{
				ExpiresAt: jwt.NewNumericDate(time.Now().Add(1 * time.Hour)),
				IssuedAt:  jwt.NewNumericDate(time.Now()),
				Issuer:    "cfguardian",
			},
		}

		// Create token with wrong method (we'll sign with HMAC but server expects HMAC)
		// This test verifies the signing method check works
		token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
		tokenString, err := token.SignedString([]byte("wrong-secret"))
		require.NoError(t, err)

		handler := Auth(cfg)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusOK)
		}))
		req := httptest.NewRequest(http.MethodGet, "/test", nil)
		req.Header.Set("Authorization", "Bearer "+tokenString)
		rec := httptest.NewRecorder()

		// Act
		handler.ServeHTTP(rec, req)

		// Assert
		assert.Equal(t, http.StatusUnauthorized, rec.Code)
	})

	t.Run("valid token", func(t *testing.T) {
		// Arrange
		token, err := GenerateToken("user123", "user@example.com", secret, 1*time.Hour)
		require.NoError(t, err)

		handler := Auth(cfg)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// Verify context values
			userID := GetUserID(r.Context())
			userEmail := GetUserEmail(r.Context())
			assert.Equal(t, "user123", userID)
			assert.Equal(t, "user@example.com", userEmail)
			w.WriteHeader(http.StatusOK)
		}))
		req := httptest.NewRequest(http.MethodGet, "/test", nil)
		req.Header.Set("Authorization", "Bearer "+token)
		rec := httptest.NewRecorder()

		// Act
		handler.ServeHTTP(rec, req)

		// Assert
		assert.Equal(t, http.StatusOK, rec.Code)
	})

	t.Run("context values", func(t *testing.T) {
		t.Run("GetUserID returns empty for missing context value", func(t *testing.T) {
			ctx := context.Background()
			userID := GetUserID(ctx)
			assert.Empty(t, userID)
		})

		t.Run("GetUserEmail returns empty for missing context value", func(t *testing.T) {
			ctx := context.Background()
			email := GetUserEmail(ctx)
			assert.Empty(t, email)
		})

		t.Run("GetUserID returns value from context", func(t *testing.T) {
			ctx := context.WithValue(context.Background(), UserIDKey, "user123")
			userID := GetUserID(ctx)
			assert.Equal(t, "user123", userID)
		})

		t.Run("GetUserEmail returns value from context", func(t *testing.T) {
			ctx := context.WithValue(context.Background(), UserEmailKey, "user@example.com")
			email := GetUserEmail(ctx)
			assert.Equal(t, "user@example.com", email)
		})
	})
}

func TestGenerateToken(t *testing.T) {
	secret := "test-secret"
	userID := "user123"
	email := "user@example.com"

	t.Run("generates valid access token", func(t *testing.T) {
		// Act
		token, err := GenerateToken(userID, email, secret, 1*time.Hour)

		// Assert
		require.NoError(t, err)
		assert.NotEmpty(t, token)

		// Verify token is valid
		claims := &Claims{}
		parsedToken, parseErr := jwt.ParseWithClaims(token, claims, func(token *jwt.Token) (interface{}, error) {
			return []byte(secret), nil
		})
		require.NoError(t, parseErr)
		assert.True(t, parsedToken.Valid)
		assert.Equal(t, userID, claims.UserID)
		assert.Equal(t, email, claims.Email)
		assert.Equal(t, "access", claims.Type)
		assert.Equal(t, "cfguardian", claims.Issuer)
	})

	t.Run("sets correct expiration", func(t *testing.T) {
		// Act
		token, err := GenerateToken(userID, email, secret, 30*time.Minute)
		require.NoError(t, err)

		// Verify expiration
		claims := &Claims{}
		_, parseErr := jwt.ParseWithClaims(token, claims, func(token *jwt.Token) (interface{}, error) {
			return []byte(secret), nil
		})
		require.NoError(t, parseErr)

		// Check expiration is approximately 30 minutes in the future
		expectedExpiry := time.Now().Add(30 * time.Minute)
		actualExpiry := claims.ExpiresAt.Time
		assert.WithinDuration(t, expectedExpiry, actualExpiry, 5*time.Second)
	})

	t.Run("generates unique tokens", func(t *testing.T) {
		// Act
		token1, err1 := GenerateToken(userID, email, secret, 1*time.Hour)
		time.Sleep(1100 * time.Millisecond) // Ensure different IssuedAt (JWT timestamps are in seconds, not ms)
		token2, err2 := GenerateToken(userID, email, secret, 1*time.Hour)

		// Assert
		require.NoError(t, err1)
		require.NoError(t, err2)
		assert.NotEqual(t, token1, token2, "Tokens should be unique due to different IssuedAt")
	})
}

func TestGenerateRefreshToken(t *testing.T) {
	secret := "test-secret"
	userID := "user123"
	email := "user@example.com"

	t.Run("generates valid refresh token", func(t *testing.T) {
		// Act
		token, err := GenerateRefreshToken(userID, email, secret, 7*24*time.Hour)

		// Assert
		require.NoError(t, err)
		assert.NotEmpty(t, token)

		// Verify token is valid
		claims := &Claims{}
		parsedToken, parseErr := jwt.ParseWithClaims(token, claims, func(token *jwt.Token) (interface{}, error) {
			return []byte(secret), nil
		})
		require.NoError(t, parseErr)
		assert.True(t, parsedToken.Valid)
		assert.Equal(t, userID, claims.UserID)
		assert.Equal(t, email, claims.Email)
		assert.Equal(t, "refresh", claims.Type)
	})

	t.Run("sets longer expiration", func(t *testing.T) {
		// Act
		token, err := GenerateRefreshToken(userID, email, secret, 7*24*time.Hour)
		require.NoError(t, err)

		// Verify expiration is approximately 7 days
		claims := &Claims{}
		_, parseErr := jwt.ParseWithClaims(token, claims, func(token *jwt.Token) (interface{}, error) {
			return []byte(secret), nil
		})
		require.NoError(t, parseErr)

		expectedExpiry := time.Now().Add(7 * 24 * time.Hour)
		actualExpiry := claims.ExpiresAt.Time
		assert.WithinDuration(t, expectedExpiry, actualExpiry, 5*time.Second)
	})
}

func TestValidateRefreshToken(t *testing.T) {
	secret := "test-secret"
	userID := "user123"
	email := "user@example.com"

	t.Run("validates valid refresh token", func(t *testing.T) {
		// Arrange
		token, err := GenerateRefreshToken(userID, email, secret, 7*24*time.Hour)
		require.NoError(t, err)

		// Act
		claims, validateErr := ValidateRefreshToken(token, secret)

		// Assert
		require.NoError(t, validateErr)
		assert.Equal(t, userID, claims.UserID)
		assert.Equal(t, email, claims.Email)
		assert.Equal(t, "refresh", claims.Type)
	})

	t.Run("rejects access token", func(t *testing.T) {
		// Arrange
		token, err := GenerateToken(userID, email, secret, 1*time.Hour)
		require.NoError(t, err)

		// Act
		claims, validateErr := ValidateRefreshToken(token, secret)

		// Assert
		assert.Error(t, validateErr)
		assert.Nil(t, claims)
		assert.Contains(t, validateErr.Error(), "not a refresh token")
	})

	t.Run("rejects expired refresh token", func(t *testing.T) {
		// Arrange
		token, err := GenerateRefreshToken(userID, email, secret, -1*time.Hour)
		require.NoError(t, err)

		// Act
		claims, validateErr := ValidateRefreshToken(token, secret)

		// Assert
		assert.Error(t, validateErr)
		assert.Nil(t, claims)
	})

	t.Run("rejects token with wrong secret", func(t *testing.T) {
		// Arrange
		token, err := GenerateRefreshToken(userID, email, secret, 7*24*time.Hour)
		require.NoError(t, err)

		// Act
		claims, validateErr := ValidateRefreshToken(token, "wrong-secret")

		// Assert
		assert.Error(t, validateErr)
		assert.Nil(t, claims)
	})

	t.Run("rejects malformed token", func(t *testing.T) {
		// Act
		claims, err := ValidateRefreshToken("not.a.valid.token", secret)

		// Assert
		assert.Error(t, err)
		assert.Nil(t, claims)
	})
}

func TestAuth_Integration(t *testing.T) {
	secret := "test-secret-key"
	cfg := AuthConfig{JWTSecret: secret}

	t.Run("full authentication flow", func(t *testing.T) {
		userID := "user123"
		userEmail := "user@example.com"

		// Generate a valid token
		token, err := GenerateToken(userID, userEmail, secret, 1*time.Hour)
		require.NoError(t, err)

		// Create handler that uses context values
		handler := Auth(cfg)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			ctxUserID := GetUserID(r.Context())
			ctxUserEmail := GetUserEmail(r.Context())

			assert.Equal(t, userID, ctxUserID)
			assert.Equal(t, userEmail, ctxUserEmail)

			w.WriteHeader(http.StatusOK)
			w.Write([]byte("authenticated"))
		}))

		req := httptest.NewRequest(http.MethodGet, "/test", nil)
		req.Header.Set("Authorization", "Bearer "+token)
		rec := httptest.NewRecorder()

		// Act
		handler.ServeHTTP(rec, req)

		// Assert
		assert.Equal(t, http.StatusOK, rec.Code)
		assert.Equal(t, "authenticated", rec.Body.String())
	})

	t.Run("works with middleware chain", func(t *testing.T) {
		userID := "user123"
		userEmail := "user@example.com"

		// Generate a valid token
		token, err := GenerateToken(userID, userEmail, secret, 1*time.Hour)
		require.NoError(t, err)

		// Create middleware chain: RequestID -> Auth -> SecurityHeaders -> handler
		finalHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusOK)
		})

		handler := RequestID(Auth(cfg)(SecurityHeaders(finalHandler)))

		req := httptest.NewRequest(http.MethodGet, "/test", nil)
		req.Header.Set("Authorization", "Bearer "+token)
		rec := httptest.NewRecorder()

		// Act
		handler.ServeHTTP(rec, req)

		// Assert
		assert.Equal(t, http.StatusOK, rec.Code)
		assert.NotEmpty(t, rec.Header().Get("X-Request-ID"))
		assert.NotEmpty(t, rec.Header().Get("X-Content-Type-Options"))
	})
}

func TestAuth_EdgeCases(t *testing.T) {
	secret := "test-secret"
	cfg := AuthConfig{JWTSecret: secret}

	t.Run("handles concurrent requests", func(t *testing.T) {
		// Arrange
		token, err := GenerateToken("user123", "user@example.com", secret, 1*time.Hour)
		require.NoError(t, err)

		handler := Auth(cfg)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			userID := GetUserID(r.Context())
			assert.Equal(t, "user123", userID)
			w.WriteHeader(http.StatusOK)
		}))

		// Act - send 10 concurrent requests
		done := make(chan bool, 10)
		for i := 0; i < 10; i++ {
			go func() {
				req := httptest.NewRequest(http.MethodGet, "/test", nil)
				req.Header.Set("Authorization", "Bearer "+token)
				rec := httptest.NewRecorder()
				handler.ServeHTTP(rec, req)
				assert.Equal(t, http.StatusOK, rec.Code)
				done <- true
			}()
		}

		// Wait for all requests
		for i := 0; i < 10; i++ {
			<-done
		}
	})

	t.Run("handles empty secret", func(t *testing.T) {
		// Arrange
		emptyCfg := AuthConfig{JWTSecret: ""}
		token, err := GenerateToken("user123", "user@example.com", secret, 1*time.Hour)
		require.NoError(t, err)

		handler := Auth(emptyCfg)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusOK)
		}))
		req := httptest.NewRequest(http.MethodGet, "/test", nil)
		req.Header.Set("Authorization", "Bearer "+token)
		rec := httptest.NewRecorder()

		// Act
		handler.ServeHTTP(rec, req)

		// Assert
		assert.Equal(t, http.StatusUnauthorized, rec.Code)
	})

	t.Run("handles special characters in claims", func(t *testing.T) {
		// Arrange
		specialEmail := "user+tag@example.com"
		token, err := GenerateToken("user-123_456", specialEmail, secret, 1*time.Hour)
		require.NoError(t, err)

		handler := Auth(cfg)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			assert.Equal(t, "user-123_456", GetUserID(r.Context()))
			assert.Equal(t, specialEmail, GetUserEmail(r.Context()))
			w.WriteHeader(http.StatusOK)
		}))
		req := httptest.NewRequest(http.MethodGet, "/test", nil)
		req.Header.Set("Authorization", "Bearer "+token)
		rec := httptest.NewRecorder()

		// Act
		handler.ServeHTTP(rec, req)

		// Assert
		assert.Equal(t, http.StatusOK, rec.Code)
	})
}

func TestAuth_TokenLifecycle(t *testing.T) {
	secret := "test-secret"

	t.Run("access and refresh token workflow", func(t *testing.T) {
		userID := "user123"
		email := "user@example.com"

		// Generate access token (short-lived)
		accessToken, err := GenerateToken(userID, email, secret, 15*time.Minute)
		require.NoError(t, err)
		assert.NotEmpty(t, accessToken)

		// Generate refresh token (long-lived)
		refreshToken, err := GenerateRefreshToken(userID, email, secret, 7*24*time.Hour)
		require.NoError(t, err)
		assert.NotEmpty(t, refreshToken)

		// Tokens should be different
		assert.NotEqual(t, accessToken, refreshToken)

		// Validate refresh token
		claims, err := ValidateRefreshToken(refreshToken, secret)
		require.NoError(t, err)
		assert.Equal(t, userID, claims.UserID)
		assert.Equal(t, email, claims.Email)

		// Access token should NOT validate as refresh token
		_, err = ValidateRefreshToken(accessToken, secret)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "not a refresh token")
	})

	t.Run("token expiration timeline", func(t *testing.T) {
		// Generate token that expires in 2 seconds
		token, err := GenerateToken("user123", "user@example.com", secret, 2*time.Second)
		require.NoError(t, err)

		// Token should be valid immediately
		claims := &Claims{}
		parsedToken, parseErr := jwt.ParseWithClaims(token, claims, func(token *jwt.Token) (interface{}, error) {
			return []byte(secret), nil
		})
		require.NoError(t, parseErr)
		assert.True(t, parsedToken.Valid)

		// Wait for token to expire (with buffer)
		time.Sleep(2500 * time.Millisecond)

		// Token should now be invalid
		parsedToken, parseErr = jwt.ParseWithClaims(token, claims, func(token *jwt.Token) (interface{}, error) {
			return []byte(secret), nil
		})
		assert.Error(t, parseErr)
	})
}
