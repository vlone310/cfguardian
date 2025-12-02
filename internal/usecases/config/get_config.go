package config

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/vlone310/cfguardian/internal/ports/outbound"
)

// GetConfigRequest holds get config request data
type GetConfigRequest struct {
	ProjectID string
	Key       string
}

// GetConfigResponse holds config data
type GetConfigResponse struct {
	ProjectID       string
	Key             string
	SchemaID        string
	Version         int64
	Content         json.RawMessage
	UpdatedByUserID string
	CreatedAt       string
	UpdatedAt       string
}

// GetConfigUseCase handles retrieving a config
type GetConfigUseCase struct {
	configRepo outbound.ConfigRepository
}

// NewGetConfigUseCase creates a new GetConfigUseCase
func NewGetConfigUseCase(configRepo outbound.ConfigRepository) *GetConfigUseCase {
	return &GetConfigUseCase{
		configRepo: configRepo,
	}
}

// Execute retrieves a config by project ID and key
func (uc *GetConfigUseCase) Execute(ctx context.Context, req GetConfigRequest) (*GetConfigResponse, error) {
	// Validate input
	if req.ProjectID == "" {
		return nil, fmt.Errorf("project ID is required")
	}
	if req.Key == "" {
		return nil, fmt.Errorf("config key is required")
	}
	
	// Get config from repository
	config, err := uc.configRepo.Get(ctx, req.ProjectID, req.Key)
	if err != nil {
		return nil, fmt.Errorf("config not found: %w", err)
	}
	
	return &GetConfigResponse{
		ProjectID:       config.ProjectID,
		Key:             config.Key,
		SchemaID:        config.SchemaID,
		Version:         config.Version,
		Content:         config.Content,
		UpdatedByUserID: config.UpdatedByUserID,
		CreatedAt:       config.CreatedAt,
		UpdatedAt:       config.UpdatedAt,
	}, nil
}

