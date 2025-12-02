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

// ConfigSchemaRepositoryAdapter implements outbound.ConfigSchemaRepository using PostgreSQL
type ConfigSchemaRepositoryAdapter struct {
	pool    *pgxpool.Pool
	queries *sqlc.Queries
}

// NewConfigSchemaRepositoryAdapter creates a new PostgreSQL config schema repository
func NewConfigSchemaRepositoryAdapter(pool *pgxpool.Pool) *ConfigSchemaRepositoryAdapter {
	return &ConfigSchemaRepositoryAdapter{
		pool:    pool,
		queries: sqlc.New(pool),
	}
}

// Create creates a new config schema
func (r *ConfigSchemaRepositoryAdapter) Create(ctx context.Context, params outbound.CreateConfigSchemaParams) (*outbound.ConfigSchema, error) {
	schema, err := r.queries.CreateConfigSchema(ctx, sqlc.CreateConfigSchemaParams{
		ID:              params.ID,
		Name:            params.Name,
		SchemaContent:   params.SchemaContent,
		CreatedByUserID: params.CreatedByUserID,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to create config schema: %w", err)
	}
	
	return r.modelToOutbound(&schema), nil
}

// GetByID retrieves a config schema by ID
func (r *ConfigSchemaRepositoryAdapter) GetByID(ctx context.Context, id string) (*outbound.ConfigSchema, error) {
	schema, err := r.queries.GetConfigSchemaByID(ctx, id)
	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, fmt.Errorf("config schema not found")
		}
		return nil, fmt.Errorf("failed to get config schema: %w", err)
	}
	
	return r.modelToOutbound(&schema), nil
}

// GetByName retrieves a config schema by name
func (r *ConfigSchemaRepositoryAdapter) GetByName(ctx context.Context, name string) (*outbound.ConfigSchema, error) {
	schema, err := r.queries.GetConfigSchemaByName(ctx, name)
	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, fmt.Errorf("config schema not found")
		}
		return nil, fmt.Errorf("failed to get config schema: %w", err)
	}
	
	return r.modelToOutbound(&schema), nil
}

// List retrieves all config schemas
func (r *ConfigSchemaRepositoryAdapter) List(ctx context.Context) ([]*outbound.ConfigSchema, error) {
	schemas, err := r.queries.ListConfigSchemas(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to list config schemas: %w", err)
	}
	
	result := make([]*outbound.ConfigSchema, len(schemas))
	for i, schema := range schemas {
		result[i] = r.modelToOutbound(&schema)
	}
	
	return result, nil
}

// ListByCreator retrieves config schemas created by a specific user
func (r *ConfigSchemaRepositoryAdapter) ListByCreator(ctx context.Context, creatorUserID string) ([]*outbound.ConfigSchema, error) {
	schemas, err := r.queries.ListConfigSchemasByCreator(ctx, creatorUserID)
	if err != nil {
		return nil, fmt.Errorf("failed to list config schemas by creator: %w", err)
	}
	
	result := make([]*outbound.ConfigSchema, len(schemas))
	for i, schema := range schemas {
		result[i] = r.modelToOutbound(&schema)
	}
	
	return result, nil
}

// Update updates a config schema
func (r *ConfigSchemaRepositoryAdapter) Update(ctx context.Context, params outbound.UpdateConfigSchemaParams) (*outbound.ConfigSchema, error) {
	var name pgtype.Text
	var schemaContent pgtype.Text
	
	// Only update fields that are provided
	if params.Name != nil {
		name = pgtype.Text{String: *params.Name, Valid: true}
	}
	if params.SchemaContent != nil {
		schemaContent = pgtype.Text{String: *params.SchemaContent, Valid: true}
	}
	
	schema, err := r.queries.UpdateConfigSchema(ctx, params.ID, name, schemaContent)
	if err != nil {
		return nil, fmt.Errorf("failed to update config schema: %w", err)
	}
	
	return r.modelToOutbound(&schema), nil
}

// Delete deletes a config schema
func (r *ConfigSchemaRepositoryAdapter) Delete(ctx context.Context, id string) error {
	err := r.queries.DeleteConfigSchema(ctx, id)
	if err != nil {
		return fmt.Errorf("failed to delete config schema: %w", err)
	}
	return nil
}

// Exists checks if a config schema exists by ID
func (r *ConfigSchemaRepositoryAdapter) Exists(ctx context.Context, id string) (bool, error) {
	exists, err := r.queries.ConfigSchemaExists(ctx, id)
	if err != nil {
		return false, fmt.Errorf("failed to check config schema existence: %w", err)
	}
	return exists, nil
}

// ExistsByName checks if a config schema exists by name
func (r *ConfigSchemaRepositoryAdapter) ExistsByName(ctx context.Context, name string) (bool, error) {
	exists, err := r.queries.ConfigSchemaExistsByName(ctx, name)
	if err != nil {
		return false, fmt.Errorf("failed to check config schema existence by name: %w", err)
	}
	return exists, nil
}

// Count returns the total number of config schemas
func (r *ConfigSchemaRepositoryAdapter) Count(ctx context.Context) (int64, error) {
	count, err := r.queries.CountConfigSchemas(ctx)
	if err != nil {
		return 0, fmt.Errorf("failed to count config schemas: %w", err)
	}
	return count, nil
}

// CountConfigsUsing returns the number of configs using this schema
func (r *ConfigSchemaRepositoryAdapter) CountConfigsUsing(ctx context.Context, schemaID string) (int64, error) {
	count, err := r.queries.CountConfigsUsingSchema(ctx, schemaID)
	if err != nil {
		return 0, fmt.Errorf("failed to count configs using schema: %w", err)
	}
	return count, nil
}

// modelToOutbound converts SQLC model to outbound model
func (r *ConfigSchemaRepositoryAdapter) modelToOutbound(schema *sqlc.ConfigSchema) *outbound.ConfigSchema {
	return &outbound.ConfigSchema{
		ID:              schema.ID,
		Name:            schema.Name,
		SchemaContent:   schema.SchemaContent,
		CreatedByUserID: schema.CreatedByUserID,
		CreatedAt:       schema.CreatedAt.Time.Format("2006-01-02T15:04:05Z07:00"),
		UpdatedAt:       schema.UpdatedAt.Time.Format("2006-01-02T15:04:05Z07:00"),
	}
}

