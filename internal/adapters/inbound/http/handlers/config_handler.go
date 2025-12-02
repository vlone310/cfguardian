package handlers

import (
	"encoding/json"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/vlone310/cfguardian/internal/adapters/inbound/http/common"
	"github.com/vlone310/cfguardian/internal/adapters/inbound/http/middleware"
	"github.com/vlone310/cfguardian/internal/usecases/config"
)

// ConfigHandler handles configuration management endpoints
type ConfigHandler struct {
	createUseCase   *config.CreateConfigUseCase
	getUseCase      *config.GetConfigUseCase
	updateUseCase   *config.UpdateConfigUseCase
	deleteUseCase   *config.DeleteConfigUseCase
	rollbackUseCase *config.RollbackConfigUseCase
}

// NewConfigHandler creates a new ConfigHandler
func NewConfigHandler(
	createUseCase *config.CreateConfigUseCase,
	getUseCase *config.GetConfigUseCase,
	updateUseCase *config.UpdateConfigUseCase,
	deleteUseCase *config.DeleteConfigUseCase,
	rollbackUseCase *config.RollbackConfigUseCase,
) *ConfigHandler {
	return &ConfigHandler{
		createUseCase:   createUseCase,
		getUseCase:      getUseCase,
		updateUseCase:   updateUseCase,
		deleteUseCase:   deleteUseCase,
		rollbackUseCase: rollbackUseCase,
	}
}

// Create handles config creation
// POST /api/v1/projects/{projectId}/configs
func (h *ConfigHandler) Create(w http.ResponseWriter, r *http.Request) {
	projectID := chi.URLParam(r, "projectId")
	
	var reqBody struct {
		Key      string          `json:"key"`
		SchemaID string          `json:"schema_id"`
		Content  json.RawMessage `json:"content"`
	}
	
	if err := json.NewDecoder(r.Body).Decode(&reqBody); err != nil {
		common.BadRequest(w, "Invalid request body")
		return
	}
	
	// Get user ID from auth context
	userID := middleware.GetUserID(r.Context())
	if userID == "" {
		common.Unauthorized(w, "User not authenticated")
		return
	}
	
	resp, err := h.createUseCase.Execute(r.Context(), config.CreateConfigRequest{
		ProjectID:       projectID,
		Key:             reqBody.Key,
		SchemaID:        reqBody.SchemaID,
		Content:         reqBody.Content,
		UpdatedByUserID: userID,
	})
	if err != nil {
		common.BadRequest(w, err.Error())
		return
	}
	
	common.Created(w, resp)
}

// Get handles getting a config
// GET /api/v1/projects/{projectId}/configs/{configKey}
func (h *ConfigHandler) Get(w http.ResponseWriter, r *http.Request) {
	projectID := chi.URLParam(r, "projectId")
	configKey := chi.URLParam(r, "configKey")
	
	resp, err := h.getUseCase.Execute(r.Context(), config.GetConfigRequest{
		ProjectID: projectID,
		Key:       configKey,
	})
	if err != nil {
		common.NotFound(w, err.Error())
		return
	}
	
	common.OK(w, resp)
}

// Update handles config update with optimistic locking
// PUT /api/v1/projects/{projectId}/configs/{configKey}
func (h *ConfigHandler) Update(w http.ResponseWriter, r *http.Request) {
	projectID := chi.URLParam(r, "projectId")
	configKey := chi.URLParam(r, "configKey")
	
	var reqBody struct {
		ExpectedVersion int64           `json:"expected_version"`
		Content         json.RawMessage `json:"content"`
	}
	
	if err := json.NewDecoder(r.Body).Decode(&reqBody); err != nil {
		common.BadRequest(w, "Invalid request body")
		return
	}
	
	// Get user ID from auth context
	userID := middleware.GetUserID(r.Context())
	if userID == "" {
		common.Unauthorized(w, "User not authenticated")
		return
	}
	
	resp, err := h.updateUseCase.Execute(r.Context(), config.UpdateConfigRequest{
		ProjectID:       projectID,
		Key:             configKey,
		ExpectedVersion: reqBody.ExpectedVersion,
		Content:         reqBody.Content,
		UpdatedByUserID: userID,
	})
	if err != nil {
		// Check if it's a version conflict
		if isVersionConflict(err) {
			common.Conflict(w, err.Error())
			return
		}
		common.BadRequest(w, err.Error())
		return
	}
	
	common.OK(w, resp)
}

// Delete handles config deletion
// DELETE /api/v1/projects/{projectId}/configs/{configKey}
func (h *ConfigHandler) Delete(w http.ResponseWriter, r *http.Request) {
	projectID := chi.URLParam(r, "projectId")
	configKey := chi.URLParam(r, "configKey")
	
	// Get user ID from auth context
	userID := middleware.GetUserID(r.Context())
	if userID == "" {
		common.Unauthorized(w, "User not authenticated")
		return
	}
	
	err := h.deleteUseCase.Execute(r.Context(), config.DeleteConfigRequest{
		ProjectID:       projectID,
		Key:             configKey,
		DeletedByUserID: userID,
	})
	if err != nil {
		common.BadRequest(w, err.Error())
		return
	}
	
	common.NoContent(w)
}

// Rollback handles config rollback to a previous version
// POST /api/v1/projects/{projectId}/configs/{configKey}/rollback
func (h *ConfigHandler) Rollback(w http.ResponseWriter, r *http.Request) {
	projectID := chi.URLParam(r, "projectId")
	configKey := chi.URLParam(r, "configKey")
	
	var reqBody struct {
		TargetVersion   int64 `json:"target_version"`
		ExpectedVersion int64 `json:"expected_version"`
	}
	
	if err := json.NewDecoder(r.Body).Decode(&reqBody); err != nil {
		common.BadRequest(w, "Invalid request body")
		return
	}
	
	// Get user ID from auth context
	userID := middleware.GetUserID(r.Context())
	if userID == "" {
		common.Unauthorized(w, "User not authenticated")
		return
	}
	
	resp, err := h.rollbackUseCase.Execute(r.Context(), config.RollbackConfigRequest{
		ProjectID:         projectID,
		Key:               configKey,
		TargetVersion:     reqBody.TargetVersion,
		ExpectedVersion:   reqBody.ExpectedVersion,
		RolledBackByUserID: userID,
	})
	if err != nil {
		// Check if it's a version conflict
		if isVersionConflict(err) {
			common.Conflict(w, err.Error())
			return
		}
		common.BadRequest(w, err.Error())
		return
	}
	
	common.OK(w, resp)
}

// isVersionConflict checks if an error is a version conflict
func isVersionConflict(err error) bool {
	if err == nil {
		return false
	}
	// TODO: Use proper error type checking with domain/services.VersionConflictError
	errMsg := err.Error()
	return errMsg == "version mismatch" || 
		   stringContains(errMsg, "version conflict") ||
		   stringContains(errMsg, "concurrent modification")
}

// stringContains checks if a string contains a substring
func stringContains(s, substr string) bool {
	if len(substr) > len(s) {
		return false
	}
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}

