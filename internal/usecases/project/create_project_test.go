package project

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"github.com/vlone310/cfguardian/internal/domain/services"
	"github.com/vlone310/cfguardian/internal/ports/outbound"
)

// MockProjectRepository is a mock implementation of ProjectRepository
type MockProjectRepository struct {
	mock.Mock
}

func (m *MockProjectRepository) Create(ctx context.Context, params outbound.CreateProjectParams) (*outbound.Project, error) {
	args := m.Called(ctx, params)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*outbound.Project), args.Error(1)
}

func (m *MockProjectRepository) GetByID(ctx context.Context, id string) (*outbound.Project, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*outbound.Project), args.Error(1)
}

func (m *MockProjectRepository) GetByAPIKey(ctx context.Context, apiKey string) (*outbound.Project, error) {
	args := m.Called(ctx, apiKey)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*outbound.Project), args.Error(1)
}

func (m *MockProjectRepository) Update(ctx context.Context, params outbound.UpdateProjectParams) (*outbound.Project, error) {
	args := m.Called(ctx, params)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*outbound.Project), args.Error(1)
}

func (m *MockProjectRepository) Delete(ctx context.Context, id string) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}

func (m *MockProjectRepository) ListByOwner(ctx context.Context, ownerUserID string) ([]*outbound.Project, error) {
	args := m.Called(ctx, ownerUserID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*outbound.Project), args.Error(1)
}

func (m *MockProjectRepository) List(ctx context.Context) ([]*outbound.Project, error) {
	args := m.Called(ctx)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*outbound.Project), args.Error(1)
}

func (m *MockProjectRepository) Exists(ctx context.Context, id string) (bool, error) {
	args := m.Called(ctx, id)
	return args.Get(0).(bool), args.Error(1)
}

func (m *MockProjectRepository) ExistsByName(ctx context.Context, name string) (bool, error) {
	args := m.Called(ctx, name)
	return args.Get(0).(bool), args.Error(1)
}

func (m *MockProjectRepository) ExistsByAPIKey(ctx context.Context, apiKey string) (bool, error) {
	args := m.Called(ctx, apiKey)
	return args.Get(0).(bool), args.Error(1)
}

func (m *MockProjectRepository) Count(ctx context.Context) (int64, error) {
	args := m.Called(ctx)
	return args.Get(0).(int64), args.Error(1)
}

func (m *MockProjectRepository) CountByOwner(ctx context.Context, ownerUserID string) (int64, error) {
	args := m.Called(ctx, ownerUserID)
	return args.Get(0).(int64), args.Error(1)
}

// MockUserRepository for user existence checks
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

// MockRoleRepository for role assignment
type MockRoleRepository struct {
	mock.Mock
}

func (m *MockRoleRepository) Assign(ctx context.Context, params outbound.AssignRoleParams) (*outbound.Role, error) {
	args := m.Called(ctx, params)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*outbound.Role), args.Error(1)
}

func (m *MockRoleRepository) Revoke(ctx context.Context, userID, projectID string) error {
	args := m.Called(ctx, userID, projectID)
	return args.Error(0)
}

func (m *MockRoleRepository) GetRole(ctx context.Context, userID, projectID string) (*outbound.Role, error) {
	args := m.Called(ctx, userID, projectID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*outbound.Role), args.Error(1)
}


func (m *MockRoleRepository) CountByProject(ctx context.Context, projectID string) (int64, error) {
	args := m.Called(ctx, projectID)
	return args.Get(0).(int64), args.Error(1)
}

func (m *MockRoleRepository) CountByUser(ctx context.Context, userID string) (int64, error) {
	args := m.Called(ctx, userID)
	return args.Get(0).(int64), args.Error(1)
}

func (m *MockRoleRepository) ListByLevel(ctx context.Context, level outbound.RoleLevel) ([]*outbound.Role, error) {
	args := m.Called(ctx, level)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*outbound.Role), args.Error(1)
}

func (m *MockRoleRepository) Get(ctx context.Context, userID, projectID string) (*outbound.Role, error) {
	args := m.Called(ctx, userID, projectID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*outbound.Role), args.Error(1)
}

func (m *MockRoleRepository) GetUserRole(ctx context.Context, userID, projectID string) (outbound.RoleLevel, error) {
	args := m.Called(ctx, userID, projectID)
	return args.Get(0).(outbound.RoleLevel), args.Error(1)
}

func (m *MockRoleRepository) ListUserRoles(ctx context.Context, userID string) ([]*outbound.UserRoleWithProject, error) {
	args := m.Called(ctx, userID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*outbound.UserRoleWithProject), args.Error(1)
}

