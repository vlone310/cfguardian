package entities

import (
	"time"

	"github.com/vlone310/cfguardian/internal/domain/valueobjects"
)

// User represents a user entity with management UI access
type User struct {
	id           string
	email        valueobjects.Email
	passwordHash string
	createdAt    time.Time
	updatedAt    time.Time
}

// NewUser creates a new User entity
func NewUser(id string, email valueobjects.Email, passwordHash string) *User {
	now := time.Now()
	return &User{
		id:           id,
		email:        email,
		passwordHash: passwordHash,
		createdAt:    now,
		updatedAt:    now,
	}
}

// ReconstructUser reconstructs a User from persistence layer (e.g., database)
func ReconstructUser(id string, email valueobjects.Email, passwordHash string, createdAt, updatedAt time.Time) *User {
	return &User{
		id:           id,
		email:        email,
		passwordHash: passwordHash,
		createdAt:    createdAt,
		updatedAt:    updatedAt,
	}
}

// ID returns the user ID
func (u *User) ID() string {
	return u.id
}

// Email returns the user's email
func (u *User) Email() valueobjects.Email {
	return u.email
}

// PasswordHash returns the password hash
func (u *User) PasswordHash() string {
	return u.passwordHash
}

// CreatedAt returns the creation timestamp
func (u *User) CreatedAt() time.Time {
	return u.createdAt
}

// UpdatedAt returns the last update timestamp
func (u *User) UpdatedAt() time.Time {
	return u.updatedAt
}

// UpdateEmail updates the user's email
func (u *User) UpdateEmail(email valueobjects.Email) {
	u.email = email
	u.updatedAt = time.Now()
}

// UpdatePasswordHash updates the password hash
func (u *User) UpdatePasswordHash(passwordHash string) {
	u.passwordHash = passwordHash
	u.updatedAt = time.Now()
}

// Equals checks if two users are the same entity (by ID)
func (u *User) Equals(other *User) bool {
	if other == nil {
		return false
	}
	return u.id == other.id
}

