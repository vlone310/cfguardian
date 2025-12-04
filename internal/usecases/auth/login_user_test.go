package auth

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"github.com/vlone310/cfguardian/internal/domain/services"
	"github.com/vlone310/cfguardian/internal/ports/outbound"
)

// MockUserRepository is a mock implementation of UserRepository
type MockUserRepository struct {
	mock.Mock
}

func (m *MockUserRepository) Create(ctx context.Context, params outbound.CreateUserParams) (*outbound.User, error) {
	args := m.Called(ctx, params)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*outbound.User), args.Error(1)
}

func (m *MockUserRepository) GetByID(ctx context.Context, id string) (*outbound.User, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*outbound.User), args.Error(1)
}

func (m *MockUserRepository) GetByEmail(ctx context.Context, email string) (*outbound.User, error) {
	args := m.Called(ctx, email)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*outbound.User), args.Error(1)
}

func (m *MockUserRepository) Update(ctx context.Context, params outbound.UpdateUserParams) (*outbound.User, error) {
	args := m.Called(ctx, params)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*outbound.User), args.Error(1)
}

func (m *MockUserRepository) Delete(ctx context.Context, id string) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}

func (m *MockUserRepository) List(ctx context.Context) ([]*outbound.User, error) {
	args := m.Called(ctx)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*outbound.User), args.Error(1)
}

func (m *MockUserRepository) Count(ctx context.Context) (int64, error) {
	args := m.Called(ctx)
	return args.Get(0).(int64), args.Error(1)
}

func (m *MockUserRepository) Exists(ctx context.Context, id string) (bool, error) {
	args := m.Called(ctx, id)
	return args.Get(0).(bool), args.Error(1)
}

func (m *MockUserRepository) ExistsByEmail(ctx context.Context, email string) (bool, error) {
	args := m.Called(ctx, email)
	return args.Get(0).(bool), args.Error(1)
}

func TestLoginUserUseCase_Execute_Success(t *testing.T) {
	// Arrange
	ctx := context.Background()
	mockRepo := new(MockUserRepository)
	hasher := services.NewPasswordHasher(10)
	
	useCase := NewLoginUserUseCase(mockRepo, hasher)
	
	// Create a test user with hashed password
	email := "test@example.com"
	password := "TestPassword123"
	passwordHash, err := hasher.Hash(password)
	require.NoError(t, err)
	
	testUser := &outbound.User{
		ID:           "user-123",
		Email:        email,
		PasswordHash: passwordHash,
	}
	
	// Mock expectations
	mockRepo.On("GetByEmail", ctx, email).Return(testUser, nil)
	
	// Act
	request := LoginRequest{
		Email:    "test@example.com",
		Password: password,
	}
	
	response, err := useCase.Execute(ctx, request)
	
	// Assert
	require.NoError(t, err)
	assert.NotNil(t, response)
	assert.Equal(t, "user-123", response.UserID)
	assert.Equal(t, "test@example.com", response.Email)
	
	// Verify mock expectations
	mockRepo.AssertExpectations(t)
}

func TestLoginUserUseCase_Execute_InvalidEmail(t *testing.T) {
	// Arrange
	ctx := context.Background()
	mockRepo := new(MockUserRepository)
	hasher := services.NewPasswordHasher(10)
	
	useCase := NewLoginUserUseCase(mockRepo, hasher)
	
	// Mock expectations - invalid email won't be found in database
	mockRepo.On("GetByEmail", ctx, "invalid-email").Return(nil, assert.AnError)
	
	// Act
	request := LoginRequest{
		Email:    "invalid-email",
		Password: "password",
	}
	
	response, err := useCase.Execute(ctx, request)
	
	// Assert
	require.Error(t, err)
	assert.Nil(t, response)
	assert.Contains(t, err.Error(), "invalid")
	
	mockRepo.AssertExpectations(t)
}

func TestLoginUserUseCase_Execute_EmptyPassword(t *testing.T) {
	// Arrange
	ctx := context.Background()
	mockRepo := new(MockUserRepository)
	hasher := services.NewPasswordHasher(10)
	
	useCase := NewLoginUserUseCase(mockRepo, hasher)
	
	// Act
	request := LoginRequest{
		Email:    "test@example.com",
		Password: "",
	}
	
	response, err := useCase.Execute(ctx, request)
	
	// Assert
	require.Error(t, err)
	assert.Nil(t, response)
	assert.Contains(t, err.Error(), "password")
	
	// Repository should not be called
	mockRepo.AssertNotCalled(t, "GetByEmail")
}

func TestLoginUserUseCase_Execute_UserNotFound(t *testing.T) {
	// Arrange
	ctx := context.Background()
	mockRepo := new(MockUserRepository)
	hasher := services.NewPasswordHasher(10)
	
	useCase := NewLoginUserUseCase(mockRepo, hasher)
	
	email := "nonexistent@example.com"
	
	// Mock expectations - user not found
	mockRepo.On("GetByEmail", ctx, email).Return(nil, assert.AnError)
	
	// Act
	request := LoginRequest{
		Email:    "nonexistent@example.com",
		Password: "password",
	}
	
	response, err := useCase.Execute(ctx, request)
	
	// Assert
	require.Error(t, err)
	assert.Nil(t, response)
	
	mockRepo.AssertExpectations(t)
}

func TestLoginUserUseCase_Execute_WrongPassword(t *testing.T) {
	// Arrange
	ctx := context.Background()
	mockRepo := new(MockUserRepository)
	hasher := services.NewPasswordHasher(10)
	
	useCase := NewLoginUserUseCase(mockRepo, hasher)
	
	// Create a test user with hashed password
	email := "test@example.com"
	correctPassword := "CorrectPassword123"
	passwordHash, err := hasher.Hash(correctPassword)
	require.NoError(t, err)
	
	testUser := &outbound.User{
		ID:           "user-123",
		Email:        email,
		PasswordHash: passwordHash,
	}
	
	// Mock expectations
	mockRepo.On("GetByEmail", ctx, email).Return(testUser, nil)
	
	// Act - try with wrong password
	request := LoginRequest{
		Email:    "test@example.com",
		Password: "WrongPassword456",
	}
	
	response, err := useCase.Execute(ctx, request)
	
	// Assert
	require.Error(t, err)
	assert.Nil(t, response)
	assert.Contains(t, err.Error(), "invalid")
	
	mockRepo.AssertExpectations(t)
}

func TestLoginUserUseCase_Execute_CaseSensitivePassword(t *testing.T) {
	// Arrange
	ctx := context.Background()
	mockRepo := new(MockUserRepository)
	hasher := services.NewPasswordHasher(10)
	
	useCase := NewLoginUserUseCase(mockRepo, hasher)
	
	// Create a test user with hashed password
	email := "test@example.com"
	correctPassword := "TestPassword123"
	passwordHash, err := hasher.Hash(correctPassword)
	require.NoError(t, err)
	
	testUser := &outbound.User{
		ID:           "user-123",
		Email:        email,
		PasswordHash: passwordHash,
	}
	
	// Mock expectations
	mockRepo.On("GetByEmail", ctx, email).Return(testUser, nil)
	
	// Act - try with wrong case
	request := LoginRequest{
		Email:    "test@example.com",
		Password: "testpassword123", // lowercase
	}
	
	response, err := useCase.Execute(ctx, request)
	
	// Assert - should fail (passwords are case-sensitive)
	require.Error(t, err)
	assert.Nil(t, response)
	
	mockRepo.AssertExpectations(t)
}

// Note: Email normalization is handled at the repository layer,
// not in the use case. The use case just passes the email as-is.

