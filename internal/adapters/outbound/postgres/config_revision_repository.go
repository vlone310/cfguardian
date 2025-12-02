package postgres

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/vlone310/cfguardian/internal/adapters/outbound/postgres/sqlc"
	"github.com/vlone310/cfguardian/internal/ports/outbound"
)

// ConfigRevisionRepositoryAdapter implements outbound.ConfigRevisionRepository using PostgreSQL
type ConfigRevisionRepositoryAdapter struct {
	pool    *pgxpool.Pool
	queries *sqlc.Queries
}

// NewConfigRevisionRepositoryAdapter creates a new PostgreSQL config revision repository
func NewConfigRevisionRepositoryAdapter(pool *pgxpool.Pool) *ConfigRevisionRepositoryAdapter {
	return &ConfigRevisionRepositoryAdapter{
		pool:    pool,
		queries: sqlc.New(pool),
	}
}

// Create creates a new config revision (immutable)
func (r *ConfigRevisionRepositoryAdapter) Create(ctx context.Context, params outbound.CreateConfigRevisionParams) (*outbound.ConfigRevision, error) {
	revision, err := r.queries.CreateConfigRevision(ctx, sqlc.CreateConfigRevisionParams{
		ID:              params.ID,
		ProjectID:       params.ProjectID,
		ConfigKey:       params.ConfigKey,
		Version:         params.Version,
		Content:         params.Content,
		CreatedByUserID: params.CreatedByUserID,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to create config revision: %w", err)
	}
	
	return r.modelToOutbound(&revision), nil
}

// GetByID retrieves a revision by ID
func (r *ConfigRevisionRepositoryAdapter) GetByID(ctx context.Context, id string) (*outbound.ConfigRevision, error) {
	revision, err := r.queries.GetConfigRevision(ctx, id)
	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, fmt.Errorf("config revision not found")
		}
		return nil, fmt.Errorf("failed to get config revision: %w", err)
	}
	
	return r.modelToOutbound(&revision), nil
}

// GetByVersion retrieves a specific version of a config
func (r *ConfigRevisionRepositoryAdapter) GetByVersion(ctx context.Context, projectID, configKey string, version int64) (*outbound.ConfigRevision, error) {
	revision, err := r.queries.GetConfigRevisionByVersion(ctx, projectID, configKey, version)
	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, fmt.Errorf("config revision not found for version %d", version)
		}
		return nil, fmt.Errorf("failed to get config revision: %w", err)
	}
	
	return r.modelToOutbound(&revision), nil
}

// List retrieves all revisions for a config (ordered by version desc)
func (r *ConfigRevisionRepositoryAdapter) List(ctx context.Context, projectID, configKey string) ([]*outbound.ConfigRevision, error) {
	revisions, err := r.queries.ListConfigRevisions(ctx, projectID, configKey)
	if err != nil {
		return nil, fmt.Errorf("failed to list config revisions: %w", err)
	}
	
	return r.modelsToOutbound(revisions), nil
}

// ListPaginated retrieves revisions with pagination
func (r *ConfigRevisionRepositoryAdapter) ListPaginated(ctx context.Context, params outbound.ListRevisionsParams) ([]*outbound.ConfigRevision, error) {
	revisions, err := r.queries.ListConfigRevisionsPaginated(ctx, sqlc.ListConfigRevisionsPaginatedParams{
		ProjectID: params.ProjectID,
		ConfigKey: params.ConfigKey,
		Limit:     params.Limit,
		Offset:    params.Offset,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to list config revisions: %w", err)
	}
	
	return r.modelsToOutbound(revisions), nil
}

// ListAllByProject retrieves all revisions for a project
func (r *ConfigRevisionRepositoryAdapter) ListAllByProject(ctx context.Context, projectID string) ([]*outbound.ConfigRevision, error) {
	revisions, err := r.queries.ListAllRevisionsByProject(ctx, projectID)
	if err != nil {
		return nil, fmt.Errorf("failed to list all revisions by project: %w", err)
	}
	
	return r.modelsToOutbound(revisions), nil
}

// ListByUser retrieves revisions created by a specific user
func (r *ConfigRevisionRepositoryAdapter) ListByUser(ctx context.Context, userID string, limit int32) ([]*outbound.ConfigRevision, error) {
	revisions, err := r.queries.ListRevisionsByUser(ctx, userID, limit)
	if err != nil {
		return nil, fmt.Errorf("failed to list revisions by user: %w", err)
	}
	
	return r.modelsToOutbound(revisions), nil
}

// GetLatest retrieves the latest revision for a config
func (r *ConfigRevisionRepositoryAdapter) GetLatest(ctx context.Context, projectID, configKey string) (*outbound.ConfigRevision, error) {
	revision, err := r.queries.GetLatestRevision(ctx, projectID, configKey)
	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, fmt.Errorf("no revisions found")
		}
		return nil, fmt.Errorf("failed to get latest revision: %w", err)
	}
	
	return r.modelToOutbound(&revision), nil
}

