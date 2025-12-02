package services

import (
	"crypto/rand"
	"encoding/base64"
	"fmt"

	"github.com/vlone310/cfguardian/internal/domain/valueobjects"
)

// APIKeyGenerator generates secure API keys
type APIKeyGenerator struct{}

// NewAPIKeyGenerator creates a new APIKeyGenerator
func NewAPIKeyGenerator() *APIKeyGenerator {
	return &APIKeyGenerator{}
}

// Generate generates a new cryptographically secure API key
func (akg *APIKeyGenerator) Generate() (valueobjects.APIKey, error) {
	// Generate random bytes
	randomBytes := make([]byte, 24) // 24 bytes = 32 chars in base64
	if _, err := rand.Read(randomBytes); err != nil {
		return valueobjects.APIKey{}, fmt.Errorf("failed to generate random bytes: %w", err)
	}
	
	// Encode to base64 URL-safe (without padding)
	randomPart := base64.RawURLEncoding.EncodeToString(randomBytes)
	
	// Ensure we have exactly 32 characters
	if len(randomPart) > valueobjects.APIKeyRandomPartLength {
		randomPart = randomPart[:valueobjects.APIKeyRandomPartLength]
	}
	
	// Construct API key with prefix
	apiKeyString := valueobjects.APIKeyPrefix + randomPart
	
	// Validate and return
	apiKey, err := valueobjects.NewAPIKey(apiKeyString)
	if err != nil {
		return valueobjects.APIKey{}, fmt.Errorf("failed to create api key: %w", err)
	}
	
	return apiKey, nil
}

// MustGenerate generates a new API key or panics
// Use only in situations where failure is unacceptable
func (akg *APIKeyGenerator) MustGenerate() valueobjects.APIKey {
	apiKey, err := akg.Generate()
	if err != nil {
		panic(fmt.Sprintf("failed to generate api key: %v", err))
	}
	return apiKey
}

// Validate validates an API key format
func (akg *APIKeyGenerator) Validate(key string) error {
	_, err := valueobjects.NewAPIKey(key)
	return err
}

