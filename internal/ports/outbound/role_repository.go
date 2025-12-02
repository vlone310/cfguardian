package outbound

import (
	"context"
)

// RoleLevel represents the level of access a user has
type RoleLevel string

const (
	RoleLevelAdmin  RoleLevel = "admin"
	RoleLevelEditor RoleLevel = "editor"
	RoleLevelViewer RoleLevel = "viewer"
)

// Role represents a user's role within a project
type Role struct {
	UserID    string
	ProjectID string
	RoleLevel RoleLevel
	CreatedAt string
	UpdatedAt string
}

// UserRoleWithProject includes project name for convenient display
type UserRoleWithProject struct {
	Role
	ProjectName string
}

// ProjectRoleWithUser includes user email for convenient display
type ProjectRoleWithUser struct {
	Role
	UserEmail string
}

// AssignRoleParams holds parameters for assigning a role
type AssignRoleParams struct {
	UserID    string
	ProjectID string
	RoleLevel RoleLevel
}

// RoleRepository defines the interface for role data access
type RoleRepository interface {
	// Assign assigns or updates a role for a user in a project
	Assign(ctx context.Context, params AssignRoleParams) (*Role, error)
	
	// Get retrieves a specific role
	Get(ctx context.Context, userID, projectID string) (*Role, error)
	
	// GetUserRole retrieves just the role level for a user in a project
	GetUserRole(ctx context.Context, userID, projectID string) (RoleLevel, error)
	
	// ListUserRoles retrieves all roles for a specific user (with project names)
	ListUserRoles(ctx context.Context, userID string) ([]*UserRoleWithProject, error)
	
	// ListProjectRoles retrieves all roles for a specific project (with user emails)
	ListProjectRoles(ctx context.Context, projectID string) ([]*ProjectRoleWithUser, error)
	
	// ListByLevel retrieves all roles with a specific level
	ListByLevel(ctx context.Context, level RoleLevel) ([]*Role, error)
	
	// Update updates a role
	Update(ctx context.Context, userID, projectID string, level RoleLevel) (*Role, error)
	
	// Revoke revokes a specific role
	Revoke(ctx context.Context, userID, projectID string) error
	
	// RevokeAllUserRoles revokes all roles for a user
	RevokeAllUserRoles(ctx context.Context, userID string) error
	
	// RevokeAllProjectRoles revokes all roles for a project
	RevokeAllProjectRoles(ctx context.Context, projectID string) error
	
	// Exists checks if a role exists
	Exists(ctx context.Context, userID, projectID string) (bool, error)
	
	// HasRole checks if a user has a specific role in a project
	HasRole(ctx context.Context, userID, projectID string, level RoleLevel) (bool, error)
	
	// HasMinimumRole checks if a user has at least the specified role level
	// (e.g., if user is admin, they also satisfy editor and viewer checks)
	HasMinimumRole(ctx context.Context, userID, projectID string, minLevel RoleLevel) (bool, error)
	
	// CountByProject returns the number of roles in a project
	CountByProject(ctx context.Context, projectID string) (int64, error)
	
	// CountByUser returns the number of roles for a user
	CountByUser(ctx context.Context, userID string) (int64, error)
}