// GetLatestN retrieves the N most recent revisions
func (r *ConfigRevisionRepositoryAdapter) GetLatestN(ctx context.Context, projectID, configKey string, n int32) ([]*outbound.ConfigRevision, error) {
	revisions, err := r.queries.GetLatestNRevisions(ctx, projectID, configKey, n)
	if err != nil {
		return nil, fmt.Errorf("failed to get latest revisions: %w", err)
	}
	
	return r.modelsToOutbound(revisions), nil
}

// GetHistory retrieves revision history with user emails
func (r *ConfigRevisionRepositoryAdapter) GetHistory(ctx context.Context, projectID, configKey string, limit int32) ([]*outbound.ConfigRevisionWithEmail, error) {
	historyRows, err := r.queries.GetRevisionHistory(ctx, projectID, configKey, limit)
	if err != nil {
		return nil, fmt.Errorf("failed to get revision history: %w", err)
	}
	
	result := make([]*outbound.ConfigRevisionWithEmail, len(historyRows))
	for i, row := range historyRows {
		result[i] = &outbound.ConfigRevisionWithEmail{
			ConfigRevision: outbound.ConfigRevision{
				ID:              row.ID,
				ProjectID:       row.ProjectID,
				ConfigKey:       row.ConfigKey,
				Version:         row.Version,
				Content:         json.RawMessage(row.Content),
				CreatedByUserID: row.CreatedByUserID,
				CreatedAt:       row.CreatedAt.Time.Format("2006-01-02T15:04:05Z07:00"),
			},
			CreatedByEmail: row.CreatedByEmail,
		}
	}
	
	return result, nil
}

// GetCreatedAfter retrieves revisions created after a specific time
func (r *ConfigRevisionRepositoryAdapter) GetCreatedAfter(ctx context.Context, projectID, configKey, afterTime string) ([]*outbound.ConfigRevision, error) {
	// Parse time string to pgtype.Timestamp
	// For now, simplified implementation
	// TODO: Proper time parsing
	var timestamp pgtype.Timestamp
	timestamp.Scan(afterTime)
	
	revisions, err := r.queries.GetRevisionsCreatedAfter(ctx, projectID, configKey, timestamp)
	if err != nil {
		return nil, fmt.Errorf("failed to get revisions created after: %w", err)
	}
	
	return r.modelsToOutbound(revisions), nil
}

// GetInVersionRange retrieves revisions within a version range
func (r *ConfigRevisionRepositoryAdapter) GetInVersionRange(ctx context.Context, params outbound.GetRevisionRangeParams) ([]*outbound.ConfigRevision, error) {
	revisions, err := r.queries.GetRevisionsInVersionRange(ctx, sqlc.GetRevisionsInVersionRangeParams{
		ProjectID: params.ProjectID,
		ConfigKey: params.ConfigKey,
		Version:   params.MinVersion,
		Version_2: params.MaxVersion,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to get revisions in version range: %w", err)
	}
	
	return r.modelsToOutbound(revisions), nil
}

// Count returns the total number of revisions for a config
func (r *ConfigRevisionRepositoryAdapter) Count(ctx context.Context, projectID, configKey string) (int64, error) {
	count, err := r.queries.CountRevisions(ctx, projectID, configKey)
	if err != nil {
		return 0, fmt.Errorf("failed to count revisions: %w", err)
	}
	return count, nil
}

// CountByProject returns the total number of revisions in a project
func (r *ConfigRevisionRepositoryAdapter) CountByProject(ctx context.Context, projectID string) (int64, error) {
	count, err := r.queries.CountRevisionsByProject(ctx, projectID)
	if err != nil {
		return 0, fmt.Errorf("failed to count revisions by project: %w", err)
	}
	return count, nil
}

// DeleteOld deletes old revisions (keep only versions >= minVersion)
func (r *ConfigRevisionRepositoryAdapter) DeleteOld(ctx context.Context, projectID, configKey string, minVersion int64) error {
	err := r.queries.DeleteOldRevisions(ctx, projectID, configKey, minVersion)
	if err != nil {
		return fmt.Errorf("failed to delete old revisions: %w", err)
	}
	return nil
}

// modelToOutbound converts SQLC model to outbound model
func (r *ConfigRevisionRepositoryAdapter) modelToOutbound(revision *sqlc.ConfigRevision) *outbound.ConfigRevision {
	return &outbound.ConfigRevision{
		ID:              revision.ID,
		ProjectID:       revision.ProjectID,
		ConfigKey:       revision.ConfigKey,
		Version:         revision.Version,
		Content:         json.RawMessage(revision.Content),
		CreatedByUserID: revision.CreatedByUserID,
		CreatedAt:       revision.CreatedAt.Time.Format("2006-01-02T15:04:05Z07:00"),
	}
}

// modelsToOutbound converts multiple SQLC models to outbound models
func (r *ConfigRevisionRepositoryAdapter) modelsToOutbound(revisions []sqlc.ConfigRevision) []*outbound.ConfigRevision {
	result := make([]*outbound.ConfigRevision, len(revisions))
	for i, rev := range revisions {
		result[i] = r.modelToOutbound(&rev)
	}
	return result
}

// min returns the minimum of two integers
func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

