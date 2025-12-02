package entities

import (
	"time"

	"github.com/vlone310/cfguardian/internal/domain/valueobjects"
)

// Role represents a user's access level within a project
type Role struct {
	userID    string
	projectID string
	level     valueobjects.RoleLevel
	createdAt time.Time
	updatedAt time.Time
}

// NewRole creates a new Role entity
func NewRole(userID, projectID string, level valueobjects.RoleLevel) *Role {
	now := time.Now()
	return &Role{
		userID:    userID,
		projectID: projectID,
		level:     level,
		createdAt: now,
		updatedAt: now,
	}
}

// ReconstructRole reconstructs a Role from persistence layer
func ReconstructRole(userID, projectID string, level valueobjects.RoleLevel, createdAt, updatedAt time.Time) *Role {
	return &Role{
		userID:    userID,
		projectID: projectID,
		level:     level,
		createdAt: createdAt,
		updatedAt: updatedAt,
	}
}

// UserID returns the user ID
func (r *Role) UserID() string {
	return r.userID
}

// ProjectID returns the project ID
func (r *Role) ProjectID() string {
	return r.projectID
}

// Level returns the role level
func (r *Role) Level() valueobjects.RoleLevel {
	return r.level
}

// CreatedAt returns the creation timestamp
func (r *Role) CreatedAt() time.Time {
	return r.createdAt
}

// UpdatedAt returns the last update timestamp
func (r *Role) UpdatedAt() time.Time {
	return r.updatedAt
}

// UpdateLevel updates the role level
func (r *Role) UpdateLevel(level valueobjects.RoleLevel) {
	r.level = level
	r.updatedAt = time.Now()
}

// CanRead checks if this role can read configs
func (r *Role) CanRead() bool {
	return r.level.CanRead()
}

// CanWrite checks if this role can write configs
func (r *Role) CanWrite() bool {
	return r.level.CanWrite()
}

// CanAdmin checks if this role can perform admin operations
func (r *Role) CanAdmin() bool {
	return r.level.CanAdmin()
}

// HasPermissionFor checks if this role has permission for a specific role level
func (r *Role) HasPermissionFor(requiredLevel valueobjects.RoleLevel) bool {
	return r.level.IncludesLevel(requiredLevel)
}

// IsAdmin checks if this is an admin role
func (r *Role) IsAdmin() bool {
	return r.level == valueobjects.RoleLevelAdmin
}

// IsEditor checks if this is an editor role
func (r *Role) IsEditor() bool {
	return r.level == valueobjects.RoleLevelEditor
}

// IsViewer checks if this is a viewer role
func (r *Role) IsViewer() bool {
	return r.level == valueobjects.RoleLevelViewer
}

// Equals checks if two roles are the same entity (by user ID and project ID)
func (r *Role) Equals(other *Role) bool {
	if other == nil {
		return false
	}
	return r.userID == other.userID && r.projectID == other.projectID
}

