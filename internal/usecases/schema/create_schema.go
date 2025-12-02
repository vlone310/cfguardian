package schema

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"github.com/vlone310/cfguardian/internal/domain/entities"
	"github.com/vlone310/cfguardian/internal/domain/services"
	"github.com/vlone310/cfguardian/internal/ports/outbound"
)

// CreateSchemaRequest holds schema creation data
type CreateSchemaRequest struct {
	Name          string
	SchemaContent string
	CreatedByUserID string
}

// CreateSchemaResponse holds created schema data
type CreateSchemaResponse struct {
	ID              string
	Name            string
	SchemaContent   string
	CreatedByUserID string
	CreatedAt       string
}

// CreateSchemaUseCase handles config schema creation
type CreateSchemaUseCase struct {
	schemaRepo      outbound.ConfigSchemaRepository
	schemaValidator *services.SchemaValidator
}

// NewCreateSchemaUseCase creates a new CreateSchemaUseCase
func NewCreateSchemaUseCase(
	schemaRepo outbound.ConfigSchemaRepository,
	schemaValidator *services.SchemaValidator,
) *CreateSchemaUseCase {
	return &CreateSchemaUseCase{
		schemaRepo:      schemaRepo,
		schemaValidator: schemaValidator,
	}
}

// Execute creates a new config schema
func (uc *CreateSchemaUseCase) Execute(ctx context.Context, req CreateSchemaRequest) (*CreateSchemaResponse, error) {
	// Validate input
	if req.Name == "" {
		return nil, fmt.Errorf("schema name is required")
	}
	if req.SchemaContent == "" {
		return nil, fmt.Errorf("schema content is required")
	}
	if req.CreatedByUserID == "" {
		return nil, fmt.Errorf("creator user ID is required")
	}
	
	// Validate that the schema content is a valid JSON Schema
	if err := uc.schemaValidator.ValidateSchema(req.SchemaContent); err != nil {
		return nil, fmt.Errorf("invalid JSON Schema: %w", err)
	}
	
	// Check if schema name already exists
	nameExists, err := uc.schemaRepo.ExistsByName(ctx, req.Name)
	if err != nil {
		return nil, fmt.Errorf("failed to check if schema name exists: %w", err)
	}
	if nameExists {
		return nil, fmt.Errorf("schema with name '%s' already exists", req.Name)
	}
	
	// Create domain entity
	schemaID := uuid.New().String()
	schemaEntity, err := entities.NewConfigSchema(
		schemaID,
		req.Name,
		req.SchemaContent,
		req.CreatedByUserID,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create schema entity: %w", err)
	}
	
	// Persist to repository
	schema, err := uc.schemaRepo.Create(ctx, outbound.CreateConfigSchemaParams{
		ID:              schemaEntity.ID(),
		Name:            schemaEntity.Name(),
		SchemaContent:   schemaEntity.SchemaContent(),
		CreatedByUserID: schemaEntity.CreatedByUserID(),
	})
	if err != nil {
		return nil, fmt.Errorf("failed to create schema: %w", err)
	}
	
	return &CreateSchemaResponse{
		ID:              schema.ID,
		Name:            schema.Name,
		SchemaContent:   schema.SchemaContent,
		CreatedByUserID: schema.CreatedByUserID,
		CreatedAt:       schema.CreatedAt,
	}, nil
}

