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

// UpdateConfigRequest holds config update data with optimistic locking
type UpdateConfigRequest struct {
	ProjectID       string
	Key             string
	ExpectedVersion int64           // For optimistic locking
	Content         json.RawMessage
	UpdatedByUserID string
}

// UpdateConfigResponse holds updated config data
type UpdateConfigResponse struct {
	ProjectID       string
	Key             string
	SchemaID        string
	Version         int64
	Content         json.RawMessage
	UpdatedByUserID string
	UpdatedAt       string
}

// UpdateConfigUseCase handles config updates with optimistic locking and validation
type UpdateConfigUseCase struct {
	configRepo      outbound.ConfigRepository
	revisionRepo    outbound.ConfigRevisionRepository
	schemaRepo      outbound.ConfigSchemaRepository
	schemaValidator *services.SchemaValidator
	versionManager  *services.VersionManager
}

// NewUpdateConfigUseCase creates a new UpdateConfigUseCase
func NewUpdateConfigUseCase(
	configRepo outbound.ConfigRepository,
	revisionRepo outbound.ConfigRevisionRepository,
	schemaRepo outbound.ConfigSchemaRepository,
	schemaValidator *services.SchemaValidator,
	versionManager *services.VersionManager,
) *UpdateConfigUseCase {
	return &UpdateConfigUseCase{
		configRepo:      configRepo,
		revisionRepo:    revisionRepo,
		schemaRepo:      schemaRepo,
		schemaValidator: schemaValidator,
		versionManager:  versionManager,
	}
}

// Execute updates a config with optimistic locking
func (uc *UpdateConfigUseCase) Execute(ctx context.Context, req UpdateConfigRequest) (*UpdateConfigResponse, error) {
	// Validate input
	if req.ProjectID == "" {
		return nil, fmt.Errorf("project ID is required")
	}
	if req.Key == "" {
		return nil, fmt.Errorf("config key is required")
	}
	if req.ExpectedVersion < 1 {
		return nil, fmt.Errorf("expected version must be >= 1")
	}
	if len(req.Content) == 0 {
		return nil, fmt.Errorf("config content is required")
	}
	if req.UpdatedByUserID == "" {
		return nil, fmt.Errorf("user ID is required")
	}
	
	// Get current config
	currentConfig, err := uc.configRepo.Get(ctx, req.ProjectID, req.Key)
	if err != nil {
		return nil, fmt.Errorf("config not found: %w", err)
	}
	
	// Create version value objects for validation
	expectedVersion, err := valueobjects.NewVersion(req.ExpectedVersion)
	if err != nil {
		return nil, fmt.Errorf("invalid expected version: %w", err)
	}
	
	currentVersion, err := valueobjects.NewVersion(currentConfig.Version)
	if err != nil {
		return nil, fmt.Errorf("invalid current version: %w", err)
	}
	
	// Validate optimistic lock - this is the key to preventing concurrent modifications!
	if err := uc.versionManager.ValidateUpdate(expectedVersion, currentVersion, req.Key); err != nil {
		return nil, err // Returns VersionConflictError
	}
	
	// Get schema for validation
	schema, err := uc.schemaRepo.GetByID(ctx, currentConfig.SchemaID)
	if err != nil {
		return nil, fmt.Errorf("schema not found: %w", err)
	}
	
	// Validate new content against schema
	if err := uc.schemaValidator.ValidateOrError(schema.SchemaContent, req.Content); err != nil {
		return nil, fmt.Errorf("content validation failed: %w", err)
	}
	
	// Update config in repository (version will be incremented)
	updatedConfig, err := uc.configRepo.Update(ctx, outbound.UpdateConfigParams{
		ProjectID:       req.ProjectID,
		Key:             req.Key,
		ExpectedVersion: req.ExpectedVersion,
		Content:         req.Content,
		UpdatedByUserID: req.UpdatedByUserID,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to update config: %w", err)
	}
	
	// Create revision for audit log
	revisionID := uuid.New().String()
	newVersion, _ := valueobjects.NewVersion(updatedConfig.Version)
	revisionEntity := entities.NewConfigRevision(
		revisionID,
		updatedConfig.ProjectID,
		updatedConfig.Key,
		newVersion,
		updatedConfig.Content,
		updatedConfig.UpdatedByUserID,
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
		// Log error but don't fail update
		fmt.Printf("Warning: failed to create config revision: %v\n", err)
	}
	
	// TODO: Publish ConfigUpdated event
	_ = events.NewConfigUpdated(
		uuid.New().String(),
		updatedConfig.ProjectID,
		updatedConfig.Key,
		updatedConfig.SchemaID,
		currentVersion,
		newVersion,
		updatedConfig.Content,
		updatedConfig.UpdatedByUserID,
	)
	
	return &UpdateConfigResponse{
		ProjectID:       updatedConfig.ProjectID,
		Key:             updatedConfig.Key,
		SchemaID:        updatedConfig.SchemaID,
		Version:         updatedConfig.Version,
		Content:         updatedConfig.Content,
		UpdatedByUserID: updatedConfig.UpdatedByUserID,
		UpdatedAt:       updatedConfig.UpdatedAt,
	}, nil
}

