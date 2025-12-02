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

// RollbackConfigRequest holds config rollback data
type RollbackConfigRequest struct {
	ProjectID         string
	Key               string
	TargetVersion     int64
	ExpectedVersion   int64 // Current version for optimistic locking
	RolledBackByUserID string
}

// RollbackConfigResponse holds rollback result
type RollbackConfigResponse struct {
	ProjectID       string
	Key             string
	Version         int64 // New version after rollback
	Content         json.RawMessage
	UpdatedByUserID string
	UpdatedAt       string
}

// RollbackConfigUseCase handles rolling back a config to a previous version
type RollbackConfigUseCase struct {
	configRepo      outbound.ConfigRepository
	revisionRepo    outbound.ConfigRevisionRepository
	schemaRepo      outbound.ConfigSchemaRepository
	schemaValidator *services.SchemaValidator
	versionManager  *services.VersionManager
}

// NewRollbackConfigUseCase creates a new RollbackConfigUseCase
func NewRollbackConfigUseCase(
	configRepo outbound.ConfigRepository,
	revisionRepo outbound.ConfigRevisionRepository,
	schemaRepo outbound.ConfigSchemaRepository,
	schemaValidator *services.SchemaValidator,
	versionManager  *services.VersionManager,
) *RollbackConfigUseCase {
	return &RollbackConfigUseCase{
		configRepo:      configRepo,
		revisionRepo:    revisionRepo,
		schemaRepo:      schemaRepo,
		schemaValidator: schemaValidator,
		versionManager:  versionManager,
	}
}

// Execute rolls back a config to a previous version
func (uc *RollbackConfigUseCase) Execute(ctx context.Context, req RollbackConfigRequest) (*RollbackConfigResponse, error) {
	// Validate input
	if req.ProjectID == "" {
		return nil, fmt.Errorf("project ID is required")
	}
	if req.Key == "" {
		return nil, fmt.Errorf("config key is required")
	}
	if req.TargetVersion < 1 {
		return nil, fmt.Errorf("target version must be >= 1")
	}
	if req.ExpectedVersion < 1 {
		return nil, fmt.Errorf("expected version must be >= 1")
	}
	if req.RolledBackByUserID == "" {
		return nil, fmt.Errorf("user ID is required")
	}
	
	// Get current config
	currentConfig, err := uc.configRepo.Get(ctx, req.ProjectID, req.Key)
	if err != nil {
		return nil, fmt.Errorf("config not found: %w", err)
	}
	
	// Create version value objects
	currentVersion, err := valueobjects.NewVersion(currentConfig.Version)
	if err != nil {
		return nil, fmt.Errorf("invalid current version: %w", err)
	}
	
	expectedVersion, err := valueobjects.NewVersion(req.ExpectedVersion)
	if err != nil {
		return nil, fmt.Errorf("invalid expected version: %w", err)
	}
	
	targetVersion, err := valueobjects.NewVersion(req.TargetVersion)
	if err != nil {
		return nil, fmt.Errorf("invalid target version: %w", err)
	}
	
	// Validate optimistic lock
	if err := uc.versionManager.ValidateUpdate(expectedVersion, currentVersion, req.Key); err != nil {
		return nil, err
	}
	
	// Validate rollback target
	if !uc.versionManager.IsValidRollbackTarget(currentVersion, targetVersion) {
		return nil, fmt.Errorf("invalid rollback target: version %d (current: %d)", req.TargetVersion, currentConfig.Version)
	}
	
	// Get target revision
	targetRevision, err := uc.revisionRepo.GetByVersion(ctx, req.ProjectID, req.Key, req.TargetVersion)
	if err != nil {
		return nil, fmt.Errorf("target version not found: %w", err)
	}
	
	// Get schema for validation (use current schema, not historical)
	schema, err := uc.schemaRepo.GetByID(ctx, currentConfig.SchemaID)
	if err != nil {
		return nil, fmt.Errorf("schema not found: %w", err)
	}
	
	// Validate target content against current schema
	if err := uc.schemaValidator.ValidateOrError(schema.SchemaContent, targetRevision.Content); err != nil {
		return nil, fmt.Errorf("rollback validation failed (target content doesn't match current schema): %w", err)
	}
	
	// Update config to target content (creates new version)
	updatedConfig, err := uc.configRepo.Update(ctx, outbound.UpdateConfigParams{
		ProjectID:       req.ProjectID,
		Key:             req.Key,
		ExpectedVersion: req.ExpectedVersion,
		Content:         targetRevision.Content,
		UpdatedByUserID: req.RolledBackByUserID,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to rollback config: %w", err)
	}
	
	// Create revision for the rollback
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
		fmt.Printf("Warning: failed to create rollback revision: %v\n", err)
	}
	
	// TODO: Publish ConfigRolledBack event
	_ = events.NewConfigRolledBack(
		uuid.New().String(),
		updatedConfig.ProjectID,
		updatedConfig.Key,
		currentVersion,
		targetVersion,
		updatedConfig.Content,
		req.RolledBackByUserID,
	)
	
	return &RollbackConfigResponse{
		ProjectID:       updatedConfig.ProjectID,
		Key:             updatedConfig.Key,
		Version:         updatedConfig.Version,
		Content:         updatedConfig.Content,
		UpdatedByUserID: updatedConfig.UpdatedByUserID,
		UpdatedAt:       updatedConfig.UpdatedAt,
	}, nil
}

