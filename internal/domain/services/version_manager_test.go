package services

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/vlone310/cfguardian/internal/domain/valueobjects"
)

func TestVersionManager_InitialVersion(t *testing.T) {
	manager := NewVersionManager()

	// Act
	version := manager.InitialVersion()

	// Assert
	assert.Equal(t, int64(1), version.Value(), "Initial version should be 1")
	assert.Equal(t, "1", version.String(), "Initial version string should be '1'")
}

func TestVersionManager_NextVersion(t *testing.T) {
	manager := NewVersionManager()

	tests := []struct {
		name           string
		currentVersion valueobjects.Version
		expectedNext   int64
	}{
		{
			name:           "increment from 1",
			currentVersion: valueobjects.MustNewVersion(1),
			expectedNext:   2,
		},
		{
			name:           "increment from 5",
			currentVersion: valueobjects.MustNewVersion(5),
			expectedNext:   6,
		},
		{
			name:           "increment from 100",
			currentVersion: valueobjects.MustNewVersion(100),
			expectedNext:   101,
		},
		{
			name:           "increment from 999",
			currentVersion: valueobjects.MustNewVersion(999),
			expectedNext:   1000,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Act
			nextVersion := manager.NextVersion(tt.currentVersion)

			// Assert
			assert.Equal(t, tt.expectedNext, nextVersion.Value())
			assert.True(t, nextVersion.IsGreaterThan(tt.currentVersion),
				"Next version should be greater than current")
		})
	}
}

