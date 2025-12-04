package valueobjects

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewVersion(t *testing.T) {
	tests := []struct {
		name    string
		value   int64
		wantErr bool
	}{
		{
			name:    "valid version 1",
			value:   1,
			wantErr: false,
		},
		{
			name:    "valid version 100",
			value:   100,
			wantErr: false,
		},
		{
			name:    "zero version (invalid)",
			value:   0,
			wantErr: true,
		},
		{
			name:    "negative version",
			value:   -1,
			wantErr: true,
		},
		{
			name:    "large negative version",
			value:   -999,
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			version, err := NewVersion(tt.value)

			if tt.wantErr {
				require.Error(t, err)
				assert.Contains(t, err.Error(), "version must be")
			} else {
				require.NoError(t, err)
				assert.Equal(t, tt.value, version.Value())
			}
		})
	}
}

func TestVersion_Increment(t *testing.T) {
	// Start at version 1 (initial version)
	v := InitialVersion()
	assert.Equal(t, int64(1), v.Value())
	
	// Increment to version 2
	v = v.Increment()
	assert.Equal(t, int64(2), v.Value())
	
	// Increment to version 3
	v = v.Increment()
	assert.Equal(t, int64(3), v.Value())
	
	// Multiple increments
	for i := 0; i < 10; i++ {
		v = v.Increment()
	}
	assert.Equal(t, int64(13), v.Value())
}

func TestVersion_Equals(t *testing.T) {
	v1, _ := NewVersion(5)
	v2, _ := NewVersion(5)
	v3, _ := NewVersion(10)
	
	// Same versions should be equal
	assert.True(t, v1.Equals(v2))
	assert.True(t, v2.Equals(v1))
	
	// Different versions should not be equal
	assert.False(t, v1.Equals(v3))
	assert.False(t, v3.Equals(v1))
}

func TestVersion_IsGreaterThan(t *testing.T) {
	v1, _ := NewVersion(1)
	v2, _ := NewVersion(2)
	v10, _ := NewVersion(10)
	
	// v2 is greater than v1
	assert.True(t, v2.IsGreaterThan(v1))
	assert.False(t, v1.IsGreaterThan(v2))
	
	// v10 is greater than v1
	assert.True(t, v10.IsGreaterThan(v1))
	assert.False(t, v1.IsGreaterThan(v10))
	
	// Same version is not greater
	assert.False(t, v1.IsGreaterThan(v1))
}

func TestVersion_Value(t *testing.T) {
	tests := []struct {
		name  string
		value int64
	}{
		{"version 1", 1},
		{"version 42", 42},
		{"version 999", 999},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			v, err := NewVersion(tt.value)
			require.NoError(t, err)
			assert.Equal(t, tt.value, v.Value())
		})
	}
}

func TestVersion_OptimisticLockingScenario(t *testing.T) {
	// Simulate optimistic locking scenario
	
	// Initial version
	currentVersion := InitialVersion()
	assert.Equal(t, int64(1), currentVersion.Value())
	
	// User A reads config at version 1
	userAVersion := currentVersion
	
	// User B reads config at version 1
	userBVersion := currentVersion
	
	// User A updates (increments to version 2)
	userANewVersion := userAVersion.Increment()
	assert.Equal(t, int64(2), userANewVersion.Value())
	
	// User B tries to update with old version (should fail check)
	assert.True(t, userBVersion.Equals(currentVersion))
	assert.False(t, userBVersion.Equals(userANewVersion))
	
	// User B should detect version conflict
	if !userBVersion.Equals(userANewVersion) {
		// Version conflict detected - this is expected behavior
		assert.True(t, userANewVersion.IsGreaterThan(userBVersion))
	}
}

