package role

import (
	"context"
	"fmt"

	"github.com/vlone310/cfguardian/internal/ports/outbound"
)

// RevokeRoleRequest holds role revocation data
type RevokeRoleRequest struct {
	UserID    string `json:"user_id"`
	ProjectID string `json:"project_id"`
}

// RevokeRoleUseCase handles role revocation
type RevokeRoleUseCase struct {
	roleRepo outbound.RoleRepository
}

// NewRevokeRoleUseCase creates a new RevokeRoleUseCase
func NewRevokeRoleUseCase(roleRepo outbound.RoleRepository) *RevokeRoleUseCase {
	return &RevokeRoleUseCase{
		roleRepo: roleRepo,
	}
}

// Execute revokes a role for a user in a project
func (uc *RevokeRoleUseCase) Execute(ctx context.Context, req RevokeRoleRequest) error {
	// Validate input
	if req.UserID == "" {
		return fmt.Errorf("user ID is required")
	}
	if req.ProjectID == "" {
		return fmt.Errorf("project ID is required")
	}
	
	// Check if role exists
	exists, err := uc.roleRepo.Exists(ctx, req.UserID, req.ProjectID)
	if err != nil {
		return fmt.Errorf("failed to check if role exists: %w", err)
	}
	if !exists {
		return fmt.Errorf("role not found")
	}
	
	// Revoke role
	if err := uc.roleRepo.Revoke(ctx, req.UserID, req.ProjectID); err != nil {
		return fmt.Errorf("failed to revoke role: %w", err)
	}
	
	return nil
}

