package services

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/vlone310/cfguardian/internal/domain/valueobjects"
)

func TestAPIKeyGenerator_Generate(t *testing.T) {
	generator := NewAPIKeyGenerator()

	t.Run("generates valid API key", func(t *testing.T) {
		// Act
		apiKey, err := generator.Generate()

		// Assert
		require.NoError(t, err)
		assert.NotEmpty(t, apiKey.String())

		// Check format
		keyStr := apiKey.String()
		assert.True(t, strings.HasPrefix(keyStr, valueobjects.APIKeyPrefix),
			"API key should start with prefix")
		assert.Equal(t, valueobjects.APIKeyLength, len(keyStr),
			"API key should have correct total length")
	})

	t.Run("generates unique keys", func(t *testing.T) {
		// Generate multiple keys
		numKeys := 100
		keys := make(map[string]bool, numKeys)

		for i := 0; i < numKeys; i++ {
			apiKey, err := generator.Generate()
			require.NoError(t, err)

			keyStr := apiKey.String()

			// Check for duplicates
			assert.False(t, keys[keyStr], "Generated duplicate API key")
			keys[keyStr] = true
		}

		// All keys should be unique
		assert.Equal(t, numKeys, len(keys), "All generated keys should be unique")
	})

	t.Run("generates cryptographically secure keys", func(t *testing.T) {
		// Generate two keys and verify they're different
		key1, err1 := generator.Generate()
		key2, err2 := generator.Generate()

		require.NoError(t, err1)
		require.NoError(t, err2)
		assert.NotEqual(t, key1.String(), key2.String(),
			"Consecutive keys should be different (cryptographically random)")
	})

	t.Run("generated keys are valid", func(t *testing.T) {
		// Generate key
		apiKey, err := generator.Generate()
		require.NoError(t, err)

		// Validate using the generator's Validate method
		err = generator.Validate(apiKey.String())
		assert.NoError(t, err, "Generated key should be valid")
	})

	t.Run("generated keys have correct structure", func(t *testing.T) {
		// Generate key
		apiKey, err := generator.Generate()
		require.NoError(t, err)

		keyStr := apiKey.String()

		// Check prefix
		assert.True(t, strings.HasPrefix(keyStr, "cfg_"),
			"Should have 'cfg_' prefix")

		// Check length after prefix
		randomPart := strings.TrimPrefix(keyStr, valueobjects.APIKeyPrefix)
		assert.Equal(t, valueobjects.APIKeyRandomPartLength, len(randomPart),
			"Random part should have correct length")

		// Check characters (base64 URL-safe: A-Za-z0-9_-)
		for _, char := range randomPart {
			valid := (char >= 'A' && char <= 'Z') ||
				(char >= 'a' && char <= 'z') ||
				(char >= '0' && char <= '9') ||
				char == '_' || char == '-'
			assert.True(t, valid, "Character '%c' is not valid base64 URL-safe", char)
		}
	})
}

func TestAPIKeyGenerator_MustGenerate(t *testing.T) {
	generator := NewAPIKeyGenerator()

	t.Run("successfully generates key", func(t *testing.T) {
		// Act
		apiKey := generator.MustGenerate()

		// Assert
		assert.NotEmpty(t, apiKey.String())
		assert.True(t, strings.HasPrefix(apiKey.String(), valueobjects.APIKeyPrefix))
	})

	t.Run("does not panic on success", func(t *testing.T) {
		// Should not panic
		assert.NotPanics(t, func() {
			_ = generator.MustGenerate()
		})
	})

	t.Run("generates valid keys consistently", func(t *testing.T) {
		// Generate multiple times
		for i := 0; i < 10; i++ {
			apiKey := generator.MustGenerate()

			// Each should be valid
			err := generator.Validate(apiKey.String())
			assert.NoError(t, err, "MustGenerate should produce valid keys")
		}
	})
}

