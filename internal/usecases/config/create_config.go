package config

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/google/uuid"
	"github.com/vlone310/cfguardian/internal/domain/entities"
	"github.com/vlone310/cfguardian/internal/domain/events"
	"github.com/vlone310/cfguardian/internal/domain/services"
	"github.com/vlone310/cfguardian/internal/domain/valueobjects"
	"github.com/vlone310/cfguardian/internal/ports/outbound"
)

// CreateConfigRequest holds config creation data
type CreateConfigRequest struct {
	ProjectID       string
	Key             string
	SchemaID        string
	Content         json.RawMessage
	UpdatedByUserID string
}

// CreateConfigResponse holds created config data
type CreateConfigResponse struct {
	ProjectID       string
	Key             string
	SchemaID        string
	Version         int64
	Content         json.RawMessage
	UpdatedByUserID string
	CreatedAt       string
}

// CreateConfigUseCase handles config creation
type CreateConfigUseCase struct {
	configRepo      outbound.ConfigRepository
	revisionRepo    outbound.ConfigRevisionRepository
	schemaRepo      outbound.ConfigSchemaRepository
	projectRepo     outbound.ProjectRepository
	schemaValidator *services.SchemaValidator
}

// NewCreateConfigUseCase creates a new CreateConfigUseCase
func NewCreateConfigUseCase(
	configRepo outbound.ConfigRepository,
	revisionRepo outbound.ConfigRevisionRepository,
	schemaRepo outbound.ConfigSchemaRepository,
	projectRepo outbound.ProjectRepository,
	schemaValidator *services.SchemaValidator,
) *CreateConfigUseCase {
	return &CreateConfigUseCase{
		configRepo:      configRepo,
		revisionRepo:    revisionRepo,
		schemaRepo:      schemaRepo,
		projectRepo:     projectRepo,
		schemaValidator: schemaValidator,
	}
}

// Execute creates a new config with schema validation
func (uc *CreateConfigUseCase) Execute(ctx context.Context, req CreateConfigRequest) (*CreateConfigResponse, error) {
	// Validate input
	if req.ProjectID == "" {
		return nil, fmt.Errorf("project ID is required")
	}
	if req.Key == "" {
		return nil, fmt.Errorf("config key is required")
	}
	if req.SchemaID == "" {
		return nil, fmt.Errorf("schema ID is required")
	}
	if len(req.Content) == 0 {
		return nil, fmt.Errorf("config content is required")
	}
	if req.UpdatedByUserID == "" {
		return nil, fmt.Errorf("user ID is required")
	}
	
	// Verify project exists
	projectExists, err := uc.projectRepo.Exists(ctx, req.ProjectID)
	if err != nil {
		return nil, fmt.Errorf("failed to verify project exists: %w", err)
	}
	if !projectExists {
		return nil, fmt.Errorf("project not found")
	}
	
	// Check if config already exists
	configExists, err := uc.configRepo.Exists(ctx, req.ProjectID, req.Key)
	if err != nil {
		return nil, fmt.Errorf("failed to check if config exists: %w", err)
	}
	if configExists {
		return nil, fmt.Errorf("config with key '%s' already exists in project", req.Key)
	}
	
	// Get schema for validation
	schema, err := uc.schemaRepo.GetByID(ctx, req.SchemaID)
	if err != nil {
		return nil, fmt.Errorf("schema not found: %w", err)
	}
	
	// Validate content against schema
	if err := uc.schemaValidator.ValidateOrError(schema.SchemaContent, req.Content); err != nil {
		return nil, fmt.Errorf("content validation failed: %w", err)
	}
	
	// Create domain entity
	configEntity := entities.NewConfig(
		req.ProjectID,
		req.Key,
		req.SchemaID,
		req.Content,
		req.UpdatedByUserID,
	)
	
	// Persist to repository
	config, err := uc.configRepo.Create(ctx, outbound.CreateConfigParams{
		ProjectID:       configEntity.ProjectID(),
		Key:             configEntity.Key(),
		SchemaID:        configEntity.SchemaID(),
		Content:         configEntity.Content(),
		UpdatedByUserID: configEntity.UpdatedByUserID(),
	})
	if err != nil {
		return nil, fmt.Errorf("failed to create config: %w", err)
	}
	
	// Create initial revision
	revisionID := uuid.New().String()
	revisionEntity := entities.NewConfigRevision(
		revisionID,
		config.ProjectID,
		config.Key,
		valueobjects.MustNewVersion(config.Version),
		config.Content,
		config.UpdatedByUserID,
	)
	
	_, err = uc.revisionRepo.Create(ctx, outbound.CreateConfigRevisionParams{
		ID:              revisionEntity.ID(),
		ProjectID:       revisionEntity.ProjectID(),
		ConfigKey:       revisionEntity.ConfigKey(),
		Version:         revisionEntity.Version().Value(),
		Content:         revisionEntity.Content(),
		CreatedByUserID: revisionEntity.CreatedByUserID(),
	})
	if err != nil {
		// Log error but don't fail config creation
		// Revision is for audit, config is the source of truth
		fmt.Printf("Warning: failed to create config revision: %v\n", err)
	}
	
	// TODO: Publish ConfigCreated event
	_ = events.NewConfigCreated(
		uuid.New().String(),
		config.ProjectID,
		config.Key,
		config.SchemaID,
		valueobjects.MustNewVersion(config.Version),
		config.Content,
		config.UpdatedByUserID,
	)
	
	return &CreateConfigResponse{
		ProjectID:       config.ProjectID,
		Key:             config.Key,
		SchemaID:        config.SchemaID,
		Version:         config.Version,
		Content:         config.Content,
		UpdatedByUserID: config.UpdatedByUserID,
		CreatedAt:       config.CreatedAt,
	}, nil
}

