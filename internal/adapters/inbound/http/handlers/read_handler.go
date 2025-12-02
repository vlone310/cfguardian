package handlers

import (
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/vlone310/cfguardian/internal/adapters/inbound/http/common"
	"github.com/vlone310/cfguardian/internal/usecases/config"
)

// ReadHandler handles public client read API endpoints
type ReadHandler struct {
	readUseCase *config.ReadConfigByAPIKeyUseCase
}

// NewReadHandler creates a new ReadHandler
func NewReadHandler(readUseCase *config.ReadConfigByAPIKeyUseCase) *ReadHandler {
	return &ReadHandler{
		readUseCase: readUseCase,
	}
}

// Read handles reading a config by API key
// GET /api/v1/read/{apiKey}/{configKey}
func (h *ReadHandler) Read(w http.ResponseWriter, r *http.Request) {
	apiKey := chi.URLParam(r, "apiKey")
	configKey := chi.URLParam(r, "configKey")
	
	resp, err := h.readUseCase.Execute(r.Context(), config.ReadConfigByAPIKeyRequest{
		APIKey: apiKey,
		Key:    configKey,
	})
	if err != nil {
		common.NotFound(w, "Config not found")
		return
	}
	
	common.OK(w, resp)
}

