package project

import (
	"context"
	"fmt"

	"github.com/vlone310/cfguardian/internal/ports/outbound"
)

// GetProjectRequest holds get project request data
type GetProjectRequest struct {
	ProjectID string
}

// GetProjectResponse holds project data
type GetProjectResponse struct {
	ID          string
	Name        string
	APIKey      string
	OwnerUserID string
	CreatedAt   string
	UpdatedAt   string
}

// GetProjectUseCase handles retrieving a project by ID
type GetProjectUseCase struct {
	projectRepo outbound.ProjectRepository
}

// NewGetProjectUseCase creates a new GetProjectUseCase
func NewGetProjectUseCase(projectRepo outbound.ProjectRepository) *GetProjectUseCase {
	return &GetProjectUseCase{
		projectRepo: projectRepo,
	}
}

// Execute retrieves a project by ID
func (uc *GetProjectUseCase) Execute(ctx context.Context, req GetProjectRequest) (*GetProjectResponse, error) {
	if req.ProjectID == "" {
		return nil, fmt.Errorf("project ID is required")
	}
	
	// Get project from repository
	project, err := uc.projectRepo.GetByID(ctx, req.ProjectID)
	if err != nil {
		return nil, fmt.Errorf("project not found: %w", err)
	}
	
	return &GetProjectResponse{
		ID:          project.ID,
		Name:        project.Name,
		APIKey:      project.APIKey,
		OwnerUserID: project.OwnerUserID,
		CreatedAt:   project.CreatedAt,
		UpdatedAt:   project.UpdatedAt,
	}, nil
}

