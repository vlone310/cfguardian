package user

import (
	"context"
	"fmt"

	"github.com/vlone310/cfguardian/internal/ports/outbound"
)

// UserListItem represents a user in the list
type UserListItem struct {
	ID        string
	Email     string
	CreatedAt string
	UpdatedAt string
}

// ListUsersResponse holds the list of users
type ListUsersResponse struct {
	Users []*UserListItem
	Total int64
}

// ListUsersUseCase handles listing all users
type ListUsersUseCase struct {
	userRepo outbound.UserRepository
}

// NewListUsersUseCase creates a new ListUsersUseCase
func NewListUsersUseCase(userRepo outbound.UserRepository) *ListUsersUseCase {
	return &ListUsersUseCase{
		userRepo: userRepo,
	}
}

// Execute retrieves all users
func (uc *ListUsersUseCase) Execute(ctx context.Context) (*ListUsersResponse, error) {
	// Get all users
	users, err := uc.userRepo.List(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to list users: %w", err)
	}
	
	// Get total count
	count, err := uc.userRepo.Count(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to count users: %w", err)
	}
	
	// Convert to response format
	items := make([]*UserListItem, len(users))
	for i, user := range users {
		items[i] = &UserListItem{
			ID:        user.ID,
			Email:     user.Email,
			CreatedAt: user.CreatedAt,
			UpdatedAt: user.UpdatedAt,
		}
	}
	
	return &ListUsersResponse{
		Users: items,
		Total: count,
	}, nil
}

