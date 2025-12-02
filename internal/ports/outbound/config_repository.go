package outbound

import (
	"context"
	"encoding/json"
)

// Config represents a configuration entry with optimistic locking
type Config struct {
	ProjectID       string
	Key             string
	SchemaID        string
	Version         int64
	Content         json.RawMessage
	UpdatedByUserID string
	CreatedAt       string
	UpdatedAt       string
}

// CreateConfigParams holds parameters for creating a config
type CreateConfigParams struct {
	ProjectID       string
	Key             string
	SchemaID        string
	Content         json.RawMessage
	UpdatedByUserID string
}

// UpdateConfigParams holds parameters for updating a config with optimistic locking
type UpdateConfigParams struct {
	ProjectID       string
	Key             string
	ExpectedVersion int64 // For optimistic locking
	Content         json.RawMessage
	UpdatedByUserID string
}

// ChangeSchemaParams holds parameters for changing a config's schema
type ChangeSchemaParams struct {
	ProjectID       string
	Key             string
	SchemaID        string
	UpdatedByUserID string
}

// SearchConfigsParams holds parameters for searching configs
type SearchConfigsParams struct {
	ProjectID string
	KeyPattern string
	Limit      int32
}

// ConfigRepository defines the interface for config data access
// This repository handles the authoritative config state with optimistic locking
type ConfigRepository interface {
	// Create creates a new config (version starts at 1)
	Create(ctx context.Context, params CreateConfigParams) (*Config, error)
	
	// Get retrieves a config by project ID and key
	Get(ctx context.Context, projectID, key string) (*Config, error)
	
	// GetWithVersion retrieves a config only if it matches the expected version
	GetWithVersion(ctx context.Context, projectID, key string, version int64) (*Config, error)
	
	// ListByProject retrieves all configs for a project
	ListByProject(ctx context.Context, projectID string) ([]*Config, error)
	
	// ListBySchema retrieves all configs using a specific schema
	ListBySchema(ctx context.Context, schemaID string) ([]*Config, error)
	
	// Update updates a config with optimistic locking
	// Returns error if version mismatch (concurrent modification detected)
	Update(ctx context.Context, params UpdateConfigParams) (*Config, error)
	
	// ChangeSchema changes the schema ID for a config
	ChangeSchema(ctx context.Context, params ChangeSchemaParams) (*Config, error)
	
	// Delete deletes a config
	Delete(ctx context.Context, projectID, key string) error
	
	// Exists checks if a config exists
	Exists(ctx context.Context, projectID, key string) (bool, error)
	
	// GetVersion retrieves the current version of a config
	GetVersion(ctx context.Context, projectID, key string) (int64, error)
	
	// LockForUpdate locks a config for update and returns its current version
	// Use within a transaction for pessimistic locking
	LockForUpdate(ctx context.Context, projectID, key string) (int64, error)
	
	// Search searches configs by key pattern
	Search(ctx context.Context, params SearchConfigsParams) ([]*Config, error)
	
	// GetUpdatedAfter retrieves configs updated after a specific time
	GetUpdatedAfter(ctx context.Context, projectID string, afterTime string) ([]*Config, error)
	
	// GetUpdatedByUser retrieves configs updated by a specific user
	GetUpdatedByUser(ctx context.Context, userID string, limit int32) ([]*Config, error)
	
	// CountByProject returns the number of configs in a project
	CountByProject(ctx context.Context, projectID string) (int64, error)
	
	// CountBySchema returns the number of configs using a schema
	CountBySchema(ctx context.Context, schemaID string) (int64, error)
}