func (m *MockRoleRepository) ListProjectRoles(ctx context.Context, projectID string) ([]*outbound.ProjectRoleWithUser, error) {
	args := m.Called(ctx, projectID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*outbound.ProjectRoleWithUser), args.Error(1)
}

func (m *MockRoleRepository) Update(ctx context.Context, userID, projectID string, level outbound.RoleLevel) (*outbound.Role, error) {
	args := m.Called(ctx, userID, projectID, level)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*outbound.Role), args.Error(1)
}

func (m *MockRoleRepository) RevokeAllUserRoles(ctx context.Context, userID string) error {
	args := m.Called(ctx, userID)
	return args.Error(0)
}

func (m *MockRoleRepository) RevokeAllProjectRoles(ctx context.Context, projectID string) error {
	args := m.Called(ctx, projectID)
	return args.Error(0)
}

func (m *MockRoleRepository) Exists(ctx context.Context, userID, projectID string) (bool, error) {
	args := m.Called(ctx, userID, projectID)
	return args.Get(0).(bool), args.Error(1)
}

func (m *MockRoleRepository) HasRole(ctx context.Context, userID, projectID string, level outbound.RoleLevel) (bool, error) {
	args := m.Called(ctx, userID, projectID, level)
	return args.Get(0).(bool), args.Error(1)
}

func (m *MockRoleRepository) HasMinimumRole(ctx context.Context, userID, projectID string, minLevel outbound.RoleLevel) (bool, error) {
	args := m.Called(ctx, userID, projectID, minLevel)
	return args.Get(0).(bool), args.Error(1)
}

func TestCreateProjectUseCase_Execute_Success(t *testing.T) {
	// Arrange
	ctx := context.Background()
	mockProjectRepo := new(MockProjectRepository)
	mockUserRepo := new(MockUserRepository)
	mockRoleRepo := new(MockRoleRepository)
	apiKeyGen := services.NewAPIKeyGenerator()
	
	useCase := NewCreateProjectUseCase(mockProjectRepo, mockUserRepo, mockRoleRepo, apiKeyGen)
	
	// Mock expectations
	mockUserRepo.On("Exists", ctx, "user-123").Return(true, nil)
	mockProjectRepo.On("ExistsByName", ctx, "My Project").Return(false, nil)
	mockProjectRepo.On("Create", ctx, mock.AnythingOfType("outbound.CreateProjectParams")).Return(&outbound.Project{
		ID:          "proj-123",
		Name:        "My Project",
		APIKey:      "cfg_abcd1234567890abcdef1234567890ab",
		OwnerUserID: "user-123",
	}, nil)
	mockRoleRepo.On("Assign", ctx, mock.AnythingOfType("outbound.AssignRoleParams")).Return(&outbound.Role{
		UserID:    "user-123",
		ProjectID: "proj-123",
		RoleLevel: outbound.RoleLevelAdmin,
	}, nil)
	
	// Act
	request := CreateProjectRequest{
		Name:        "My Project",
		OwnerUserID: "user-123",
	}
	
	response, err := useCase.Execute(ctx, request)
	
	// Assert
	require.NoError(t, err)
	assert.NotNil(t, response)
	assert.Equal(t, "My Project", response.Name)
	assert.Equal(t, "user-123", response.OwnerUserID)
	assert.NotEmpty(t, response.ID)
	assert.NotEmpty(t, response.APIKey)
	
	// API key should have correct format
	assert.Contains(t, response.APIKey, "cfg_")
	assert.Equal(t, 36, len(response.APIKey))
	
	mockProjectRepo.AssertExpectations(t)
	mockRoleRepo.AssertExpectations(t)
}

func TestCreateProjectUseCase_Execute_EmptyName(t *testing.T) {
	// Arrange
	ctx := context.Background()
	mockProjectRepo := new(MockProjectRepository)
	mockUserRepo := new(MockUserRepository)
	mockRoleRepo := new(MockRoleRepository)
	apiKeyGen := services.NewAPIKeyGenerator()
	
	useCase := NewCreateProjectUseCase(mockProjectRepo, mockUserRepo, mockRoleRepo, apiKeyGen)
	
	// Act
	request := CreateProjectRequest{
		Name:        "",
		OwnerUserID: "user-123",
	}
	
	response, err := useCase.Execute(ctx, request)
	
	// Assert
	require.Error(t, err)
	assert.Nil(t, response)
	assert.Contains(t, err.Error(), "name")
	
	// Repository should not be called
	mockProjectRepo.AssertNotCalled(t, "Create")
	mockRoleRepo.AssertNotCalled(t, "Assign")
}

