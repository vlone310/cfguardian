package common

import (
	"encoding/json"
	"net/http"
)

// ErrorResponse represents an API error response
type ErrorResponse struct {
	Error   string                 `json:"error"`
	Code    string                 `json:"code"`
	Details map[string]interface{} `json:"details,omitempty"`
}

// SuccessResponse represents a generic success response
type SuccessResponse struct {
	Data    interface{} `json:"data,omitempty"`
	Message string      `json:"message,omitempty"`
}

// RespondJSON writes a JSON response
func RespondJSON(w http.ResponseWriter, status int, payload interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	
	if payload != nil {
		if err := json.NewEncoder(w).Encode(payload); err != nil {
			http.Error(w, "Failed to encode response", http.StatusInternalServerError)
		}
	}
}

// RespondError writes an error response
func RespondError(w http.ResponseWriter, status int, message, code string) {
	RespondJSON(w, status, ErrorResponse{
		Error: message,
		Code:  code,
	})
}

// RespondErrorWithDetails writes an error response with additional details
func RespondErrorWithDetails(w http.ResponseWriter, status int, message, code string, details map[string]interface{}) {
	RespondJSON(w, status, ErrorResponse{
		Error:   message,
		Code:    code,
		Details: details,
	})
}

// Common error responses

// BadRequest responds with 400 Bad Request
func BadRequest(w http.ResponseWriter, message string) {
	RespondError(w, http.StatusBadRequest, message, "BAD_REQUEST")
}

// Unauthorized responds with 401 Unauthorized
func Unauthorized(w http.ResponseWriter, message string) {
	RespondError(w, http.StatusUnauthorized, message, "UNAUTHORIZED")
}

// Forbidden responds with 403 Forbidden
func Forbidden(w http.ResponseWriter, message string) {
	RespondError(w, http.StatusForbidden, message, "FORBIDDEN")
}

// NotFound responds with 404 Not Found
func NotFound(w http.ResponseWriter, message string) {
	RespondError(w, http.StatusNotFound, message, "NOT_FOUND")
}

// Conflict responds with 409 Conflict
func Conflict(w http.ResponseWriter, message string) {
	RespondError(w, http.StatusConflict, message, "CONFLICT")
}

// InternalServerError responds with 500 Internal Server Error
func InternalServerError(w http.ResponseWriter, message string) {
	RespondError(w, http.StatusInternalServerError, message, "INTERNAL_SERVER_ERROR")
}

// Created responds with 201 Created
func Created(w http.ResponseWriter, payload interface{}) {
	RespondJSON(w, http.StatusCreated, payload)
}

// OK responds with 200 OK
func OK(w http.ResponseWriter, payload interface{}) {
	RespondJSON(w, http.StatusOK, payload)
}

// NoContent responds with 204 No Content
func NoContent(w http.ResponseWriter) {
	w.WriteHeader(http.StatusNoContent)
}

