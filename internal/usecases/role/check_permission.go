package role

import (
	"context"
	"fmt"

	"github.com/vlone310/cfguardian/internal/domain/valueobjects"
	"github.com/vlone310/cfguardian/internal/ports/outbound"
)

// CheckPermissionRequest holds permission check data
type CheckPermissionRequest struct {
	UserID           string
	ProjectID        string
	RequiredRoleLevel string
}

// CheckPermissionResponse holds permission check result
type CheckPermissionResponse struct {
	Allowed       bool
	UserRoleLevel string
}

// CheckPermissionUseCase handles permission checking for authorization
type CheckPermissionUseCase struct {
	roleRepo outbound.RoleRepository
}

// NewCheckPermissionUseCase creates a new CheckPermissionUseCase
func NewCheckPermissionUseCase(roleRepo outbound.RoleRepository) *CheckPermissionUseCase {
	return &CheckPermissionUseCase{
		roleRepo: roleRepo,
	}
}

// Execute checks if a user has the required permission in a project
func (uc *CheckPermissionUseCase) Execute(ctx context.Context, req CheckPermissionRequest) (*CheckPermissionResponse, error) {
	// Validate input
	if req.UserID == "" {
		return nil, fmt.Errorf("user ID is required")
	}
	if req.ProjectID == "" {
		return nil, fmt.Errorf("project ID is required")
	}
	if req.RequiredRoleLevel == "" {
		return nil, fmt.Errorf("required role level is required")
	}
	
	// Validate required role level
	requiredLevel, err := valueobjects.NewRoleLevel(req.RequiredRoleLevel)
	if err != nil {
		return nil, fmt.Errorf("invalid required role level: %w", err)
	}
	
	// Get user's role in the project
	userRoleLevel, err := uc.roleRepo.GetUserRole(ctx, req.UserID, req.ProjectID)
	if err != nil {
		// User has no role in this project
		return &CheckPermissionResponse{
			Allowed:       false,
			UserRoleLevel: "",
		}, nil
	}
	
	// Convert repository role level to domain value object
	userLevel, err := valueobjects.NewRoleLevel(string(userRoleLevel))
	if err != nil {
		return nil, fmt.Errorf("invalid user role level from database: %w", err)
	}
	
	// Check if user's role includes the required level
	allowed := userLevel.IncludesLevel(requiredLevel)
	
	return &CheckPermissionResponse{
		Allowed:       allowed,
		UserRoleLevel: userLevel.String(),
	}, nil
}

// RequirePermission checks permission and returns error if not allowed
func (uc *CheckPermissionUseCase) RequirePermission(ctx context.Context, userID, projectID, requiredLevel string) error {
	resp, err := uc.Execute(ctx, CheckPermissionRequest{
		UserID:           userID,
		ProjectID:        projectID,
		RequiredRoleLevel: requiredLevel,
	})
	if err != nil {
		return err
	}
	
	if !resp.Allowed {
		return fmt.Errorf("permission denied: requires %s role", requiredLevel)
	}
	
	return nil
}

// RequireAdmin is a helper that requires admin role
func (uc *CheckPermissionUseCase) RequireAdmin(ctx context.Context, userID, projectID string) error {
	return uc.RequirePermission(ctx, userID, projectID, string(valueobjects.RoleLevelAdmin))
}

// RequireEditor is a helper that requires at least editor role
func (uc *CheckPermissionUseCase) RequireEditor(ctx context.Context, userID, projectID string) error {
	return uc.RequirePermission(ctx, userID, projectID, string(valueobjects.RoleLevelEditor))
}

// RequireViewer is a helper that requires at least viewer role
func (uc *CheckPermissionUseCase) RequireViewer(ctx context.Context, userID, projectID string) error {
	return uc.RequirePermission(ctx, userID, projectID, string(valueobjects.RoleLevelViewer))
}