func TestCreateProjectUseCase_Execute_EmptyOwnerID(t *testing.T) {
	// Arrange
	ctx := context.Background()
	mockProjectRepo := new(MockProjectRepository)
	mockUserRepo := new(MockUserRepository)
	mockRoleRepo := new(MockRoleRepository)
	apiKeyGen := services.NewAPIKeyGenerator()
	
	useCase := NewCreateProjectUseCase(mockProjectRepo, mockUserRepo, mockRoleRepo, apiKeyGen)
	
	// Act
	request := CreateProjectRequest{
		Name:        "My Project",
		OwnerUserID: "",
	}
	
	response, err := useCase.Execute(ctx, request)
	
	// Assert
	require.Error(t, err)
	assert.Nil(t, response)
	assert.Contains(t, err.Error(), "owner")
	
	// Repository should not be called
	mockProjectRepo.AssertNotCalled(t, "Create")
	mockRoleRepo.AssertNotCalled(t, "Assign")
}

func TestCreateProjectUseCase_Execute_DuplicateName(t *testing.T) {
	// Arrange
	ctx := context.Background()
	mockProjectRepo := new(MockProjectRepository)
	mockUserRepo := new(MockUserRepository)
	mockRoleRepo := new(MockRoleRepository)
	apiKeyGen := services.NewAPIKeyGenerator()
	
	useCase := NewCreateProjectUseCase(mockProjectRepo, mockUserRepo, mockRoleRepo, apiKeyGen)
	
	// Mock expectations - owner exists
	mockUserRepo.On("Exists", ctx, "user-123").Return(true, nil)
	// Project with same name exists
	mockProjectRepo.On("ExistsByName", ctx, "Existing Project").Return(true, nil)
	
	// Act
	request := CreateProjectRequest{
		Name:        "Existing Project",
		OwnerUserID: "user-123",
	}
	
	response, err := useCase.Execute(ctx, request)
	
	// Assert
	require.Error(t, err)
	assert.Nil(t, response)
	assert.Contains(t, err.Error(), "already exists")
	
	mockProjectRepo.AssertExpectations(t)
	// Create should not be called
	mockProjectRepo.AssertNotCalled(t, "Create")
	mockRoleRepo.AssertNotCalled(t, "Assign")
}

func TestCreateProjectUseCase_Execute_AutoAssignsAdminRole(t *testing.T) {
	// Arrange
	ctx := context.Background()
	mockProjectRepo := new(MockProjectRepository)
	mockUserRepo := new(MockUserRepository)
	mockRoleRepo := new(MockRoleRepository)
	apiKeyGen := services.NewAPIKeyGenerator()
	
	useCase := NewCreateProjectUseCase(mockProjectRepo, mockUserRepo, mockRoleRepo, apiKeyGen)
	
	// Mock expectations
	mockUserRepo.On("Exists", ctx, "user-123").Return(true, nil)
	mockProjectRepo.On("ExistsByName", ctx, "New Project").Return(false, nil)
	mockProjectRepo.On("Create", ctx, mock.AnythingOfType("outbound.CreateProjectParams")).Return(&outbound.Project{
		ID:          "proj-123",
		Name:        "New Project",
		APIKey:      "cfg_abcd1234567890abcdef1234567890ab",
		OwnerUserID: "user-123",
	}, nil)
	
	// Capture the role assignment params
	var assignedRoleParams outbound.AssignRoleParams
	mockRoleRepo.On("Assign", ctx, mock.AnythingOfType("outbound.AssignRoleParams")).
		Run(func(args mock.Arguments) {
			assignedRoleParams = args.Get(1).(outbound.AssignRoleParams)
		}).Return(&outbound.Role{
		UserID:    "user-123",
		ProjectID: "proj-123",
		RoleLevel: outbound.RoleLevelAdmin,
	}, nil)
	
	// Act
	request := CreateProjectRequest{
		Name:        "New Project",
		OwnerUserID: "user-123",
	}
	
	response, err := useCase.Execute(ctx, request)
	
	// Assert
	require.NoError(t, err)
	assert.NotNil(t, response)
	
	// Verify admin role was assigned
	assert.Equal(t, "user-123", assignedRoleParams.UserID)
	assert.Equal(t, outbound.RoleLevelAdmin, assignedRoleParams.RoleLevel)
	
	mockProjectRepo.AssertExpectations(t)
	mockUserRepo.AssertExpectations(t)
	mockRoleRepo.AssertExpectations(t)
}

