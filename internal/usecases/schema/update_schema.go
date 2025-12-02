package schema

import (
	"context"
	"fmt"

	"github.com/vlone310/cfguardian/internal/domain/services"
	"github.com/vlone310/cfguardian/internal/ports/outbound"
)

// UpdateSchemaRequest holds schema update data
type UpdateSchemaRequest struct {
	SchemaID      string  `json:"schema_id"`
	Name          *string `json:"name,omitempty"`
	SchemaContent *string `json:"schema_content,omitempty"`
}

// UpdateSchemaResponse holds updated schema data
type UpdateSchemaResponse struct {
	ID            string `json:"id"`
	Name          string `json:"name"`
	SchemaContent string `json:"schema_content"`
	UpdatedAt     string `json:"updated_at"`
}

// UpdateSchemaUseCase handles config schema updates
type UpdateSchemaUseCase struct {
	schemaRepo      outbound.ConfigSchemaRepository
	schemaValidator *services.SchemaValidator
}

// NewUpdateSchemaUseCase creates a new UpdateSchemaUseCase
func NewUpdateSchemaUseCase(
	schemaRepo outbound.ConfigSchemaRepository,
	schemaValidator *services.SchemaValidator,
) *UpdateSchemaUseCase {
	return &UpdateSchemaUseCase{
		schemaRepo:      schemaRepo,
		schemaValidator: schemaValidator,
	}
}

// Execute updates a config schema
func (uc *UpdateSchemaUseCase) Execute(ctx context.Context, req UpdateSchemaRequest) (*UpdateSchemaResponse, error) {
	// Validate input
	if req.SchemaID == "" {
		return nil, fmt.Errorf("schema ID is required")
	}
	
	// Check if schema exists
	exists, err := uc.schemaRepo.Exists(ctx, req.SchemaID)
	if err != nil {
		return nil, fmt.Errorf("failed to check if schema exists: %w", err)
	}
	if !exists {
		return nil, fmt.Errorf("schema not found")
	}
	
	// Validate schema content if provided
	if req.SchemaContent != nil && *req.SchemaContent != "" {
		if err := uc.schemaValidator.ValidateSchema(*req.SchemaContent); err != nil {
			return nil, fmt.Errorf("invalid JSON Schema: %w", err)
		}
	}
	
	// Check if new name already exists (if name is being changed)
	if req.Name != nil && *req.Name != "" {
		nameExists, err := uc.schemaRepo.ExistsByName(ctx, *req.Name)
		if err != nil {
			return nil, fmt.Errorf("failed to check if schema name exists: %w", err)
		}
		if nameExists {
			// Check if it's the same schema (allowed to keep same name)
			existing, _ := uc.schemaRepo.GetByName(ctx, *req.Name)
			if existing != nil && existing.ID != req.SchemaID {
				return nil, fmt.Errorf("schema with name '%s' already exists", *req.Name)
			}
		}
	}
	
	// Update schema
	schema, err := uc.schemaRepo.Update(ctx, outbound.UpdateConfigSchemaParams{
		ID:            req.SchemaID,
		Name:          req.Name,
		SchemaContent: req.SchemaContent,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to update schema: %w", err)
	}
	
	return &UpdateSchemaResponse{
		ID:            schema.ID,
		Name:          schema.Name,
		SchemaContent: schema.SchemaContent,
		UpdatedAt:     schema.UpdatedAt,
	}, nil
}

