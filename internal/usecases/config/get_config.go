package config

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/vlone310/cfguardian/internal/ports/outbound"
)

// GetConfigRequest holds get config request data
type GetConfigRequest struct {
	ProjectID string `json:"project_id"`
	Key       string `json:"key"`
}

// GetConfigResponse holds config data
type GetConfigResponse struct {
	ProjectID       string          `json:"project_id"`
	Key             string          `json:"key"`
	SchemaID        string          `json:"schema_id"`
	Version         int64           `json:"version"`
	Content         json.RawMessage `json:"content"`
	UpdatedByUserID string          `json:"updated_by_user_id"`
	CreatedAt       string          `json:"created_at"`
	UpdatedAt       string          `json:"updated_at"`
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

