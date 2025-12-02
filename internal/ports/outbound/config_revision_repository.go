package outbound

import (
	"context"
	"encoding/json"
)

// ConfigRevision represents an immutable historical record of a config
type ConfigRevision struct {
	ID              string
	ProjectID       string
	ConfigKey       string
	Version         int64
	Content         json.RawMessage
	CreatedByUserID string
	CreatedAt       string
}

// ConfigRevisionWithEmail includes the creator's email for display
type ConfigRevisionWithEmail struct {
	ConfigRevision
	CreatedByEmail string
}

// CreateConfigRevisionParams holds parameters for creating a config revision
type CreateConfigRevisionParams struct {
	ID              string
	ProjectID       string
	ConfigKey       string
	Version         int64
	Content         json.RawMessage
	CreatedByUserID string
}

// ListRevisionsParams holds parameters for listing revisions with pagination
type ListRevisionsParams struct {
	ProjectID string
	ConfigKey string
	Limit     int32
	Offset    int32
}

// GetRevisionRangeParams holds parameters for getting revisions in a version range
type GetRevisionRangeParams struct {
	ProjectID     string
	ConfigKey     string
	MinVersion    int64
	MaxVersion    int64
}

// ConfigRevisionRepository defines the interface for config revision data access
// This repository handles the immutable audit log of configuration changes
type ConfigRevisionRepository interface {
	// Create creates a new config revision (immutable)
	Create(ctx context.Context, params CreateConfigRevisionParams) (*ConfigRevision, error)
	
	// GetByID retrieves a revision by ID
	GetByID(ctx context.Context, id string) (*ConfigRevision, error)
	
	// GetByVersion retrieves a specific version of a config
	GetByVersion(ctx context.Context, projectID, configKey string, version int64) (*ConfigRevision, error)
	
	// List retrieves all revisions for a config (ordered by version desc)
	List(ctx context.Context, projectID, configKey string) ([]*ConfigRevision, error)
	
	// ListPaginated retrieves revisions with pagination
	ListPaginated(ctx context.Context, params ListRevisionsParams) ([]*ConfigRevision, error)
	
	// ListAllByProject retrieves all revisions for a project
	ListAllByProject(ctx context.Context, projectID string) ([]*ConfigRevision, error)
	
	// ListByUser retrieves revisions created by a specific user
	ListByUser(ctx context.Context, userID string, limit int32) ([]*ConfigRevision, error)
	
	// GetLatest retrieves the latest revision for a config
	GetLatest(ctx context.Context, projectID, configKey string) (*ConfigRevision, error)
	
	// GetLatestN retrieves the N most recent revisions
	GetLatestN(ctx context.Context, projectID, configKey string, n int32) ([]*ConfigRevision, error)
	
	// GetHistory retrieves revision history with user emails
	GetHistory(ctx context.Context, projectID, configKey string, limit int32) ([]*ConfigRevisionWithEmail, error)
	
	// GetCreatedAfter retrieves revisions created after a specific time
	GetCreatedAfter(ctx context.Context, projectID, configKey, afterTime string) ([]*ConfigRevision, error)
	
	// GetInVersionRange retrieves revisions within a version range
	GetInVersionRange(ctx context.Context, params GetRevisionRangeParams) ([]*ConfigRevision, error)
	
	// Count returns the total number of revisions for a config
	Count(ctx context.Context, projectID, configKey string) (int64, error)
	
	// CountByProject returns the total number of revisions in a project
	CountByProject(ctx context.Context, projectID string) (int64, error)
	
	// DeleteOld deletes old revisions (keep only versions >= minVersion)
	// Use carefully - this is for cleanup/retention policies
	DeleteOld(ctx context.Context, projectID, configKey string, minVersion int64) error
}

