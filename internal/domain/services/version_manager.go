package services

import (
	"fmt"

	"github.com/vlone310/cfguardian/internal/domain/valueobjects"
)

// VersionConflictError represents an optimistic locking conflict
type VersionConflictError struct {
	Expected valueobjects.Version
	Current  valueobjects.Version
	Key      string
}

// Error implements the error interface
func (e VersionConflictError) Error() string {
	return fmt.Sprintf(
		"version conflict for key '%s': expected version %s, but current version is %s (concurrent modification detected)",
		e.Key,
		e.Expected.String(),
		e.Current.String(),
	)
}

// IsVersionConflict checks if an error is a version conflict error
func IsVersionConflict(err error) bool {
	_, ok := err.(VersionConflictError)
	return ok
}

// VersionManager manages version-related operations for optimistic locking
type VersionManager struct{}

// NewVersionManager creates a new VersionManager
func NewVersionManager() *VersionManager {
	return &VersionManager{}
}

// InitialVersion returns the initial version for new configs
func (vm *VersionManager) InitialVersion() valueobjects.Version {
	return valueobjects.InitialVersion()
}

// NextVersion calculates the next version
func (vm *VersionManager) NextVersion(current valueobjects.Version) valueobjects.Version {
	return current.Next()
}

// ValidateUpdate validates that an update can proceed based on version matching
func (vm *VersionManager) ValidateUpdate(expected, current valueobjects.Version, key string) error {
	if !expected.Equals(current) {
		return VersionConflictError{
			Expected: expected,
			Current:  current,
			Key:      key,
		}
	}
	return nil
}

// CanUpdate checks if an update can proceed (version matches)
func (vm *VersionManager) CanUpdate(expected, current valueobjects.Version) bool {
	return expected.Equals(current)
}

// CompareVersions compares two versions
// Returns: -1 if v1 < v2, 0 if v1 == v2, 1 if v1 > v2
func (vm *VersionManager) CompareVersions(v1, v2 valueobjects.Version) int {
	if v1.IsLessThan(v2) {
		return -1
	}
	if v1.IsGreaterThan(v2) {
		return 1
	}
	return 0
}

// IsNewer checks if version1 is newer than version2
func (vm *VersionManager) IsNewer(v1, v2 valueobjects.Version) bool {
	return v1.IsGreaterThan(v2)
}

// IsOlder checks if version1 is older than version2
func (vm *VersionManager) IsOlder(v1, v2 valueobjects.Version) bool {
	return v1.IsLessThan(v2)
}

// IsSame checks if two versions are the same
func (vm *VersionManager) IsSame(v1, v2 valueobjects.Version) bool {
	return v1.Equals(v2)
}

// CalculateVersionDelta calculates the difference between two versions
func (vm *VersionManager) CalculateVersionDelta(from, to valueobjects.Version) int64 {
	return to.Value() - from.Value()
}

// IsValidRollbackTarget checks if rolling back to targetVersion is valid
// (target version must be less than current version and >= 1)
func (vm *VersionManager) IsValidRollbackTarget(current, target valueobjects.Version) bool {
	return target.IsLessThan(current) && target.Value() >= 1
}

