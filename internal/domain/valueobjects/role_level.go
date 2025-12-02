package valueobjects

import (
	"fmt"
	"strings"
)

// RoleLevel represents the access level of a user in a project
type RoleLevel string

const (
	// RoleLevelAdmin has full control over all resources
	RoleLevelAdmin RoleLevel = "admin"
	
	// RoleLevelEditor can read and write configs
	RoleLevelEditor RoleLevel = "editor"
	
	// RoleLevelViewer can only read configs
	RoleLevelViewer RoleLevel = "viewer"
)

// AllRoleLevels returns all valid role levels
func AllRoleLevels() []RoleLevel {
	return []RoleLevel{RoleLevelAdmin, RoleLevelEditor, RoleLevelViewer}
}

// NewRoleLevel creates a new RoleLevel with validation
func NewRoleLevel(level string) (RoleLevel, error) {
	normalized := strings.ToLower(strings.TrimSpace(level))
	
	switch normalized {
	case string(RoleLevelAdmin):
		return RoleLevelAdmin, nil
	case string(RoleLevelEditor):
		return RoleLevelEditor, nil
	case string(RoleLevelViewer):
		return RoleLevelViewer, nil
	default:
		return "", fmt.Errorf("invalid role level: %s (must be admin, editor, or viewer)", level)
	}
}

// MustNewRoleLevel creates a new RoleLevel or panics if invalid
func MustNewRoleLevel(level string) RoleLevel {
	rl, err := NewRoleLevel(level)
	if err != nil {
		panic(fmt.Sprintf("invalid role level: %v", err))
	}
	return rl
}

// String returns the string representation
func (r RoleLevel) String() string {
	return string(r)
}

// IsValid checks if the role level is valid
func (r RoleLevel) IsValid() bool {
	switch r {
	case RoleLevelAdmin, RoleLevelEditor, RoleLevelViewer:
		return true
	default:
		return false
	}
}

// CanRead checks if this role level can read configs
func (r RoleLevel) CanRead() bool {
	return r.IsValid() // All valid roles can read
}

// CanWrite checks if this role level can write configs
func (r RoleLevel) CanWrite() bool {
	return r == RoleLevelAdmin || r == RoleLevelEditor
}

// CanAdmin checks if this role level can perform admin operations
func (r RoleLevel) CanAdmin() bool {
	return r == RoleLevelAdmin
}

// IncludesLevel checks if this role includes another level
// Admin includes Editor and Viewer
// Editor includes Viewer
func (r RoleLevel) IncludesLevel(other RoleLevel) bool {
	if r == other {
		return true
	}
	
	switch r {
	case RoleLevelAdmin:
		// Admin includes all levels
		return true
	case RoleLevelEditor:
		// Editor includes viewer
		return other == RoleLevelViewer
	case RoleLevelViewer:
		// Viewer includes no other levels
		return false
	default:
		return false
	}
}

// Priority returns the priority level (higher = more privileges)
// Admin: 3, Editor: 2, Viewer: 1
func (r RoleLevel) Priority() int {
	switch r {
	case RoleLevelAdmin:
		return 3
	case RoleLevelEditor:
		return 2
	case RoleLevelViewer:
		return 1
	default:
		return 0
	}
}

// IsGreaterThan checks if this role has higher privileges than another
func (r RoleLevel) IsGreaterThan(other RoleLevel) bool {
	return r.Priority() > other.Priority()
}

// IsGreaterThanOrEqual checks if this role has equal or higher privileges
func (r RoleLevel) IsGreaterThanOrEqual(other RoleLevel) bool {
	return r.Priority() >= other.Priority()
}

