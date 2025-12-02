package handlers

import (
	"encoding/json"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/vlone310/cfguardian/internal/adapters/inbound/http/common"
	"github.com/vlone310/cfguardian/internal/usecases/user"
)

// UserHandler handles user management endpoints
type UserHandler struct {
	createUseCase *user.CreateUserUseCase
	listUseCase   *user.ListUsersUseCase
	getUseCase    *user.GetUserUseCase
	deleteUseCase *user.DeleteUserUseCase
}

// NewUserHandler creates a new UserHandler
func NewUserHandler(
	createUseCase *user.CreateUserUseCase,
	listUseCase *user.ListUsersUseCase,
	getUseCase *user.GetUserUseCase,
	deleteUseCase *user.DeleteUserUseCase,
) *UserHandler {
	return &UserHandler{
		createUseCase: createUseCase,
		listUseCase:   listUseCase,
		getUseCase:    getUseCase,
		deleteUseCase: deleteUseCase,
	}
}

// Create handles user creation
// POST /api/v1/users
func (h *UserHandler) Create(w http.ResponseWriter, r *http.Request) {
	var req user.CreateUserRequest
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

// List handles listing all users
// GET /api/v1/users
func (h *UserHandler) List(w http.ResponseWriter, r *http.Request) {
	resp, err := h.listUseCase.Execute(r.Context())
	if err != nil {
		common.InternalServerError(w, err.Error())
		return
	}
	
	common.OK(w, resp)
}

// Get handles getting a user by ID
// GET /api/v1/users/{userId}
func (h *UserHandler) Get(w http.ResponseWriter, r *http.Request) {
	userID := chi.URLParam(r, "userId")
	
	resp, err := h.getUseCase.Execute(r.Context(), user.GetUserRequest{
		UserID: userID,
	})
	if err != nil {
		common.NotFound(w, err.Error())
		return
	}
	
	common.OK(w, resp)
}

// Delete handles user deletion
// DELETE /api/v1/users/{userId}
func (h *UserHandler) Delete(w http.ResponseWriter, r *http.Request) {
	userID := chi.URLParam(r, "userId")
	
	err := h.deleteUseCase.Execute(r.Context(), user.DeleteUserRequest{
		UserID: userID,
	})
	if err != nil {
		common.BadRequest(w, err.Error())
		return
	}
	
	common.NoContent(w)
}

