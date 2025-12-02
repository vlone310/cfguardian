package project

import (
	"context"
	"fmt"

	"github.com/vlone310/cfguardian/internal/ports/outbound"
)

// ProjectListItem represents a project in the list
type ProjectListItem struct {
	ID          string
	Name        string
	APIKey      string
	OwnerUserID string
	CreatedAt   string
	UpdatedAt   string
}

// ListProjectsRequest holds list request parameters
type ListProjectsRequest struct {
	// Optional: filter by owner
	OwnerUserID *string
}

// ListProjectsResponse holds the list of projects
type ListProjectsResponse struct {
	Projects []*ProjectListItem
	Total    int64
}

// ListProjectsUseCase handles listing projects
type ListProjectsUseCase struct {
	projectRepo outbound.ProjectRepository
}

// NewListProjectsUseCase creates a new ListProjectsUseCase
func NewListProjectsUseCase(projectRepo outbound.ProjectRepository) *ListProjectsUseCase {
	return &ListProjectsUseCase{
		projectRepo: projectRepo,
	}
}

// Execute retrieves projects (optionally filtered by owner)
func (uc *ListProjectsUseCase) Execute(ctx context.Context, req ListProjectsRequest) (*ListProjectsResponse, error) {
	var projects []*outbound.Project
	var count int64
	var err error
	
	// Filter by owner if specified
	if req.OwnerUserID != nil && *req.OwnerUserID != "" {
		projects, err = uc.projectRepo.ListByOwner(ctx, *req.OwnerUserID)
		if err != nil {
			return nil, fmt.Errorf("failed to list projects by owner: %w", err)
		}
		
		count, err = uc.projectRepo.CountByOwner(ctx, *req.OwnerUserID)
		if err != nil {
			return nil, fmt.Errorf("failed to count projects by owner: %w", err)
		}
	} else {
		// List all projects
		projects, err = uc.projectRepo.List(ctx)
		if err != nil {
			return nil, fmt.Errorf("failed to list projects: %w", err)
		}
		
		count, err = uc.projectRepo.Count(ctx)
		if err != nil {
			return nil, fmt.Errorf("failed to count projects: %w", err)
		}
	}
	
	// Convert to response format
	items := make([]*ProjectListItem, len(projects))
	for i, project := range projects {
		items[i] = &ProjectListItem{
			ID:          project.ID,
			Name:        project.Name,
			APIKey:      project.APIKey,
			OwnerUserID: project.OwnerUserID,
			CreatedAt:   project.CreatedAt,
			UpdatedAt:   project.UpdatedAt,
		}
	}
	
	return &ListProjectsResponse{
		Projects: items,
		Total:    count,
	}, nil
}

