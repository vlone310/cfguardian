package role

import (
	"context"
	"fmt"

	"github.com/vlone310/cfguardian/internal/domain/entities"
	"github.com/vlone310/cfguardian/internal/domain/valueobjects"
	"github.com/vlone310/cfguardian/internal/ports/outbound"
)

// AssignRoleRequest holds role assignment data
type AssignRoleRequest struct {
	UserID    string `json:"user_id"`
	ProjectID string `json:"project_id"`
	RoleLevel string `json:"role_level"`
}

// AssignRoleResponse holds assigned role data
type AssignRoleResponse struct {
	UserID    string `json:"user_id"`
	ProjectID string `json:"project_id"`
	RoleLevel string `json:"role_level"`
	CreatedAt string `json:"created_at"`
	UpdatedAt string `json:"updated_at"`
}

// AssignRoleUseCase handles role assignment
type AssignRoleUseCase struct {
	roleRepo    outbound.RoleRepository
	userRepo    outbound.UserRepository
	projectRepo outbound.ProjectRepository
}

// NewAssignRoleUseCase creates a new AssignRoleUseCase
func NewAssignRoleUseCase(
	roleRepo outbound.RoleRepository,
	userRepo outbound.UserRepository,
	projectRepo outbound.ProjectRepository,
) *AssignRoleUseCase {
	return &AssignRoleUseCase{
		roleRepo:    roleRepo,
		userRepo:    userRepo,
		projectRepo: projectRepo,
	}
}

// Execute assigns or updates a role for a user in a project
func (uc *AssignRoleUseCase) Execute(ctx context.Context, req AssignRoleRequest) (*AssignRoleResponse, error) {
	// Validate input
	if req.UserID == "" {
		return nil, fmt.Errorf("user ID is required")
	}
	if req.ProjectID == "" {
		return nil, fmt.Errorf("project ID is required")
	}
	if req.RoleLevel == "" {
		return nil, fmt.Errorf("role level is required")
	}
	
	// Validate role level
	roleLevel, err := valueobjects.NewRoleLevel(req.RoleLevel)
	if err != nil {
		return nil, fmt.Errorf("invalid role level: %w", err)
	}
	
	// Verify user exists
	userExists, err := uc.userRepo.Exists(ctx, req.UserID)
	if err != nil {
		return nil, fmt.Errorf("failed to verify user exists: %w", err)
	}
	if !userExists {
		return nil, fmt.Errorf("user not found")
	}
	
	// Verify project exists
	projectExists, err := uc.projectRepo.Exists(ctx, req.ProjectID)
	if err != nil {
		return nil, fmt.Errorf("failed to verify project exists: %w", err)
	}
	if !projectExists {
		return nil, fmt.Errorf("project not found")
	}
	
	// Create domain entity
	roleEntity := entities.NewRole(req.UserID, req.ProjectID, roleLevel)
	
	// Assign role (upsert - creates or updates)
	role, err := uc.roleRepo.Assign(ctx, outbound.AssignRoleParams{
		UserID:    roleEntity.UserID(),
		ProjectID: roleEntity.ProjectID(),
		RoleLevel: outbound.RoleLevel(roleEntity.Level()),
	})
	if err != nil {
		return nil, fmt.Errorf("failed to assign role: %w", err)
	}
	
	return &AssignRoleResponse{
		UserID:    role.UserID,
		ProjectID: role.ProjectID,
		RoleLevel: string(role.RoleLevel),
		CreatedAt: role.CreatedAt,
		UpdatedAt: role.UpdatedAt,
	}, nil
}

