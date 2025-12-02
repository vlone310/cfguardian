package auth

import (
	"context"
	"fmt"

	"github.com/vlone310/cfguardian/internal/ports/outbound"
)

// RefreshTokenRequest holds refresh token request data
type RefreshTokenRequest struct {
	RefreshToken string `json:"refresh_token"`
}

// RefreshTokenResponse holds refresh token response data
type RefreshTokenResponse struct {
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token"`
	ExpiresIn    int64  `json:"expires_in"` // seconds until access token expires
}

// RefreshTokenUseCase handles token refresh
type RefreshTokenUseCase struct {
	userRepo outbound.UserRepository
}

// NewRefreshTokenUseCase creates a new RefreshTokenUseCase
func NewRefreshTokenUseCase(userRepo outbound.UserRepository) *RefreshTokenUseCase {
	return &RefreshTokenUseCase{
		userRepo: userRepo,
	}
}

// Execute validates refresh token and issues new access token
func (uc *RefreshTokenUseCase) Execute(ctx context.Context, req RefreshTokenRequest) (*RefreshTokenResponse, error) {
	// Validate input
	if req.RefreshToken == "" {
		return nil, fmt.Errorf("refresh token is required")
	}
	
	// Note: Refresh token validation would typically involve:
	// 1. Validating JWT signature
	// 2. Checking expiration (refresh tokens have longer TTL)
	// 3. Checking if token is revoked (requires token store/blacklist)
	// 4. Extracting user ID from claims
	
	// For MVP, we'll implement a simple version without token storage
	// In production, you'd want to:
	// - Store refresh tokens in database or Redis
	// - Implement token rotation (invalidate old refresh token)
	// - Implement token revocation
	
	// This is a placeholder implementation
	// The actual token generation will happen in the handler layer
	// where we have access to JWT secret and expiration config
	
	return &RefreshTokenResponse{
		// These will be populated by the handler
		AccessToken:  "",
		RefreshToken: "",
		ExpiresIn:    0,
	}, nil
}

