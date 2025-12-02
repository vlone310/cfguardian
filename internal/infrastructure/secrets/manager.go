package secrets

import (
	"crypto/rand"
	"encoding/base64"
	"fmt"
	"os"
	"strings"
)

// Manager handles secret management and rotation
type Manager struct {
	secrets map[string]string
}

// NewManager creates a new secrets manager
func NewManager() *Manager {
	return &Manager{
		secrets: make(map[string]string),
	}
}

// LoadFromEnv loads secrets from environment variables
func (m *Manager) LoadFromEnv(keys ...string) error {
	for _, key := range keys {
		value := os.Getenv(key)
		if value == "" {
			return fmt.Errorf("required secret %s not found in environment", key)
		}
		m.secrets[key] = value
	}
	return nil
}

// Get retrieves a secret by key
func (m *Manager) Get(key string) (string, error) {
	value, exists := m.secrets[key]
	if !exists {
		return "", fmt.Errorf("secret %s not found", key)
	}
	return value, nil
}

// MustGet retrieves a secret or panics
func (m *Manager) MustGet(key string) string {
	value, err := m.Get(key)
	if err != nil {
		panic(err)
	}
	return value
}

// Set sets a secret value
func (m *Manager) Set(key, value string) {
	m.secrets[key] = value
}

// GenerateSecret generates a random secret of specified length
func (m *Manager) GenerateSecret(length int) (string, error) {
	bytes := make([]byte, length)
	if _, err := rand.Read(bytes); err != nil {
		return "", fmt.Errorf("failed to generate secret: %w", err)
	}
	return base64.URLEncoding.EncodeToString(bytes)[:length], nil
}

// RotateSecret generates a new secret and updates the key
func (m *Manager) RotateSecret(key string, length int) (string, error) {
	newSecret, err := m.GenerateSecret(length)
	if err != nil {
		return "", err
	}
	m.Set(key, newSecret)
	return newSecret, nil
}

// Mask returns a masked version of a secret for logging
// Shows only first and last few characters
func Mask(secret string) string {
	if len(secret) <= 8 {
		return "***"
	}
	return secret[:4] + "..." + secret[len(secret)-4:]
}

// MaskEmail returns a masked version of an email for logging
func MaskEmail(email string) string {
	parts := strings.Split(email, "@")
	if len(parts) != 2 {
		return "***@***"
	}
	
	local := parts[0]
	domain := parts[1]
	
	if len(local) <= 2 {
		return "**@" + domain
	}
	
	return local[:1] + "***@" + domain
}

// ValidateSecretStrength checks if a secret meets minimum requirements
func ValidateSecretStrength(secret string, minLength int) error {
	if len(secret) < minLength {
		return fmt.Errorf("secret must be at least %d characters", minLength)
	}
	
	// Check for entropy (basic check)
	uniqueChars := make(map[rune]bool)
	for _, char := range secret {
		uniqueChars[char] = true
	}
	
	if len(uniqueChars) < minLength/2 {
		return fmt.Errorf("secret has insufficient entropy")
	}
	
	return nil
}

// RedactSensitiveFields removes sensitive fields from structured data
// Use this before logging user data
func RedactSensitiveFields(data map[string]interface{}) map[string]interface{} {
	sensitiveKeys := []string{
		"password",
		"password_hash",
		"token",
		"access_token",
		"refresh_token",
		"api_key",
		"secret",
		"jwt_secret",
		"auth_token",
	}
	
	redacted := make(map[string]interface{})
	for k, v := range data {
		// Check if key is sensitive
		isSensitive := false
		lowerKey := strings.ToLower(k)
		for _, sensitiveKey := range sensitiveKeys {
			if strings.Contains(lowerKey, sensitiveKey) {
				isSensitive = true
				break
			}
		}
		
		if isSensitive {
			redacted[k] = "[REDACTED]"
		} else {
			redacted[k] = v
		}
	}
	
	return redacted
}

