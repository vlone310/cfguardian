package auth

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"github.com/vlone310/cfguardian/internal/domain/services"
	"github.com/vlone310/cfguardian/internal/domain/valueobjects"
	"github.com/vlone310/cfguardian/internal/ports/outbound"
)

// RegisterRequest holds user registration data
type RegisterRequest struct {
	Email    string
	Password string
}

// RegisterResponse holds registration response data
type RegisterResponse struct {
	UserID string
	Email  string
}

// RegisterUserUseCase handles user registration
type RegisterUserUseCase struct {
	userRepo       outbound.UserRepository
	passwordHasher *services.PasswordHasher
}

// NewRegisterUserUseCase creates a new RegisterUserUseCase
func NewRegisterUserUseCase(
	userRepo outbound.UserRepository,
	passwordHasher *services.PasswordHasher,
) *RegisterUserUseCase {
	return &RegisterUserUseCase{
		userRepo:       userRepo,
		passwordHasher: passwordHasher,
	}
}

// Execute performs user registration
func (uc *RegisterUserUseCase) Execute(ctx context.Context, req RegisterRequest) (*RegisterResponse, error) {
	// Validate email
	email, err := valueobjects.NewEmail(req.Email)
	if err != nil {
		return nil, fmt.Errorf("invalid email: %w", err)
	}
	
	// Validate password
	if req.Password == "" {
		return nil, fmt.Errorf("password is required")
	}
	if len(req.Password) < 8 {
		return nil, fmt.Errorf("password must be at least 8 characters")
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
	
	// Create user
	userID := uuid.New().String()
	user, err := uc.userRepo.Create(ctx, outbound.CreateUserParams{
		ID:           userID,
		Email:        email.Value(),
		PasswordHash: passwordHash,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to create user: %w", err)
	}
	
	return &RegisterResponse{
		UserID: user.ID,
		Email:  user.Email,
	}, nil
}

