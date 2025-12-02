package project

import (
	"context"
	"fmt"

	"github.com/vlone310/cfguardian/internal/ports/outbound"
)

// DeleteProjectRequest holds delete project request data
type DeleteProjectRequest struct {
	ProjectID string
}

// DeleteProjectUseCase handles project deletion (admin only)
type DeleteProjectUseCase struct {
	projectRepo outbound.ProjectRepository
}

// NewDeleteProjectUseCase creates a new DeleteProjectUseCase
func NewDeleteProjectUseCase(projectRepo outbound.ProjectRepository) *DeleteProjectUseCase {
	return &DeleteProjectUseCase{
		projectRepo: projectRepo,
	}
}

// Execute deletes a project
func (uc *DeleteProjectUseCase) Execute(ctx context.Context, req DeleteProjectRequest) error {
	if req.ProjectID == "" {
		return fmt.Errorf("project ID is required")
	}
	
	// Check if project exists
	exists, err := uc.projectRepo.Exists(ctx, req.ProjectID)
	if err != nil {
		return fmt.Errorf("failed to check if project exists: %w", err)
	}
	if !exists {
		return fmt.Errorf("project not found")
	}
	
	// Note: Foreign key CASCADE will automatically delete:
	// - All configs in the project
	// - All config revisions
	// - All roles associated with the project
	
	// Delete project (cascading deletes handled by database)
	if err := uc.projectRepo.Delete(ctx, req.ProjectID); err != nil {
		return fmt.Errorf("failed to delete project: %w", err)
	}
	
	return nil
}

