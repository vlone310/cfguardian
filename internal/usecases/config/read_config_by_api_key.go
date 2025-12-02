package config

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/vlone310/cfguardian/internal/domain/valueobjects"
	"github.com/vlone310/cfguardian/internal/ports/outbound"
)

// ReadConfigByAPIKeyRequest holds client config read request
type ReadConfigByAPIKeyRequest struct {
	APIKey string
	Key    string
}

// ReadConfigByAPIKeyResponse holds config data for clients
type ReadConfigByAPIKeyResponse struct {
	Key     string
	Version int64
	Content json.RawMessage
}

// ReadConfigByAPIKeyUseCase handles read-only config access for clients
// This is the primary endpoint for external applications to fetch their configs
type ReadConfigByAPIKeyUseCase struct {
	projectRepo outbound.ProjectRepository
	configRepo  outbound.ConfigRepository
}

// NewReadConfigByAPIKeyUseCase creates a new ReadConfigByAPIKeyUseCase
func NewReadConfigByAPIKeyUseCase(
	projectRepo outbound.ProjectRepository,
	configRepo outbound.ConfigRepository,
) *ReadConfigByAPIKeyUseCase {
	return &ReadConfigByAPIKeyUseCase{
		projectRepo: projectRepo,
		configRepo:  configRepo,
	}
}

// Execute reads a config by API key (client access)
func (uc *ReadConfigByAPIKeyUseCase) Execute(ctx context.Context, req ReadConfigByAPIKeyRequest) (*ReadConfigByAPIKeyResponse, error) {
	// Validate input
	if req.APIKey == "" {
		return nil, fmt.Errorf("API key is required")
	}
	if req.Key == "" {
		return nil, fmt.Errorf("config key is required")
	}
	
	// Validate and find project by API key
	apiKey, err := valueobjects.NewAPIKey(req.APIKey)
	if err != nil {
		return nil, fmt.Errorf("invalid API key format")
	}
	
	project, err := uc.projectRepo.GetByAPIKey(ctx, apiKey.Value())
	if err != nil {
		return nil, fmt.Errorf("invalid API key")
	}
	
	// Get config from project
	config, err := uc.configRepo.Get(ctx, project.ID, req.Key)
	if err != nil {
		return nil, fmt.Errorf("config not found")
	}
	
	// Return config (without sensitive metadata)
	return &ReadConfigByAPIKeyResponse{
		Key:     config.Key,
		Version: config.Version,
		Content: config.Content,
	}, nil
}

// ExecuteMultiple reads multiple configs by API key in one call
func (uc *ReadConfigByAPIKeyUseCase) ExecuteMultiple(ctx context.Context, apiKey string, keys []string) (map[string]*ReadConfigByAPIKeyResponse, error) {
	if apiKey == "" {
		return nil, fmt.Errorf("API key is required")
	}
	if len(keys) == 0 {
		return nil, fmt.Errorf("at least one config key is required")
	}
	
	// Validate and find project
	apiKeyVO, err := valueobjects.NewAPIKey(apiKey)
	if err != nil {
		return nil, fmt.Errorf("invalid API key format")
	}
	
	project, err := uc.projectRepo.GetByAPIKey(ctx, apiKeyVO.Value())
	if err != nil {
		return nil, fmt.Errorf("invalid API key")
	}
	
	// Get all configs for the project
	configs, err := uc.configRepo.ListByProject(ctx, project.ID)
	if err != nil {
		return nil, fmt.Errorf("failed to list configs: %w", err)
	}
	
	// Build map of requested configs
	result := make(map[string]*ReadConfigByAPIKeyResponse)
	configMap := make(map[string]*outbound.Config)
	for _, config := range configs {
		configMap[config.Key] = config
	}
	
	// Return only requested keys
	for _, key := range keys {
		if config, found := configMap[key]; found {
			result[key] = &ReadConfigByAPIKeyResponse{
				Key:     config.Key,
				Version: config.Version,
				Content: config.Content,
			}
		}
	}
	
	return result, nil
}

