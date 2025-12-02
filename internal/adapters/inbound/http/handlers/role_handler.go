package handlers

import (
	"encoding/json"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/vlone310/cfguardian/internal/adapters/inbound/http/common"
	"github.com/vlone310/cfguardian/internal/usecases/role"
)

// RoleHandler handles role management endpoints
type RoleHandler struct {
	assignUseCase *role.AssignRoleUseCase
	revokeUseCase *role.RevokeRoleUseCase
}

// NewRoleHandler creates a new RoleHandler
func NewRoleHandler(
	assignUseCase *role.AssignRoleUseCase,
	revokeUseCase *role.RevokeRoleUseCase,
) *RoleHandler {
	return &RoleHandler{
		assignUseCase: assignUseCase,
		revokeUseCase: revokeUseCase,
	}
}

// Assign handles role assignment
// POST /api/v1/projects/{projectId}/roles
func (h *RoleHandler) Assign(w http.ResponseWriter, r *http.Request) {
	projectID := chi.URLParam(r, "projectId")
	
	var reqBody struct {
		UserID    string `json:"user_id"`
		RoleLevel string `json:"role_level"`
	}
	
	if err := json.NewDecoder(r.Body).Decode(&reqBody); err != nil {
		common.BadRequest(w, "Invalid request body")
		return
	}
	
	resp, err := h.assignUseCase.Execute(r.Context(), role.AssignRoleRequest{
		UserID:    reqBody.UserID,
		ProjectID: projectID,
		RoleLevel: reqBody.RoleLevel,
	})
	if err != nil {
		common.BadRequest(w, err.Error())
		return
	}
	
	common.OK(w, resp)
}

// Revoke handles role revocation
// DELETE /api/v1/projects/{projectId}/roles/{userId}
func (h *RoleHandler) Revoke(w http.ResponseWriter, r *http.Request) {
	projectID := chi.URLParam(r, "projectId")
	userID := chi.URLParam(r, "userId")
	
	err := h.revokeUseCase.Execute(r.Context(), role.RevokeRoleRequest{
		UserID:    userID,
		ProjectID: projectID,
	})
	if err != nil {
		common.BadRequest(w, err.Error())
		return
	}
	
	common.NoContent(w)
}

