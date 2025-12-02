package handlers

import (
	"encoding/json"
	"net/http"

	"github.com/vlone310/cfguardian/internal/adapters/inbound/http/common"
	"github.com/vlone310/cfguardian/internal/usecases/auth"
)

// AuthHandler handles authentication endpoints
type AuthHandler struct {
	loginUseCase    *auth.LoginUserUseCase
	registerUseCase *auth.RegisterUserUseCase
}

// NewAuthHandler creates a new AuthHandler
func NewAuthHandler(
	loginUseCase *auth.LoginUserUseCase,
	registerUseCase *auth.RegisterUserUseCase,
) *AuthHandler {
	return &AuthHandler{
		loginUseCase:    loginUseCase,
		registerUseCase: registerUseCase,
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
	
	// TODO: Generate JWT token here
	// For now, we'll return the response as-is
	common.OK(w, resp)
}