func TestVersionManager_ValidateUpdate(t *testing.T) {
	manager := NewVersionManager()

	tests := []struct {
		name        string
		expected    valueobjects.Version
		current     valueobjects.Version
		key         string
		expectError bool
		errorType   string
	}{
		{
			name:        "matching versions - update allowed",
			expected:    valueobjects.MustNewVersion(5),
			current:     valueobjects.MustNewVersion(5),
			key:         "app.config",
			expectError: false,
		},
		{
			name:        "version mismatch - update denied",
			expected:    valueobjects.MustNewVersion(3),
			current:     valueobjects.MustNewVersion(5),
			key:         "app.config",
			expectError: true,
			errorType:   "VersionConflictError",
		},
		{
			name:        "expected older than current",
			expected:    valueobjects.MustNewVersion(1),
			current:     valueobjects.MustNewVersion(10),
			key:         "db.config",
			expectError: true,
			errorType:   "VersionConflictError",
		},
		{
			name:        "expected newer than current",
			expected:    valueobjects.MustNewVersion(10),
			current:     valueobjects.MustNewVersion(5),
			key:         "cache.config",
			expectError: true,
			errorType:   "VersionConflictError",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Act
			err := manager.ValidateUpdate(tt.expected, tt.current, tt.key)

			// Assert
			if tt.expectError {
				assert.Error(t, err)
				assert.True(t, IsVersionConflict(err),
					"Error should be a VersionConflictError")

				// Check error contains key
				assert.Contains(t, err.Error(), tt.key)
				assert.Contains(t, err.Error(), "conflict")
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestVersionManager_CanUpdate(t *testing.T) {
	manager := NewVersionManager()

	tests := []struct {
		name      string
		expected  valueobjects.Version
		current   valueobjects.Version
		canUpdate bool
	}{
		{
			name:      "matching versions",
			expected:  valueobjects.MustNewVersion(5),
			current:   valueobjects.MustNewVersion(5),
			canUpdate: true,
		},
		{
			name:      "expected older",
			expected:  valueobjects.MustNewVersion(3),
			current:   valueobjects.MustNewVersion(5),
			canUpdate: false,
		},
		{
			name:      "expected newer",
			expected:  valueobjects.MustNewVersion(7),
			current:   valueobjects.MustNewVersion(5),
			canUpdate: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Act
			result := manager.CanUpdate(tt.expected, tt.current)

			// Assert
			assert.Equal(t, tt.canUpdate, result)
		})
	}
}

func TestVersionManager_CompareVersions(t *testing.T) {
	manager := NewVersionManager()

	tests := []struct {
		name     string
		v1       valueobjects.Version
		v2       valueobjects.Version
		expected int
	}{
		{
			name:     "v1 < v2",
			v1:       valueobjects.MustNewVersion(3),
			v2:       valueobjects.MustNewVersion(5),
			expected: -1,
		},
		{
			name:     "v1 == v2",
			v1:       valueobjects.MustNewVersion(5),
			v2:       valueobjects.MustNewVersion(5),
			expected: 0,
		},
		{
			name:     "v1 > v2",
			v1:       valueobjects.MustNewVersion(10),
			v2:       valueobjects.MustNewVersion(7),
			expected: 1,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Act
			result := manager.CompareVersions(tt.v1, tt.v2)

			// Assert
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestVersionManager_IsNewer(t *testing.T) {
	manager := NewVersionManager()

	tests := []struct {
		name     string
		v1       valueobjects.Version
		v2       valueobjects.Version
		expected bool
	}{
		{
			name:     "v1 is newer",
			v1:       valueobjects.MustNewVersion(10),
			v2:       valueobjects.MustNewVersion(5),
			expected: true,
		},
		{
			name:     "v1 is same",
			v1:       valueobjects.MustNewVersion(5),
			v2:       valueobjects.MustNewVersion(5),
			expected: false,
		},
		{
			name:     "v1 is older",
			v1:       valueobjects.MustNewVersion(3),
			v2:       valueobjects.MustNewVersion(5),
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Act
			result := manager.IsNewer(tt.v1, tt.v2)

			// Assert
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestVersionManager_IsOlder(t *testing.T) {
	manager := NewVersionManager()

	tests := []struct {
		name     string
		v1       valueobjects.Version
		v2       valueobjects.Version
		expected bool
	}{
		{
			name:     "v1 is older",
			v1:       valueobjects.MustNewVersion(3),
			v2:       valueobjects.MustNewVersion(5),
			expected: true,
		},
		{
			name:     "v1 is same",
			v1:       valueobjects.MustNewVersion(5),
			v2:       valueobjects.MustNewVersion(5),
			expected: false,
		},
		{
			name:     "v1 is newer",
			v1:       valueobjects.MustNewVersion(10),
			v2:       valueobjects.MustNewVersion(5),
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Act
			result := manager.IsOlder(tt.v1, tt.v2)

			// Assert
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestVersionManager_IsSame(t *testing.T) {
	manager := NewVersionManager()

	tests := []struct {
		name     string
		v1       valueobjects.Version
		v2       valueobjects.Version
		expected bool
	}{
		{
			name:     "same version",
			v1:       valueobjects.MustNewVersion(5),
			v2:       valueobjects.MustNewVersion(5),
			expected: true,
		},
		{
			name:     "different versions",
			v1:       valueobjects.MustNewVersion(5),
			v2:       valueobjects.MustNewVersion(6),
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Act
			result := manager.IsSame(tt.v1, tt.v2)

			// Assert
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestVersionManager_CalculateVersionDelta(t *testing.T) {
	manager := NewVersionManager()

	tests := []struct {
		name     string
		from     valueobjects.Version
		to       valueobjects.Version
		expected int64
	}{
		{
			name:     "positive delta",
			from:     valueobjects.MustNewVersion(5),
			to:       valueobjects.MustNewVersion(10),
			expected: 5,
		},
		{
			name:     "negative delta",
			from:     valueobjects.MustNewVersion(10),
			to:       valueobjects.MustNewVersion(5),
			expected: -5,
		},
		{
			name:     "zero delta",
			from:     valueobjects.MustNewVersion(7),
			to:       valueobjects.MustNewVersion(7),
			expected: 0,
		},
		{
			name:     "large delta",
			from:     valueobjects.MustNewVersion(1),
			to:       valueobjects.MustNewVersion(1000),
			expected: 999,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Act
			delta := manager.CalculateVersionDelta(tt.from, tt.to)

			// Assert
			assert.Equal(t, tt.expected, delta)
		})
	}
}

func TestVersionManager_IsValidRollbackTarget(t *testing.T) {
	manager := NewVersionManager()

	tests := []struct {
		name     string
		current  valueobjects.Version
		target   valueobjects.Version
		expected bool
	}{
		{
			name:     "valid rollback - target is older",
			current:  valueobjects.MustNewVersion(10),
			target:   valueobjects.MustNewVersion(5),
			expected: true,
		},
		{
			name:     "valid rollback - target is previous version",
			current:  valueobjects.MustNewVersion(5),
			target:   valueobjects.MustNewVersion(4),
			expected: true,
		},
		{
			name:     "invalid rollback - target equals current",
			current:  valueobjects.MustNewVersion(5),
			target:   valueobjects.MustNewVersion(5),
			expected: false,
		},
		{
			name:     "invalid rollback - target is newer",
			current:  valueobjects.MustNewVersion(5),
			target:   valueobjects.MustNewVersion(10),
			expected: false,
		},
		{
			name:     "valid rollback to version 1",
			current:  valueobjects.MustNewVersion(5),
			target:   valueobjects.MustNewVersion(1),
			expected: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Act
			result := manager.IsValidRollbackTarget(tt.current, tt.target)

			// Assert
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestVersionConflictError(t *testing.T) {
	t.Run("error message format", func(t *testing.T) {
		// Arrange
		err := VersionConflictError{
			Expected: valueobjects.MustNewVersion(3),
			Current:  valueobjects.MustNewVersion(5),
			Key:      "app.database.config",
		}

		// Act
		errMsg := err.Error()

		// Assert
		assert.Contains(t, errMsg, "app.database.config")
		assert.Contains(t, errMsg, "3")
		assert.Contains(t, errMsg, "5")
		assert.Contains(t, errMsg, "conflict")
		assert.Contains(t, errMsg, "concurrent modification")
	})

	t.Run("IsVersionConflict returns true for VersionConflictError", func(t *testing.T) {
		// Arrange
		err := VersionConflictError{
			Expected: valueobjects.MustNewVersion(1),
			Current:  valueobjects.MustNewVersion(2),
			Key:      "test",
		}

		// Act & Assert
		assert.True(t, IsVersionConflict(err))
	})

	t.Run("IsVersionConflict returns false for other errors", func(t *testing.T) {
		// Arrange
		err := errors.New("some other error")

		// Act & Assert
		assert.False(t, IsVersionConflict(err))
	})
}

func TestVersionManager_OptimisticLockingScenario(t *testing.T) {
	// This test simulates a real-world optimistic locking scenario
	manager := NewVersionManager()

	t.Run("successful update with correct version", func(t *testing.T) {
		// Current config state in database
		currentVersion := valueobjects.MustNewVersion(5)

		// User wants to update with correct version
		userProvidedVersion := valueobjects.MustNewVersion(5)
		configKey := "api.rate_limit"

		// Validate update
		err := manager.ValidateUpdate(userProvidedVersion, currentVersion, configKey)
		require.NoError(t, err, "Update should be allowed with matching versions")

		// Calculate next version
		nextVersion := manager.NextVersion(currentVersion)
		assert.Equal(t, int64(6), nextVersion.Value())
	})

	t.Run("failed update due to concurrent modification", func(t *testing.T) {
		// Current config state in database (updated by another user)
		currentVersion := valueobjects.MustNewVersion(7)

		// User wants to update but has old version
		userProvidedVersion := valueobjects.MustNewVersion(5)
		configKey := "api.rate_limit"

		// Validate update - should fail
		err := manager.ValidateUpdate(userProvidedVersion, currentVersion, configKey)
		require.Error(t, err, "Update should fail with version mismatch")
		assert.True(t, IsVersionConflict(err))

		// User cannot update
		assert.False(t, manager.CanUpdate(userProvidedVersion, currentVersion))
	})

	t.Run("successful rollback to previous version", func(t *testing.T) {
		// Current version
		currentVersion := valueobjects.MustNewVersion(10)

		// User wants to rollback to version 7
		targetVersion := valueobjects.MustNewVersion(7)

		// Validate rollback
		assert.True(t, manager.IsValidRollbackTarget(currentVersion, targetVersion))

		// Calculate delta
		delta := manager.CalculateVersionDelta(targetVersion, currentVersion)
		assert.Equal(t, int64(3), delta, "Should rollback 3 versions")
	})

	t.Run("concurrent updates by multiple users", func(t *testing.T) {
		// Initial state
		currentVersion := valueobjects.MustNewVersion(5)

		// User A reads version 5
		userAVersion := currentVersion

		// User B reads version 5
		userBVersion := currentVersion

		// User A updates first (version check passes)
		err := manager.ValidateUpdate(userAVersion, currentVersion, "config.key")
		require.NoError(t, err)

		// Simulate successful update by User A
		currentVersion = manager.NextVersion(currentVersion) // Now version 6

		// User B tries to update (version check fails)
		err = manager.ValidateUpdate(userBVersion, currentVersion, "config.key")
		require.Error(t, err, "User B's update should fail")
		assert.True(t, IsVersionConflict(err))

		// User B must retry with new version
		assert.False(t, manager.CanUpdate(userBVersion, currentVersion))
		assert.True(t, manager.CanUpdate(currentVersion, currentVersion))
	})
}

func TestVersionManager_VersionProgression(t *testing.T) {
	manager := NewVersionManager()

	t.Run("version progression through multiple updates", func(t *testing.T) {
		// Start with initial version
		version := manager.InitialVersion()
		assert.Equal(t, int64(1), version.Value())

		// Simulate 10 updates
		for i := 2; i <= 10; i++ {
			version = manager.NextVersion(version)
			assert.Equal(t, int64(i), version.Value())
		}

		// Final version should be 10
		assert.Equal(t, int64(10), version.Value())
	})

	t.Run("version comparison across progression", func(t *testing.T) {
		v1 := valueobjects.MustNewVersion(1)
		v5 := valueobjects.MustNewVersion(5)
		v10 := valueobjects.MustNewVersion(10)

		// Compare
		assert.True(t, manager.IsOlder(v1, v5))
		assert.True(t, manager.IsOlder(v5, v10))
		assert.True(t, manager.IsOlder(v1, v10))

		assert.True(t, manager.IsNewer(v10, v5))
		assert.True(t, manager.IsNewer(v5, v1))
		assert.True(t, manager.IsNewer(v10, v1))
	})
}
