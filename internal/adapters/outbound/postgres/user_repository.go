package postgres

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/vlone310/cfguardian/internal/adapters/outbound/postgres/sqlc"
	"github.com/vlone310/cfguardian/internal/ports/outbound"
)

// UserRepositoryAdapter implements outbound.UserRepository using PostgreSQL
type UserRepositoryAdapter struct {
	pool    *pgxpool.Pool
	queries *sqlc.Queries
}

// NewUserRepositoryAdapter creates a new PostgreSQL user repository
func NewUserRepositoryAdapter(pool *pgxpool.Pool) *UserRepositoryAdapter {
	return &UserRepositoryAdapter{
		pool:    pool,
		queries: sqlc.New(pool),
	}
}

// Create creates a new user
func (r *UserRepositoryAdapter) Create(ctx context.Context, params outbound.CreateUserParams) (*outbound.User, error) {
	user, err := r.queries.CreateUser(ctx, params.ID, params.Email, params.PasswordHash)
	if err != nil {
		return nil, fmt.Errorf("failed to create user: %w", err)
	}
	
	return r.modelToOutbound(&user), nil
}

// GetByID retrieves a user by ID
func (r *UserRepositoryAdapter) GetByID(ctx context.Context, id string) (*outbound.User, error) {
	user, err := r.queries.GetUserByID(ctx, id)
	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, fmt.Errorf("user not found")
		}
		return nil, fmt.Errorf("failed to get user: %w", err)
	}
	
	return r.modelToOutbound(&user), nil
}

// GetByEmail retrieves a user by email
func (r *UserRepositoryAdapter) GetByEmail(ctx context.Context, email string) (*outbound.User, error) {
	user, err := r.queries.GetUserByEmail(ctx, email)
	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, fmt.Errorf("user not found")
		}
		return nil, fmt.Errorf("failed to get user: %w", err)
	}
	
	return r.modelToOutbound(&user), nil
}

// List retrieves all users
func (r *UserRepositoryAdapter) List(ctx context.Context) ([]*outbound.User, error) {
	users, err := r.queries.ListUsers(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to list users: %w", err)
	}
	
	result := make([]*outbound.User, len(users))
	for i, user := range users {
		result[i] = r.modelToOutbound(&user)
	}
	
	return result, nil
}

// Update updates a user
func (r *UserRepositoryAdapter) Update(ctx context.Context, params outbound.UpdateUserParams) (*outbound.User, error) {
	var email pgtype.Text
	var passwordHash pgtype.Text
	
	// Only update fields that are provided
	if params.Email != nil {
		email = pgtype.Text{String: *params.Email, Valid: true}
	}
	if params.PasswordHash != nil {
		passwordHash = pgtype.Text{String: *params.PasswordHash, Valid: true}
	}
	
	user, err := r.queries.UpdateUser(ctx, params.ID, email, passwordHash)
	if err != nil {
		return nil, fmt.Errorf("failed to update user: %w", err)
	}
	
	return r.modelToOutbound(&user), nil
}

// Delete deletes a user
func (r *UserRepositoryAdapter) Delete(ctx context.Context, id string) error {
	err := r.queries.DeleteUser(ctx, id)
	if err != nil {
		return fmt.Errorf("failed to delete user: %w", err)
	}
	return nil
}

// Exists checks if a user exists by ID
func (r *UserRepositoryAdapter) Exists(ctx context.Context, id string) (bool, error) {
	exists, err := r.queries.UserExists(ctx, id)
	if err != nil {
		return false, fmt.Errorf("failed to check user existence: %w", err)
	}
	return exists, nil
}

// ExistsByEmail checks if a user exists by email
func (r *UserRepositoryAdapter) ExistsByEmail(ctx context.Context, email string) (bool, error) {
	exists, err := r.queries.UserExistsByEmail(ctx, email)
	if err != nil {
		return false, fmt.Errorf("failed to check user existence by email: %w", err)
	}
	return exists, nil
}

// Count returns the total number of users
func (r *UserRepositoryAdapter) Count(ctx context.Context) (int64, error) {
	count, err := r.queries.CountUsers(ctx)
	if err != nil {
		return 0, fmt.Errorf("failed to count users: %w", err)
	}
	return count, nil
}

// modelToOutbound converts SQLC model to outbound model
func (r *UserRepositoryAdapter) modelToOutbound(user *sqlc.User) *outbound.User {
	return &outbound.User{
		ID:           user.ID,
		Email:        user.Email,
		PasswordHash: user.PasswordHash,
		CreatedAt:    user.CreatedAt.Time.Format("2006-01-02T15:04:05Z07:00"),
		UpdatedAt:    user.UpdatedAt.Time.Format("2006-01-02T15:04:05Z07:00"),
	}
}