func TestAPIKeyGenerator_Validate(t *testing.T) {
	generator := NewAPIKeyGenerator()

	tests := []struct {
		name        string
		key         string
		expectError bool
		errorMsg    string
	}{
		{
			name:        "valid API key",
			key:         "cfg_" + strings.Repeat("a", valueobjects.APIKeyRandomPartLength),
			expectError: false,
		},
		{
			name:        "empty key",
			key:         "",
			expectError: true,
			errorMsg:    "cannot be empty",
		},
		{
			name:        "missing prefix",
			key:         "abc_" + strings.Repeat("x", valueobjects.APIKeyRandomPartLength),
			expectError: true,
			errorMsg:    "must start with",
		},
		{
			name:        "too short",
			key:         "cfg_abc",
			expectError: true,
			errorMsg:    "must be exactly",
		},
		{
			name:        "too long",
			key:         "cfg_" + strings.Repeat("x", valueobjects.APIKeyRandomPartLength+10),
			expectError: true,
			errorMsg:    "must be exactly",
		},
		{
			name:        "no prefix",
			key:         strings.Repeat("x", valueobjects.APIKeyLength),
			expectError: true,
			errorMsg:    "must start with",
		},
		{
			name:        "invalid characters",
			key:         "cfg_" + strings.Repeat("@", valueobjects.APIKeyRandomPartLength),
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Act
			err := generator.Validate(tt.key)

			// Assert
			if tt.expectError {
				assert.Error(t, err)
				if tt.errorMsg != "" {
					assert.Contains(t, err.Error(), tt.errorMsg)
				}
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestAPIKeyGenerator_Integration(t *testing.T) {
	// This test simulates a real-world project creation scenario
	generator := NewAPIKeyGenerator()

	t.Run("project creation with API key", func(t *testing.T) {
		// Simulate creating multiple projects
		projectKeys := make(map[string]string)

		for i := 0; i < 10; i++ {
			// Generate API key for project
			apiKey, err := generator.Generate()
			require.NoError(t, err)

			projectID := string(rune('A' + i)) // A, B, C, etc.
			projectKeys[projectID] = apiKey.String()
		}

		// Verify all keys are unique
		uniqueKeys := make(map[string]bool)
		for _, key := range projectKeys {
			assert.False(t, uniqueKeys[key], "Should not have duplicate keys")
			uniqueKeys[key] = true
		}
		assert.Equal(t, len(projectKeys), len(uniqueKeys))

		// Verify all keys are valid
		for projectID, key := range projectKeys {
			err := generator.Validate(key)
			assert.NoError(t, err, "Key for project %s should be valid", projectID)
		}
	})

	t.Run("API key regeneration", func(t *testing.T) {
		// Generate original key
		originalKey, err := generator.Generate()
		require.NoError(t, err)

		// Regenerate (simulate key rotation)
		newKey, err := generator.Generate()
		require.NoError(t, err)

		// Keys should be different
		assert.NotEqual(t, originalKey.String(), newKey.String(),
			"Regenerated key should be different from original")

		// Both should be valid
		assert.NoError(t, generator.Validate(originalKey.String()))
		assert.NoError(t, generator.Validate(newKey.String()))
	})
}

func TestAPIKeyGenerator_Consistency(t *testing.T) {
	generator := NewAPIKeyGenerator()

	t.Run("consistent format across generations", func(t *testing.T) {
		// Generate multiple keys
		for i := 0; i < 50; i++ {
			apiKey, err := generator.Generate()
			require.NoError(t, err)

			keyStr := apiKey.String()

			// All should have same format
			assert.True(t, strings.HasPrefix(keyStr, "cfg_"))
			assert.Equal(t, valueobjects.APIKeyLength, len(keyStr))

			// Validation should pass
			err = generator.Validate(keyStr)
			assert.NoError(t, err)
		}
	})

	t.Run("no observable pattern in generation", func(t *testing.T) {
		// Generate keys and check there's no obvious pattern
		keys := make([]string, 20)
		for i := 0; i < 20; i++ {
			apiKey, err := generator.Generate()
			require.NoError(t, err)
			keys[i] = strings.TrimPrefix(apiKey.String(), valueobjects.APIKeyPrefix)
		}

		// Check that first character varies (not all same)
		firstChars := make(map[byte]bool)
		for _, key := range keys {
			firstChars[key[0]] = true
		}
		// Should have at least some variation (highly likely with 20 random keys)
		assert.Greater(t, len(firstChars), 1,
			"First characters should vary (indicates randomness)")

		// Check that keys are not sequential or incremental
		for i := 1; i < len(keys); i++ {
			assert.NotEqual(t, keys[i-1], keys[i],
				"Keys should not be sequential")
		}
	})
}

func TestAPIKeyGenerator_EdgeCases(t *testing.T) {
	generator := NewAPIKeyGenerator()

	t.Run("rapid generation", func(t *testing.T) {
		// Generate keys in rapid succession
		keys := make([]string, 1000)
		for i := 0; i < 1000; i++ {
			apiKey, err := generator.Generate()
			require.NoError(t, err)
			keys[i] = apiKey.String()
		}

		// Check all are unique
		uniqueKeys := make(map[string]bool)
		for _, key := range keys {
			assert.False(t, uniqueKeys[key], "Should not generate duplicates in rapid succession")
			uniqueKeys[key] = true
		}
		assert.Equal(t, 1000, len(uniqueKeys))
	})

	t.Run("concurrent generation", func(t *testing.T) {
		// Test thread-safety of generation
		numGoroutines := 10
		keysPerGoroutine := 10
		results := make(chan string, numGoroutines*keysPerGoroutine)

		// Launch goroutines
		for i := 0; i < numGoroutines; i++ {
			go func() {
				for j := 0; j < keysPerGoroutine; j++ {
					apiKey, err := generator.Generate()
					if err == nil {
						results <- apiKey.String()
					}
				}
			}()
		}

		// Collect results
		keys := make(map[string]bool)
		for i := 0; i < numGoroutines*keysPerGoroutine; i++ {
			key := <-results
			assert.False(t, keys[key], "Concurrent generation should not produce duplicates")
			keys[key] = true
		}

		// All keys should be unique
		assert.Equal(t, numGoroutines*keysPerGoroutine, len(keys))
	})
}
