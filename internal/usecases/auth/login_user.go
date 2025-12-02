package auth

import (
	"context"
	"fmt"

	"github.com/vlone310/cfguardian/internal/domain/services"
	"github.com/vlone310/cfguardian/internal/ports/outbound"
)

// LoginRequest holds login request data
type LoginRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

// LoginResponse holds login response data
type LoginResponse struct {
	UserID string `json:"user_id"`
	Email  string `json:"email"`
	Token  string `json:"token"` // JWT token (will be generated in adapter layer)
}

// LoginUserUseCase handles user authentication
type LoginUserUseCase struct {
	userRepo       outbound.UserRepository
	passwordHasher *services.PasswordHasher
}

// NewLoginUserUseCase creates a new LoginUserUseCase
func NewLoginUserUseCase(
	userRepo outbound.UserRepository,
	passwordHasher *services.PasswordHasher,
) *LoginUserUseCase {
	return &LoginUserUseCase{
		userRepo:       userRepo,
		passwordHasher: passwordHasher,
	}
}

// Execute performs user login authentication
func (uc *LoginUserUseCase) Execute(ctx context.Context, req LoginRequest) (*LoginResponse, error) {
	// Validate input
	if req.Email == "" {
		return nil, fmt.Errorf("email is required")
	}
	if req.Password == "" {
		return nil, fmt.Errorf("password is required")
	}
	
	// Find user by email
	user, err := uc.userRepo.GetByEmail(ctx, req.Email)
	if err != nil {
		// Don't reveal whether user exists
		return nil, fmt.Errorf("invalid email or password")
	}
	
	// Verify password
	if err := uc.passwordHasher.Verify(req.Password, user.PasswordHash); err != nil {
		return nil, fmt.Errorf("invalid email or password")
	}
	
	// Return response (token generation will be handled by adapter layer)
	return &LoginResponse{
		UserID: user.ID,
		Email:  user.Email,
		Token:  "", // Will be populated by JWT middleware
	}, nil
}

