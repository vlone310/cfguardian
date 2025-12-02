package outbound

import (
	"context"
)

// User represents a user entity (simplified for repository interface)
type User struct {
	ID           string
	Email        string
	PasswordHash string
	CreatedAt    string
	UpdatedAt    string
}

// CreateUserParams holds parameters for creating a user
type CreateUserParams struct {
	ID           string
	Email        string
	PasswordHash string
}

// UpdateUserParams holds parameters for updating a user
type UpdateUserParams struct {
	ID           string
	Email        *string
	PasswordHash *string
}

// UserRepository defines the interface for user data access
type UserRepository interface {
	// Create creates a new user
	Create(ctx context.Context, params CreateUserParams) (*User, error)
	
	// GetByID retrieves a user by ID
	GetByID(ctx context.Context, id string) (*User, error)
	
	// GetByEmail retrieves a user by email
	GetByEmail(ctx context.Context, email string) (*User, error)
	
	// List retrieves all users
	List(ctx context.Context) ([]*User, error)
	
	// Update updates a user
	Update(ctx context.Context, params UpdateUserParams) (*User, error)
	
	// Delete deletes a user
	Delete(ctx context.Context, id string) error
	
	// Exists checks if a user exists by ID
	Exists(ctx context.Context, id string) (bool, error)
	
	// ExistsByEmail checks if a user exists by email
	ExistsByEmail(ctx context.Context, email string) (bool, error)
	
	// Count returns the total number of users
	Count(ctx context.Context) (int64, error)
}

