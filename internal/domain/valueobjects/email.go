package valueobjects

import (
	"fmt"
	"regexp"
	"strings"
)

// emailRegex is a simplified email validation regex
var emailRegex = regexp.MustCompile(`^[a-zA-Z0-9._%+\-]+@[a-zA-Z0-9.\-]+\.[a-zA-Z]{2,}$`)

// Email represents a validated email address
type Email struct {
	value string
}

// NewEmail creates a new Email value object with validation
func NewEmail(email string) (Email, error) {
	// Normalize: trim spaces and lowercase
	normalized := strings.ToLower(strings.TrimSpace(email))
	
	// Validate
	if normalized == "" {
		return Email{}, fmt.Errorf("email cannot be empty")
	}
	
	if len(normalized) > 255 {
		return Email{}, fmt.Errorf("email too long (max 255 characters)")
	}
	
	if !emailRegex.MatchString(normalized) {
		return Email{}, fmt.Errorf("invalid email format: %s", email)
	}
	
	return Email{value: normalized}, nil
}

// MustNewEmail creates a new Email or panics if invalid
// Use only when you're certain the email is valid (e.g., from database)
func MustNewEmail(email string) Email {
	e, err := NewEmail(email)
	if err != nil {
		panic(fmt.Sprintf("invalid email: %v", err))
	}
	return e
}

// Value returns the email string value
func (e Email) Value() string {
	return e.value
}

// String implements the Stringer interface
func (e Email) String() string {
	return e.value
}

// Equals checks if two emails are equal
func (e Email) Equals(other Email) bool {
	return e.value == other.value
}

// IsEmpty checks if the email is empty
func (e Email) IsEmpty() bool {
	return e.value == ""
}

// Domain returns the domain part of the email (after @)
func (e Email) Domain() string {
	parts := strings.Split(e.value, "@")
	if len(parts) == 2 {
		return parts[1]
	}
	return ""
}

// LocalPart returns the local part of the email (before @)
func (e Email) LocalPart() string {
	parts := strings.Split(e.value, "@")
	if len(parts) == 2 {
		return parts[0]
	}
	return ""
}

