package project

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"github.com/vlone310/cfguardian/internal/domain/entities"
	"github.com/vlone310/cfguardian/internal/domain/services"
	"github.com/vlone310/cfguardian/internal/domain/valueobjects"
	"github.com/vlone310/cfguardian/internal/ports/outbound"
)

// CreateProjectRequest holds project creation data
type CreateProjectRequest struct {
	Name        string `json:"name"`
	OwnerUserID string `json:"owner_user_id"`
}

// CreateProjectResponse holds created project data
type CreateProjectResponse struct {
	ID          string `json:"id"`
	Name        string `json:"name"`
	APIKey      string `json:"api_key"`
	OwnerUserID string `json:"owner_user_id"`
	CreatedAt   string `json:"created_at"`
}

// CreateProjectUseCase handles project creation
type CreateProjectUseCase struct {
	projectRepo     outbound.ProjectRepository
	userRepo        outbound.UserRepository
	roleRepo        outbound.RoleRepository
	apiKeyGenerator *services.APIKeyGenerator
}

// NewCreateProjectUseCase creates a new CreateProjectUseCase
func NewCreateProjectUseCase(
	projectRepo outbound.ProjectRepository,
	userRepo outbound.UserRepository,
	roleRepo outbound.RoleRepository,
	apiKeyGenerator *services.APIKeyGenerator,
) *CreateProjectUseCase {
	return &CreateProjectUseCase{
		projectRepo:     projectRepo,
		userRepo:        userRepo,
		roleRepo:        roleRepo,
		apiKeyGenerator: apiKeyGenerator,
	}
}

// Execute creates a new project
func (uc *CreateProjectUseCase) Execute(ctx context.Context, req CreateProjectRequest) (*CreateProjectResponse, error) {
	// Validate input
	if req.Name == "" {
		return nil, fmt.Errorf("project name is required")
	}
	if req.OwnerUserID == "" {
		return nil, fmt.Errorf("owner user ID is required")
	}
	
	// Verify owner exists
	ownerExists, err := uc.userRepo.Exists(ctx, req.OwnerUserID)
	if err != nil {
		return nil, fmt.Errorf("failed to verify owner exists: %w", err)
	}
	if !ownerExists {
		return nil, fmt.Errorf("owner user not found")
	}
	
	// Check if project name already exists
	nameExists, err := uc.projectRepo.ExistsByName(ctx, req.Name)
	if err != nil {
		return nil, fmt.Errorf("failed to check if project name exists: %w", err)
	}
	if nameExists {
		return nil, fmt.Errorf("project with name '%s' already exists", req.Name)
	}
	
	// Generate API key
	apiKey, err := uc.apiKeyGenerator.Generate()
	if err != nil {
		return nil, fmt.Errorf("failed to generate API key: %w", err)
	}
	
	// Create domain entity
	projectID := uuid.New().String()
	projectEntity := entities.NewProject(projectID, req.Name, apiKey, req.OwnerUserID)
	
	// Persist to repository
	project, err := uc.projectRepo.Create(ctx, outbound.CreateProjectParams{
		ID:          projectEntity.ID(),
		Name:        projectEntity.Name(),
		APIKey:      projectEntity.APIKey().Value(),
		OwnerUserID: projectEntity.OwnerUserID(),
	})
	if err != nil {
		return nil, fmt.Errorf("failed to create project: %w", err)
	}
	
	// Automatically assign admin role to owner
	_, err = uc.roleRepo.Assign(ctx, outbound.AssignRoleParams{
		UserID:    req.OwnerUserID,
		ProjectID: projectID,
		RoleLevel: outbound.RoleLevel(valueobjects.RoleLevelAdmin),
	})
	if err != nil {
		// Log error but don't fail project creation
		// The project is created, but owner doesn't have explicit admin role yet
		// This can be fixed manually
		return nil, fmt.Errorf("project created but failed to assign admin role: %w", err)
	}
	
	return &CreateProjectResponse{
		ID:          project.ID,
		Name:        project.Name,
		APIKey:      project.APIKey,
		OwnerUserID: project.OwnerUserID,
		CreatedAt:   project.CreatedAt,
	}, nil
}

