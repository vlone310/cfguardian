package schema

import (
	"context"
	"fmt"

	"github.com/vlone310/cfguardian/internal/ports/outbound"
)

// SchemaListItem represents a schema in the list
type SchemaListItem struct {
	ID              string
	Name            string
	CreatedByUserID string
	CreatedAt       string
	UpdatedAt       string
	ConfigsUsing    int64 // Number of configs using this schema
}

// ListSchemasResponse holds the list of schemas
type ListSchemasResponse struct {
	Schemas []*SchemaListItem
	Total   int64
}

// ListSchemasUseCase handles listing config schemas
type ListSchemasUseCase struct {
	schemaRepo outbound.ConfigSchemaRepository
}

// NewListSchemasUseCase creates a new ListSchemasUseCase
func NewListSchemasUseCase(schemaRepo outbound.ConfigSchemaRepository) *ListSchemasUseCase {
	return &ListSchemasUseCase{
		schemaRepo: schemaRepo,
	}
}

// Execute retrieves all config schemas
func (uc *ListSchemasUseCase) Execute(ctx context.Context) (*ListSchemasResponse, error) {
	// Get all schemas
	schemas, err := uc.schemaRepo.List(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to list schemas: %w", err)
	}
	
	// Get total count
	count, err := uc.schemaRepo.Count(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to count schemas: %w", err)
	}
	
	// Convert to response format (include usage count for each)
	items := make([]*SchemaListItem, len(schemas))
	for i, schema := range schemas {
		// Get count of configs using this schema
		configsUsing, err := uc.schemaRepo.CountConfigsUsing(ctx, schema.ID)
		if err != nil {
			// Don't fail the entire request, just set to 0
			configsUsing = 0
		}
		
		items[i] = &SchemaListItem{
			ID:              schema.ID,
			Name:            schema.Name,
			CreatedByUserID: schema.CreatedByUserID,
			CreatedAt:       schema.CreatedAt,
			UpdatedAt:       schema.UpdatedAt,
			ConfigsUsing:    configsUsing,
		}
	}
	
	return &ListSchemasResponse{
		Schemas: items,
		Total:   count,
	}, nil
}

