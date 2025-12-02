package handlers

import (
	"encoding/json"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/vlone310/cfguardian/internal/adapters/inbound/http/common"
	"github.com/vlone310/cfguardian/internal/adapters/inbound/http/middleware"
	"github.com/vlone310/cfguardian/internal/usecases/schema"
)

// SchemaHandler handles config schema endpoints
type SchemaHandler struct {
	createUseCase *schema.CreateSchemaUseCase
	listUseCase   *schema.ListSchemasUseCase
	updateUseCase *schema.UpdateSchemaUseCase
	deleteUseCase *schema.DeleteSchemaUseCase
}

// NewSchemaHandler creates a new SchemaHandler
func NewSchemaHandler(
	createUseCase *schema.CreateSchemaUseCase,
	listUseCase *schema.ListSchemasUseCase,
	updateUseCase *schema.UpdateSchemaUseCase,
	deleteUseCase *schema.DeleteSchemaUseCase,
) *SchemaHandler {
	return &SchemaHandler{
		createUseCase: createUseCase,
		listUseCase:   listUseCase,
		updateUseCase: updateUseCase,
		deleteUseCase: deleteUseCase,
	}
}

// Create handles schema creation
// POST /api/v1/schemas
func (h *SchemaHandler) Create(w http.ResponseWriter, r *http.Request) {
	var reqBody struct {
		Name          string `json:"name"`
		SchemaContent string `json:"schema_content"`
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
	
	resp, err := h.createUseCase.Execute(r.Context(), schema.CreateSchemaRequest{
		Name:            reqBody.Name,
		SchemaContent:   reqBody.SchemaContent,
		CreatedByUserID: userID,
	})
	if err != nil {
		common.BadRequest(w, err.Error())
		return
	}
	
	common.Created(w, resp)
}

// List handles listing schemas
// GET /api/v1/schemas
func (h *SchemaHandler) List(w http.ResponseWriter, r *http.Request) {
	resp, err := h.listUseCase.Execute(r.Context())
	if err != nil {
		common.InternalServerError(w, err.Error())
		return
	}
	
	common.OK(w, resp)
}

// Update handles schema update
// PUT /api/v1/schemas/{schemaId}
func (h *SchemaHandler) Update(w http.ResponseWriter, r *http.Request) {
	schemaID := chi.URLParam(r, "schemaId")
	
	var reqBody struct {
		Name          *string `json:"name,omitempty"`
		SchemaContent *string `json:"schema_content,omitempty"`
	}
	
	if err := json.NewDecoder(r.Body).Decode(&reqBody); err != nil {
		common.BadRequest(w, "Invalid request body")
		return
	}
	
	resp, err := h.updateUseCase.Execute(r.Context(), schema.UpdateSchemaRequest{
		SchemaID:      schemaID,
		Name:          reqBody.Name,
		SchemaContent: reqBody.SchemaContent,
	})
	if err != nil {
		common.BadRequest(w, err.Error())
		return
	}
	
	common.OK(w, resp)
}

// Delete handles schema deletion
// DELETE /api/v1/schemas/{schemaId}
func (h *SchemaHandler) Delete(w http.ResponseWriter, r *http.Request) {
	schemaID := chi.URLParam(r, "schemaId")
	
	err := h.deleteUseCase.Execute(r.Context(), schema.DeleteSchemaRequest{
		SchemaID: schemaID,
	})
	if err != nil {
		common.BadRequest(w, err.Error())
		return
	}
	
	common.NoContent(w)
}

