package valueobjects

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewAPIKey(t *testing.T) {
	tests := []struct {
		name    string
		value   string
		wantErr bool
		errMsg  string
	}{
		{
			name:    "valid API key with prefix",
			value:   "cfg_" + strings.Repeat("a", 32),
			wantErr: false,
		},
		{
			name:    "valid API key with mixed characters",
			value:   "cfg_" + strings.Repeat("a", 32),
			wantErr: false,
		},
		{
			name:    "empty API key",
			value:   "",
			wantErr: true,
			errMsg:  "API key cannot be empty",
		},
		{
			name:    "missing prefix",
			value:   strings.Repeat("a", 32),
			wantErr: true,
			errMsg:  "API key must start with",
		},
		{
			name:    "wrong prefix",
			value:   "wrong_" + strings.Repeat("a", 32),
			wantErr: true,
			errMsg:  "API key must start with",
		},
		{
			name:    "too short",
			value:   "cfg_short",
			wantErr: true,
			errMsg:  "must be exactly",
		},
		{
			name:    "too long",
			value:   "cfg_" + strings.Repeat("a", 50),
			wantErr: true,
			errMsg:  "must be exactly",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			apiKey, err := NewAPIKey(tt.value)

			if tt.wantErr {
				require.Error(t, err)
				// Make error message comparison case-insensitive
				errLower := strings.ToLower(err.Error())
				msgLower := strings.ToLower(tt.errMsg)
				assert.Contains(t, errLower, msgLower)
			} else {
				require.NoError(t, err)
				assert.Equal(t, tt.value, apiKey.String())
			}
		})
	}
}

// Note: GenerateAPIKey is in domain services, not here
// We'll test it in api_key_generator_test.go

func TestAPIKey_String(t *testing.T) {
	value := "cfg_" + strings.Repeat("a", 32)
	apiKey, err := NewAPIKey(value)
	require.NoError(t, err)
	
	assert.Equal(t, value, apiKey.String())
}

func TestAPIKey_Masked(t *testing.T) {
	value := "cfg_" + strings.Repeat("a", 32)
	apiKey, err := NewAPIKey(value)
	require.NoError(t, err)
	
	masked := apiKey.Masked()
	
	// Should show prefix and last 4 characters
	assert.True(t, strings.HasPrefix(masked, "cfg_"))
	assert.True(t, strings.Contains(masked, "..."))
	
	// Should be shorter than original
	assert.Less(t, len(masked), len(value))
	
	// Should not reveal full key
	assert.NotEqual(t, value, masked)
}

func TestAPIKey_Equals(t *testing.T) {
	key1, _ := NewAPIKey("cfg_" + strings.Repeat("a", 32))
	key2, _ := NewAPIKey("cfg_" + strings.Repeat("a", 32))
	key3, _ := NewAPIKey("cfg_" + strings.Repeat("b", 32))
	
	// Same keys should be equal
	assert.True(t, key1.Equals(key2))
	
	// Different keys should not be equal
	assert.False(t, key1.Equals(key3))
	
	// Empty key should not equal valid key
	emptyKey := APIKey{}
	assert.False(t, key1.Equals(emptyKey))
}

