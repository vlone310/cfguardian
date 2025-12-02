package auth

import (
	"context"
	"fmt"

	"github.com/vlone310/cfguardian/internal/domain/valueobjects"
	"github.com/vlone310/cfguardian/internal/ports/outbound"
)

// ValidateAPIKeyRequest holds API key validation request
type ValidateAPIKeyRequest struct {
	APIKey string
}

// ValidateAPIKeyResponse holds validation response
type ValidateAPIKeyResponse struct {
	Valid     bool
	ProjectID string
	ProjectName string
}

// ValidateAPIKeyUseCase validates API keys for client access
type ValidateAPIKeyUseCase struct {
	projectRepo outbound.ProjectRepository
}

// NewValidateAPIKeyUseCase creates a new ValidateAPIKeyUseCase
func NewValidateAPIKeyUseCase(projectRepo outbound.ProjectRepository) *ValidateAPIKeyUseCase {
	return &ValidateAPIKeyUseCase{
		projectRepo: projectRepo,
	}
}

// Execute validates an API key
func (uc *ValidateAPIKeyUseCase) Execute(ctx context.Context, req ValidateAPIKeyRequest) (*ValidateAPIKeyResponse, error) {
	// Validate API key format
	apiKey, err := valueobjects.NewAPIKey(req.APIKey)
	if err != nil {
		return &ValidateAPIKeyResponse{Valid: false}, nil
	}
	
	// Find project by API key
	project, err := uc.projectRepo.GetByAPIKey(ctx, apiKey.Value())
	if err != nil {
		// API key not found or error
		return &ValidateAPIKeyResponse{Valid: false}, nil
	}
	
	return &ValidateAPIKeyResponse{
		Valid:       true,
		ProjectID:   project.ID,
		ProjectName: project.Name,
	}, nil
}

// ValidateAndGetProject validates API key and returns project if valid
func (uc *ValidateAPIKeyUseCase) ValidateAndGetProject(ctx context.Context, apiKeyStr string) (*outbound.Project, error) {
	// Validate API key format
	apiKey, err := valueobjects.NewAPIKey(apiKeyStr)
	if err != nil {
		return nil, fmt.Errorf("invalid API key format")
	}
	
	// Find project by API key
	project, err := uc.projectRepo.GetByAPIKey(ctx, apiKey.Value())
	if err != nil {
		return nil, fmt.Errorf("invalid API key")
	}
	
	return project, nil
}

