package handlers

import (
	"encoding/json"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/vlone310/cfguardian/internal/adapters/inbound/http/common"
	"github.com/vlone310/cfguardian/internal/usecases/project"
)

// ProjectHandler handles project management endpoints
type ProjectHandler struct {
	createUseCase *project.CreateProjectUseCase
	listUseCase   *project.ListProjectsUseCase
	getUseCase    *project.GetProjectUseCase
	deleteUseCase *project.DeleteProjectUseCase
}

// NewProjectHandler creates a new ProjectHandler
func NewProjectHandler(
	createUseCase *project.CreateProjectUseCase,
	listUseCase *project.ListProjectsUseCase,
	getUseCase *project.GetProjectUseCase,
	deleteUseCase *project.DeleteProjectUseCase,
) *ProjectHandler {
	return &ProjectHandler{
		createUseCase: createUseCase,
		listUseCase:   listUseCase,
		getUseCase:    getUseCase,
		deleteUseCase: deleteUseCase,
	}
}

// Create handles project creation
// POST /api/v1/projects
func (h *ProjectHandler) Create(w http.ResponseWriter, r *http.Request) {
	var req project.CreateProjectRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		common.BadRequest(w, "Invalid request body")
		return
	}
	
	resp, err := h.createUseCase.Execute(r.Context(), req)
	if err != nil {
		common.BadRequest(w, err.Error())
		return
	}
	
	common.Created(w, resp)
}

// List handles listing projects
// GET /api/v1/projects?owner_user_id=xxx
func (h *ProjectHandler) List(w http.ResponseWriter, r *http.Request) {
	ownerUserID := r.URL.Query().Get("owner_user_id")
	
	var req project.ListProjectsRequest
	if ownerUserID != "" {
		req.OwnerUserID = &ownerUserID
	}
	
	resp, err := h.listUseCase.Execute(r.Context(), req)
	if err != nil {
		common.InternalServerError(w, err.Error())
		return
	}
	
	common.OK(w, resp)
}

// Get handles getting a project by ID
// GET /api/v1/projects/{projectId}
func (h *ProjectHandler) Get(w http.ResponseWriter, r *http.Request) {
	projectID := chi.URLParam(r, "projectId")
	
	resp, err := h.getUseCase.Execute(r.Context(), project.GetProjectRequest{
		ProjectID: projectID,
	})
	if err != nil {
		common.NotFound(w, err.Error())
		return
	}
	
	common.OK(w, resp)
}

// Delete handles project deletion
// DELETE /api/v1/projects/{projectId}
func (h *ProjectHandler) Delete(w http.ResponseWriter, r *http.Request) {
	projectID := chi.URLParam(r, "projectId")
	
	err := h.deleteUseCase.Execute(r.Context(), project.DeleteProjectRequest{
		ProjectID: projectID,
	})
	if err != nil {
		common.BadRequest(w, err.Error())
		return
	}
	
	common.NoContent(w)
}

