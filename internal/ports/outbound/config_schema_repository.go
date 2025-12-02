package outbound

import (
	"context"
)

// ConfigSchema represents a reusable JSON Schema definition
type ConfigSchema struct {
	ID              string
	Name            string
	SchemaContent   string
	CreatedByUserID string
	CreatedAt       string
	UpdatedAt       string
}

// CreateConfigSchemaParams holds parameters for creating a config schema
type CreateConfigSchemaParams struct {
	ID              string
	Name            string
	SchemaContent   string
	CreatedByUserID string
}

// UpdateConfigSchemaParams holds parameters for updating a config schema
type UpdateConfigSchemaParams struct {
	ID            string
	Name          *string
	SchemaContent *string
}

// ConfigSchemaRepository defines the interface for config schema data access
type ConfigSchemaRepository interface {
	// Create creates a new config schema
	Create(ctx context.Context, params CreateConfigSchemaParams) (*ConfigSchema, error)
	
	// GetByID retrieves a config schema by ID
	GetByID(ctx context.Context, id string) (*ConfigSchema, error)
	
	// GetByName retrieves a config schema by name
	GetByName(ctx context.Context, name string) (*ConfigSchema, error)
	
	// List retrieves all config schemas
	List(ctx context.Context) ([]*ConfigSchema, error)
	
	// ListByCreator retrieves config schemas created by a specific user
	ListByCreator(ctx context.Context, creatorUserID string) ([]*ConfigSchema, error)
	
	// Update updates a config schema
	Update(ctx context.Context, params UpdateConfigSchemaParams) (*ConfigSchema, error)
	
	// Delete deletes a config schema
	// This should fail if any configs are using this schema (foreign key constraint)
	Delete(ctx context.Context, id string) error
	
	// Exists checks if a config schema exists by ID
	Exists(ctx context.Context, id string) (bool, error)
	
	// ExistsByName checks if a config schema exists by name
	ExistsByName(ctx context.Context, name string) (bool, error)
	
	// Count returns the total number of config schemas
	Count(ctx context.Context) (int64, error)
	
	// CountConfigsUsing returns the number of configs using this schema
	CountConfigsUsing(ctx context.Context, schemaID string) (int64, error)
}

