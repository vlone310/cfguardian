package user

import (
	"context"
	"fmt"

	"github.com/vlone310/cfguardian/internal/ports/outbound"
)

// DeleteUserRequest holds delete user request data
type DeleteUserRequest struct {
	UserID string `json:"user_id"`
}

// DeleteUserUseCase handles user deletion (admin only)
type DeleteUserUseCase struct {
	userRepo outbound.UserRepository
	roleRepo outbound.RoleRepository
}

// NewDeleteUserUseCase creates a new DeleteUserUseCase
func NewDeleteUserUseCase(
	userRepo outbound.UserRepository,
	roleRepo outbound.RoleRepository,
) *DeleteUserUseCase {
	return &DeleteUserUseCase{
		userRepo: userRepo,
		roleRepo: roleRepo,
	}
}

// Execute deletes a user
func (uc *DeleteUserUseCase) Execute(ctx context.Context, req DeleteUserRequest) error {
	if req.UserID == "" {
		return fmt.Errorf("user ID is required")
	}
	
	// Check if user exists
	exists, err := uc.userRepo.Exists(ctx, req.UserID)
	if err != nil {
		return fmt.Errorf("failed to check if user exists: %w", err)
	}
	if !exists {
		return fmt.Errorf("user not found")
	}
	
	// Note: Foreign key CASCADE will automatically delete:
	// - User's projects (and their configs, revisions, roles)
	// - User's roles in other projects
	// - User's created schemas
	
	// Delete user (cascading deletes handled by database)
	if err := uc.userRepo.Delete(ctx, req.UserID); err != nil {
		return fmt.Errorf("failed to delete user: %w", err)
	}
	
	return nil
}

