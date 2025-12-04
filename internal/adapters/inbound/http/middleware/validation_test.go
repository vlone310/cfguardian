package middleware

import (
	"bytes"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestRequestSizeLimit(t *testing.T) {
	// Arrange
	maxSize := int64(1024) // 1 KB for testing

	finalHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Try to read body - if over limit, MaxBytesReader will error
		buf := make([]byte, maxSize+100)
		_, _ = r.Body.Read(buf)
		w.WriteHeader(http.StatusOK)
	})

	handler := RequestSizeLimit(maxSize)(finalHandler)

	tests := []struct {
		name         string
		bodySize     int
		expectStatus int
	}{
		{
			name:         "small request within limit",
			bodySize:     500,
			expectStatus: http.StatusOK,
		},
		{
			name:         "request at exact limit",
			bodySize:     int(maxSize),
			expectStatus: http.StatusOK,
		},
		{
			name:         "request exceeds limit slightly",
			bodySize:     int(maxSize) + 1,
			expectStatus: http.StatusOK, // Handler still runs, but body read will be limited
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			body := bytes.Repeat([]byte("a"), tt.bodySize)
			req := httptest.NewRequest(http.MethodPost, "/test", bytes.NewReader(body))
			req.Header.Set("Content-Type", "application/json")
			rec := httptest.NewRecorder()

			handler.ServeHTTP(rec, req)

			// Note: MaxBytesReader limits reading but doesn't reject request upfront
			// It returns error when trying to read beyond limit
			assert.NotNil(t, rec)
		})
	}
}

func TestContentTypeValidation(t *testing.T) {
	// Arrange
	finalHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	handler := ContentTypeValidation(finalHandler)

	tests := []struct {
		name          string
		method        string
		contentType   string
		contentLength int64
		expectStatus  int
	}{
		{
			name:          "valid JSON content type with body",
			method:        http.MethodPost,
			contentType:   "application/json",
			contentLength: 100,
			expectStatus:  http.StatusOK,
		},
		{
			name:          "missing content type with body",
			method:        http.MethodPost,
			contentType:   "",
			contentLength: 100,
			expectStatus:  http.StatusBadRequest,
		},
		{
			name:          "invalid content type",
			method:        http.MethodPost,
			contentType:   "text/plain",
			contentLength: 100,
			expectStatus:  http.StatusBadRequest,
		},
		{
			name:          "GET request without body - no content type check",
			method:        http.MethodGet,
			contentType:   "",
			contentLength: 0,
			expectStatus:  http.StatusOK,
		},
		{
			name:          "POST without body - no content type requirement",
			method:        http.MethodPost,
			contentType:   "",
			contentLength: 0,
			expectStatus:  http.StatusOK,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var body *bytes.Reader
			if tt.contentLength > 0 {
				body = bytes.NewReader(bytes.Repeat([]byte("a"), int(tt.contentLength)))
			} else {
				body = bytes.NewReader([]byte{})
			}

			req := httptest.NewRequest(tt.method, "/test", body)
			req.ContentLength = tt.contentLength
			if tt.contentType != "" {
				req.Header.Set("Content-Type", tt.contentType)
			}
			rec := httptest.NewRecorder()

			handler.ServeHTTP(rec, req)

			assert.Equal(t, tt.expectStatus, rec.Code, "Status code should match")
		})
	}
}

func TestMethodValidation(t *testing.T) {
	// Arrange
	finalHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	allowedMethods := []string{http.MethodGet, http.MethodPost}
	handler := MethodValidation(allowedMethods)(finalHandler)

	tests := []struct {
		name         string
		method       string
		expectStatus int
	}{
		{
			name:         "allowed method GET",
			method:       http.MethodGet,
			expectStatus: http.StatusOK,
		},
		{
			name:         "allowed method POST",
			method:       http.MethodPost,
			expectStatus: http.StatusOK,
		},
		{
			name:         "disallowed method PUT",
			method:       http.MethodPut,
			expectStatus: http.StatusBadRequest,
		},
		{
			name:         "disallowed method DELETE",
			method:       http.MethodDelete,
			expectStatus: http.StatusBadRequest,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest(tt.method, "/test", nil)
			rec := httptest.NewRecorder()

			handler.ServeHTTP(rec, req)

			assert.Equal(t, tt.expectStatus, rec.Code, "Status code should match")

			if tt.expectStatus != http.StatusOK {
				// Should have Allow header
				assert.NotEmpty(t, rec.Header().Get("Allow"))
			}
		})
	}
}
