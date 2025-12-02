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

// ProjectRepositoryAdapter implements outbound.ProjectRepository using PostgreSQL
type ProjectRepositoryAdapter struct {
	pool    *pgxpool.Pool
	queries *sqlc.Queries
}

// NewProjectRepositoryAdapter creates a new PostgreSQL project repository
func NewProjectRepositoryAdapter(pool *pgxpool.Pool) *ProjectRepositoryAdapter {
	return &ProjectRepositoryAdapter{
		pool:    pool,
		queries: sqlc.New(pool),
	}
}

// Create creates a new project
func (r *ProjectRepositoryAdapter) Create(ctx context.Context, params outbound.CreateProjectParams) (*outbound.Project, error) {
	project, err := r.queries.CreateProject(ctx, sqlc.CreateProjectParams{
		ID:          params.ID,
		Name:        params.Name,
		ApiKey:      params.APIKey,
		OwnerUserID: params.OwnerUserID,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to create project: %w", err)
	}
	
	return r.modelToOutbound(&project), nil
}

// GetByID retrieves a project by ID
func (r *ProjectRepositoryAdapter) GetByID(ctx context.Context, id string) (*outbound.Project, error) {
	project, err := r.queries.GetProjectByID(ctx, id)
	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, fmt.Errorf("project not found")
		}
		return nil, fmt.Errorf("failed to get project: %w", err)
	}
	
	return r.modelToOutbound(&project), nil
}

// GetByAPIKey retrieves a project by API key
func (r *ProjectRepositoryAdapter) GetByAPIKey(ctx context.Context, apiKey string) (*outbound.Project, error) {
	project, err := r.queries.GetProjectByAPIKey(ctx, apiKey)
	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, fmt.Errorf("project not found")
		}
		return nil, fmt.Errorf("failed to get project: %w", err)
	}
	
	return r.modelToOutbound(&project), nil
}

// List retrieves all projects
func (r *ProjectRepositoryAdapter) List(ctx context.Context) ([]*outbound.Project, error) {
	projects, err := r.queries.ListProjects(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to list projects: %w", err)
	}
	
	result := make([]*outbound.Project, len(projects))
	for i, project := range projects {
		result[i] = r.modelToOutbound(&project)
	}
	
	return result, nil
}

// ListByOwner retrieves projects by owner user ID
func (r *ProjectRepositoryAdapter) ListByOwner(ctx context.Context, ownerUserID string) ([]*outbound.Project, error) {
	projects, err := r.queries.ListProjectsByOwner(ctx, ownerUserID)
	if err != nil {
		return nil, fmt.Errorf("failed to list projects by owner: %w", err)
	}
	
	result := make([]*outbound.Project, len(projects))
	for i, project := range projects {
		result[i] = r.modelToOutbound(&project)
	}
	
	return result, nil
}

// Update updates a project
func (r *ProjectRepositoryAdapter) Update(ctx context.Context, params outbound.UpdateProjectParams) (*outbound.Project, error) {
	var name pgtype.Text
	var apiKey pgtype.Text
	
	// Only update fields that are provided
	if params.Name != nil {
		name = pgtype.Text{String: *params.Name, Valid: true}
	}
	if params.APIKey != nil {
		apiKey = pgtype.Text{String: *params.APIKey, Valid: true}
	}
	
	project, err := r.queries.UpdateProject(ctx, params.ID, name, apiKey)
	if err != nil {
		return nil, fmt.Errorf("failed to update project: %w", err)
	}
	
	return r.modelToOutbound(&project), nil
}

// Delete deletes a project
func (r *ProjectRepositoryAdapter) Delete(ctx context.Context, id string) error {
	err := r.queries.DeleteProject(ctx, id)
	if err != nil {
		return fmt.Errorf("failed to delete project: %w", err)
	}
	return nil
}

// Exists checks if a project exists by ID
func (r *ProjectRepositoryAdapter) Exists(ctx context.Context, id string) (bool, error) {
	exists, err := r.queries.ProjectExists(ctx, id)
	if err != nil {
		return false, fmt.Errorf("failed to check project existence: %w", err)
	}
	return exists, nil
}

// ExistsByAPIKey checks if a project exists by API key
func (r *ProjectRepositoryAdapter) ExistsByAPIKey(ctx context.Context, apiKey string) (bool, error) {
	exists, err := r.queries.ProjectExistsByAPIKey(ctx, apiKey)
	if err != nil {
		return false, fmt.Errorf("failed to check project existence by API key: %w", err)
	}
	return exists, nil
}

// ExistsByName checks if a project exists by name
func (r *ProjectRepositoryAdapter) ExistsByName(ctx context.Context, name string) (bool, error) {
	// Note: We don't have a specific query for this in SQLC yet
	// This is a simple implementation - could be optimized
	projects, err := r.queries.ListProjects(ctx)
	if err != nil {
		return false, fmt.Errorf("failed to check project existence by name: %w", err)
	}
	
	for _, project := range projects {
		if project.Name == name {
			return true, nil
		}
	}
	return false, nil
}

// Count returns the total number of projects
func (r *ProjectRepositoryAdapter) Count(ctx context.Context) (int64, error) {
	count, err := r.queries.CountProjects(ctx)
	if err != nil {
		return 0, fmt.Errorf("failed to count projects: %w", err)
	}
	return count, nil
}

// CountByOwner returns the number of projects owned by a user
func (r *ProjectRepositoryAdapter) CountByOwner(ctx context.Context, ownerUserID string) (int64, error) {
	count, err := r.queries.CountProjectsByOwner(ctx, ownerUserID)
	if err != nil {
		return 0, fmt.Errorf("failed to count projects by owner: %w", err)
	}
	return count, nil
}

// modelToOutbound converts SQLC model to outbound model
func (r *ProjectRepositoryAdapter) modelToOutbound(project *sqlc.Project) *outbound.Project {
	return &outbound.Project{
		ID:          project.ID,
		Name:        project.Name,
		APIKey:      project.ApiKey,
		OwnerUserID: project.OwnerUserID,
		CreatedAt:   project.CreatedAt.Time.Format("2006-01-02T15:04:05Z07:00"),
		UpdatedAt:   project.UpdatedAt.Time.Format("2006-01-02T15:04:05Z07:00"),
	}
}

