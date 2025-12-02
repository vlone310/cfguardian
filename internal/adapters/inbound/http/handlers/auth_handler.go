package handlers

import (
	"encoding/json"
	"net/http"
	"time"

	"github.com/vlone310/cfguardian/internal/adapters/inbound/http/common"
	"github.com/vlone310/cfguardian/internal/adapters/inbound/http/middleware"
	"github.com/vlone310/cfguardian/internal/usecases/auth"
)

// AuthHandler handles authentication endpoints
type AuthHandler struct {
	loginUseCase         *auth.LoginUserUseCase
	registerUseCase      *auth.RegisterUserUseCase
	refreshTokenUseCase  *auth.RefreshTokenUseCase
	jwtSecret            string
	jwtExpiration        time.Duration
	refreshTokenExpiration time.Duration
}

// NewAuthHandler creates a new AuthHandler
func NewAuthHandler(
	loginUseCase *auth.LoginUserUseCase,
	registerUseCase *auth.RegisterUserUseCase,
	refreshTokenUseCase *auth.RefreshTokenUseCase,
	jwtSecret string,
	jwtExpiration time.Duration,
	refreshTokenExpiration time.Duration,
) *AuthHandler {
	return &AuthHandler{
		loginUseCase:         loginUseCase,
		registerUseCase:      registerUseCase,
		refreshTokenUseCase:  refreshTokenUseCase,
		jwtSecret:            jwtSecret,
		jwtExpiration:        jwtExpiration,
		refreshTokenExpiration: refreshTokenExpiration,
	}
}

// Register handles user registration
// POST /api/v1/auth/register
func (h *AuthHandler) Register(w http.ResponseWriter, r *http.Request) {
	var req auth.RegisterRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		common.BadRequest(w, "Invalid request body")
		return
	}
	
	resp, err := h.registerUseCase.Execute(r.Context(), req)
	if err != nil {
		common.BadRequest(w, err.Error())
		return
	}
	
	common.Created(w, resp)
}

// Login handles user login
// POST /api/v1/auth/login
func (h *AuthHandler) Login(w http.ResponseWriter, r *http.Request) {
	var req auth.LoginRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		common.BadRequest(w, "Invalid request body")
		return
	}
	
	resp, err := h.loginUseCase.Execute(r.Context(), req)
	if err != nil {
		common.Unauthorized(w, err.Error())
		return
	}
	
	// Generate access token
	accessToken, err := middleware.GenerateToken(
		resp.UserID,
		resp.Email,
		h.jwtSecret,
		h.jwtExpiration,
	)
	if err != nil {
		common.InternalServerError(w, "Failed to generate access token")
		return
	}
	
	// Generate refresh token
	refreshToken, err := middleware.GenerateRefreshToken(
		resp.UserID,
		resp.Email,
		h.jwtSecret,
		h.refreshTokenExpiration,
	)
	if err != nil {
		common.InternalServerError(w, "Failed to generate refresh token")
		return
	}
	
	// Return response with tokens
	common.OK(w, map[string]interface{}{
		"user_id":       resp.UserID,
		"email":         resp.Email,
		"access_token":  accessToken,
		"refresh_token": refreshToken,
		"token_type":    "Bearer",
		"expires_in":    int64(h.jwtExpiration.Seconds()),
	})
}

// RefreshToken handles token refresh
// POST /api/v1/auth/refresh
func (h *AuthHandler) RefreshToken(w http.ResponseWriter, r *http.Request) {
	var req auth.RefreshTokenRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		common.BadRequest(w, "Invalid request body")
		return
	}
	
	// Validate refresh token
	claims, err := middleware.ValidateRefreshToken(req.RefreshToken, h.jwtSecret)
	if err != nil {
		common.Unauthorized(w, "Invalid or expired refresh token")
		return
	}
	
	// Generate new access token
	newAccessToken, err := middleware.GenerateToken(
		claims.UserID,
		claims.Email,
		h.jwtSecret,
		h.jwtExpiration,
	)
	if err != nil {
		common.InternalServerError(w, "Failed to generate access token")
		return
	}
	
	// Generate new refresh token (token rotation)
	newRefreshToken, err := middleware.GenerateRefreshToken(
		claims.UserID,
		claims.Email,
		h.jwtSecret,
		h.refreshTokenExpiration,
	)
	if err != nil {
		common.InternalServerError(w, "Failed to generate refresh token")
		return
	}
	
	// Return new tokens
	common.OK(w, map[string]interface{}{
		"access_token":  newAccessToken,
		"refresh_token": newRefreshToken,
		"token_type":    "Bearer",
		"expires_in":    int64(h.jwtExpiration.Seconds()),
	})
}

