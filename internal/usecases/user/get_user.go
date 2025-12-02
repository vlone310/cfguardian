package user

import (
	"context"
	"fmt"

	"github.com/vlone310/cfguardian/internal/ports/outbound"
)

// GetUserRequest holds get user request data
type GetUserRequest struct {
	UserID string `json:"user_id"`
}

// GetUserResponse holds user data
type GetUserResponse struct {
	ID        string `json:"id"`
	Email     string `json:"email"`
	CreatedAt string `json:"created_at"`
	UpdatedAt string `json:"updated_at"`
}

// GetUserUseCase handles retrieving a user by ID
type GetUserUseCase struct {
	userRepo outbound.UserRepository
}

// NewGetUserUseCase creates a new GetUserUseCase
func NewGetUserUseCase(userRepo outbound.UserRepository) *GetUserUseCase {
	return &GetUserUseCase{
		userRepo: userRepo,
	}
}

// Execute retrieves a user by ID
func (uc *GetUserUseCase) Execute(ctx context.Context, req GetUserRequest) (*GetUserResponse, error) {
	if req.UserID == "" {
		return nil, fmt.Errorf("user ID is required")
	}
	
	// Get user from repository
	user, err := uc.userRepo.GetByID(ctx, req.UserID)
	if err != nil {
		return nil, fmt.Errorf("user not found: %w", err)
	}
	
	return &GetUserResponse{
		ID:        user.ID,
		Email:     user.Email,
		CreatedAt: user.CreatedAt,
		UpdatedAt: user.UpdatedAt,
	}, nil
}

