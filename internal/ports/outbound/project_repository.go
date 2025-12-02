package outbound

import (
	"context"
)

// Project represents a project entity
type Project struct {
	ID          string
	Name        string
	APIKey      string
	OwnerUserID string
	CreatedAt   string
	UpdatedAt   string
}

// CreateProjectParams holds parameters for creating a project
type CreateProjectParams struct {
	ID          string
	Name        string
	APIKey      string
	OwnerUserID string
}

// UpdateProjectParams holds parameters for updating a project
type UpdateProjectParams struct {
	ID     string
	Name   *string
	APIKey *string
}

// ProjectRepository defines the interface for project data access
type ProjectRepository interface {
	// Create creates a new project
	Create(ctx context.Context, params CreateProjectParams) (*Project, error)
	
	// GetByID retrieves a project by ID
	GetByID(ctx context.Context, id string) (*Project, error)
	
	// GetByAPIKey retrieves a project by API key
	GetByAPIKey(ctx context.Context, apiKey string) (*Project, error)
	
	// List retrieves all projects
	List(ctx context.Context) ([]*Project, error)
	
	// ListByOwner retrieves projects by owner user ID
	ListByOwner(ctx context.Context, ownerUserID string) ([]*Project, error)
	
	// Update updates a project
	Update(ctx context.Context, params UpdateProjectParams) (*Project, error)
	
	// Delete deletes a project
	Delete(ctx context.Context, id string) error
	
	// Exists checks if a project exists by ID
	Exists(ctx context.Context, id string) (bool, error)
	
	// ExistsByAPIKey checks if a project exists by API key
	ExistsByAPIKey(ctx context.Context, apiKey string) (bool, error)
	
	// ExistsByName checks if a project exists by name
	ExistsByName(ctx context.Context, name string) (bool, error)
	
	// Count returns the total number of projects
	Count(ctx context.Context) (int64, error)
	
	// CountByOwner returns the number of projects owned by a user
	CountByOwner(ctx context.Context, ownerUserID string) (int64, error)
}

