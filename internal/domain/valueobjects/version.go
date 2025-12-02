package valueobjects

import (
	"fmt"
)

// Version represents an optimistic locking version number
type Version struct {
	value int64
}

// NewVersion creates a new Version with validation
func NewVersion(v int64) (Version, error) {
	if v < 1 {
		return Version{}, fmt.Errorf("version must be >= 1, got: %d", v)
	}
	return Version{value: v}, nil
}

// MustNewVersion creates a new Version or panics if invalid
func MustNewVersion(v int64) Version {
	ver, err := NewVersion(v)
	if err != nil {
		panic(fmt.Sprintf("invalid version: %v", err))
	}
	return ver
}

// InitialVersion returns the initial version (1) for new configs
func InitialVersion() Version {
	return Version{value: 1}
}

// Value returns the version number
func (v Version) Value() int64 {
	return v.value
}

// Next returns the next version number
func (v Version) Next() Version {
	return Version{value: v.value + 1}
}

// String implements the Stringer interface
func (v Version) String() string {
	return fmt.Sprintf("%d", v.value)
}

// Equals checks if two versions are equal
func (v Version) Equals(other Version) bool {
	return v.value == other.value
}

// IsGreaterThan checks if this version is greater than another
func (v Version) IsGreaterThan(other Version) bool {
	return v.value > other.value
}

// IsLessThan checks if this version is less than another
func (v Version) IsLessThan(other Version) bool {
	return v.value < other.value
}

// Increment returns a new Version incremented by 1
// Alias for Next() for clarity
func (v Version) Increment() Version {
	return v.Next()
}

// IsInitial checks if this is the initial version
func (v Version) IsInitial() bool {
	return v.value == 1
}

