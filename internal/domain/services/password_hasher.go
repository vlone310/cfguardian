package services

import (
	"fmt"

	"golang.org/x/crypto/bcrypt"
)

// PasswordHasher handles password hashing and verification
type PasswordHasher struct {
	cost int
}

// NewPasswordHasher creates a new PasswordHasher
// cost should be between bcrypt.MinCost (4) and bcrypt.MaxCost (31)
// Recommended: 10-12 for production
func NewPasswordHasher(cost int) *PasswordHasher {
	if cost < bcrypt.MinCost {
		cost = bcrypt.DefaultCost
	}
	if cost > bcrypt.MaxCost {
		cost = bcrypt.MaxCost
	}
	
	return &PasswordHasher{
		cost: cost,
	}
}

// Hash hashes a password using bcrypt
func (ph *PasswordHasher) Hash(password string) (string, error) {
	if password == "" {
		return "", fmt.Errorf("password cannot be empty")
	}
	
	if len(password) < 8 {
		return "", fmt.Errorf("password must be at least 8 characters")
	}
	
	if len(password) > 72 {
		return "", fmt.Errorf("password too long (max 72 characters for bcrypt)")
	}
	
	hash, err := bcrypt.GenerateFromPassword([]byte(password), ph.cost)
	if err != nil {
		return "", fmt.Errorf("failed to hash password: %w", err)
	}
	
	return string(hash), nil
}

// Verify verifies a password against a hash
func (ph *PasswordHasher) Verify(password, hash string) error {
	if password == "" {
		return fmt.Errorf("password cannot be empty")
	}
	
	if hash == "" {
		return fmt.Errorf("hash cannot be empty")
	}
	
	if err := bcrypt.CompareHashAndPassword([]byte(hash), []byte(password)); err != nil {
		if err == bcrypt.ErrMismatchedHashAndPassword {
			return fmt.Errorf("invalid password")
		}
		return fmt.Errorf("failed to verify password: %w", err)
	}
	
	return nil
}

// NeedsRehash checks if a password hash needs to be rehashed
// This is useful when changing the cost factor
func (ph *PasswordHasher) NeedsRehash(hash string) bool {
	cost, err := bcrypt.Cost([]byte(hash))
	if err != nil {
		return true
	}
	return cost != ph.cost
}

