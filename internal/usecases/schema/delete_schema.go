package schema

import (
	"context"
	"fmt"

	"github.com/vlone310/cfguardian/internal/ports/outbound"
)

// DeleteSchemaRequest holds schema deletion data
type DeleteSchemaRequest struct {
	SchemaID string
}

// DeleteSchemaUseCase handles config schema deletion
type DeleteSchemaUseCase struct {
	schemaRepo outbound.ConfigSchemaRepository
}

// NewDeleteSchemaUseCase creates a new DeleteSchemaUseCase
func NewDeleteSchemaUseCase(schemaRepo outbound.ConfigSchemaRepository) *DeleteSchemaUseCase {
	return &DeleteSchemaUseCase{
		schemaRepo: schemaRepo,
	}
}

// Execute deletes a config schema
func (uc *DeleteSchemaUseCase) Execute(ctx context.Context, req DeleteSchemaRequest) error {
	// Validate input
	if req.SchemaID == "" {
		return fmt.Errorf("schema ID is required")
	}
	
	// Check if schema exists
	exists, err := uc.schemaRepo.Exists(ctx, req.SchemaID)
	if err != nil {
		return fmt.Errorf("failed to check if schema exists: %w", err)
	}
	if !exists {
		return fmt.Errorf("schema not found")
	}
	
	// Check if any configs are using this schema
	configsUsing, err := uc.schemaRepo.CountConfigsUsing(ctx, req.SchemaID)
	if err != nil {
		return fmt.Errorf("failed to check configs using schema: %w", err)
	}
	if configsUsing > 0 {
		return fmt.Errorf("cannot delete schema: %d config(s) are still using it", configsUsing)
	}
	
	// Delete schema
	// Note: Foreign key RESTRICT on configs.schema_id will prevent deletion if configs exist
	if err := uc.schemaRepo.Delete(ctx, req.SchemaID); err != nil {
		return fmt.Errorf("failed to delete schema: %w", err)
	}
	
	return nil
}

