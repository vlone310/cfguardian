package valueobjects

import (
	"fmt"
	"regexp"
	"strings"
)

const (
	// APIKeyPrefix is the standard prefix for API keys
	APIKeyPrefix = "cfg_"
	
	// APIKeyLength is the total length of the API key (including prefix)
	APIKeyLength = 36 // cfg_ (4) + 32 random chars
	
	// APIKeyRandomPartLength is the length of the random part
	APIKeyRandomPartLength = 32
)

// apiKeyRegex validates the API key format
var apiKeyRegex = regexp.MustCompile(`^cfg_[a-zA-Z0-9]{32}$`)

// APIKey represents a validated API key for client access
type APIKey struct {
	value string
}

// NewAPIKey creates a new APIKey with validation
func NewAPIKey(key string) (APIKey, error) {
	// Trim spaces
	trimmed := strings.TrimSpace(key)
	
	// Validate
	if trimmed == "" {
		return APIKey{}, fmt.Errorf("api key cannot be empty")
	}
	
	if !strings.HasPrefix(trimmed, APIKeyPrefix) {
		return APIKey{}, fmt.Errorf("api key must start with '%s'", APIKeyPrefix)
	}
	
	if len(trimmed) != APIKeyLength {
		return APIKey{}, fmt.Errorf("api key must be exactly %d characters", APIKeyLength)
	}
	
	if !apiKeyRegex.MatchString(trimmed) {
		return APIKey{}, fmt.Errorf("invalid api key format")
	}
	
	return APIKey{value: trimmed}, nil
}

// MustNewAPIKey creates a new APIKey or panics if invalid
func MustNewAPIKey(key string) APIKey {
	ak, err := NewAPIKey(key)
	if err != nil {
		panic(fmt.Sprintf("invalid api key: %v", err))
	}
	return ak
}

// Value returns the API key string value
func (a APIKey) Value() string {
	return a.value
}

// String implements the Stringer interface
func (a APIKey) String() string {
	return a.value
}

// Equals checks if two API keys are equal
func (a APIKey) Equals(other APIKey) bool {
	return a.value == other.value
}

// IsEmpty checks if the API key is empty
func (a APIKey) IsEmpty() bool {
	return a.value == ""
}

// Masked returns a masked version of the API key for logging
// Shows only the prefix and last 4 characters
func (a APIKey) Masked() string {
	if len(a.value) < 8 {
		return "***"
	}
	return fmt.Sprintf("%s...%s", APIKeyPrefix, a.value[len(a.value)-4:])
}

// HasPrefix checks if the API key has the correct prefix
func (a APIKey) HasPrefix() bool {
	return strings.HasPrefix(a.value, APIKeyPrefix)
}

