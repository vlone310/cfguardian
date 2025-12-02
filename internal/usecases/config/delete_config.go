package config

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"github.com/vlone310/cfguardian/internal/domain/events"
	"github.com/vlone310/cfguardian/internal/domain/valueobjects"
	"github.com/vlone310/cfguardian/internal/ports/outbound"
)

// DeleteConfigRequest holds delete config request data
type DeleteConfigRequest struct {
	ProjectID       string `json:"project_id"`
	Key             string `json:"key"`
	DeletedByUserID string `json:"deleted_by_user_id"`
}

// DeleteConfigUseCase handles config deletion
type DeleteConfigUseCase struct {
	configRepo outbound.ConfigRepository
}

// NewDeleteConfigUseCase creates a new DeleteConfigUseCase
func NewDeleteConfigUseCase(configRepo outbound.ConfigRepository) *DeleteConfigUseCase {
	return &DeleteConfigUseCase{
		configRepo: configRepo,
	}
}

// Execute deletes a config
func (uc *DeleteConfigUseCase) Execute(ctx context.Context, req DeleteConfigRequest) error {
	// Validate input
	if req.ProjectID == "" {
		return fmt.Errorf("project ID is required")
	}
	if req.Key == "" {
		return fmt.Errorf("config key is required")
	}
	if req.DeletedByUserID == "" {
		return fmt.Errorf("user ID is required")
	}
	
	// Get config before deleting (for event)
	config, err := uc.configRepo.Get(ctx, req.ProjectID, req.Key)
	if err != nil {
		return fmt.Errorf("config not found: %w", err)
	}
	
	// Delete config
	// Note: Foreign key CASCADE will automatically delete all config_revisions
	if err := uc.configRepo.Delete(ctx, req.ProjectID, req.Key); err != nil {
		return fmt.Errorf("failed to delete config: %w", err)
	}
	
	// TODO: Publish ConfigDeleted event
	lastVersion, _ := valueobjects.NewVersion(config.Version)
	_ = events.NewConfigDeleted(
		uuid.New().String(),
		req.ProjectID,
		req.Key,
		lastVersion,
		req.DeletedByUserID,
	)
	
	return nil
}

