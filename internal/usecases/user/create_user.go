package user

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"github.com/vlone310/cfguardian/internal/domain/entities"
	"github.com/vlone310/cfguardian/internal/domain/services"
	"github.com/vlone310/cfguardian/internal/domain/valueobjects"
	"github.com/vlone310/cfguardian/internal/ports/outbound"
)

// CreateUserRequest holds user creation data
type CreateUserRequest struct {
	Email    string
	Password string
}

// CreateUserResponse holds created user data
type CreateUserResponse struct {
	UserID    string
	Email     string
	CreatedAt string
}

// CreateUserUseCase handles user creation (admin only)
type CreateUserUseCase struct {
	userRepo       outbound.UserRepository
	passwordHasher *services.PasswordHasher
}

// NewCreateUserUseCase creates a new CreateUserUseCase
func NewCreateUserUseCase(
	userRepo outbound.UserRepository,
	passwordHasher *services.PasswordHasher,
) *CreateUserUseCase {
	return &CreateUserUseCase{
		userRepo:       userRepo,
		passwordHasher: passwordHasher,
	}
}

// Execute creates a new user
func (uc *CreateUserUseCase) Execute(ctx context.Context, req CreateUserRequest) (*CreateUserResponse, error) {
	// Validate email
	email, err := valueobjects.NewEmail(req.Email)
	if err != nil {
		return nil, fmt.Errorf("invalid email: %w", err)
	}
	
	// Check if user already exists
	exists, err := uc.userRepo.ExistsByEmail(ctx, email.Value())
	if err != nil {
		return nil, fmt.Errorf("failed to check if user exists: %w", err)
	}
	if exists {
		return nil, fmt.Errorf("user with email %s already exists", email.Value())
	}
	
	// Hash password
	passwordHash, err := uc.passwordHasher.Hash(req.Password)
	if err != nil {
		return nil, fmt.Errorf("failed to hash password: %w", err)
	}
	
	// Create domain entity
	userID := uuid.New().String()
	userEntity := entities.NewUser(userID, email, passwordHash)
	
	// Persist to repository
	user, err := uc.userRepo.Create(ctx, outbound.CreateUserParams{
		ID:           userEntity.ID(),
		Email:        userEntity.Email().Value(),
		PasswordHash: userEntity.PasswordHash(),
	})
	if err != nil {
		return nil, fmt.Errorf("failed to create user: %w", err)
	}
	
	return &CreateUserResponse{
		UserID:    user.ID,
		Email:     user.Email,
		CreatedAt: user.CreatedAt,
	}, nil
}

