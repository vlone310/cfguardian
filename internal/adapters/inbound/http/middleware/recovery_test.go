package middleware

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestRecovery(t *testing.T) {
	t.Run("recovers from panic and returns 500", func(t *testing.T) {
		// Arrange
		panicMessage := "something went wrong"

		handler := Recovery(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			panic(panicMessage)
		}))

		req := httptest.NewRequest(http.MethodGet, "/test", nil)
		rec := httptest.NewRecorder()

		// Act
		handler.ServeHTTP(rec, req)

		// Assert
		assert.Equal(t, http.StatusInternalServerError, rec.Code,
			"Should return 500 status code")
		assert.Equal(t, "application/json", rec.Header().Get("Content-Type"),
			"Should set JSON content type")

		// Parse response body
		var response map[string]interface{}
		err := json.Unmarshal(rec.Body.Bytes(), &response)
		require.NoError(t, err, "Should return valid JSON")

		assert.Equal(t, "Internal server error", response["error"])
		assert.Equal(t, "INTERNAL_SERVER_ERROR", response["code"])
		assert.Contains(t, response, "request_id", "Should include request_id field")
		// Note: request_id will be empty if RequestID middleware is not used
	})

	t.Run("recovers from panic with request ID in context", func(t *testing.T) {
		// Arrange
		requestID := "test-request-123"

		handler := Recovery(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			panic("test panic")
		}))

		req := httptest.NewRequest(http.MethodGet, "/test", nil)

		// Add request ID to context
		ctx := context.WithValue(req.Context(), RequestIDKey, requestID)
		req = req.WithContext(ctx)

		rec := httptest.NewRecorder()

		// Act
		handler.ServeHTTP(rec, req)

		// Assert
		var response map[string]interface{}
		err := json.Unmarshal(rec.Body.Bytes(), &response)
		require.NoError(t, err)

		assert.Equal(t, requestID, response["request_id"],
			"Should preserve request ID from context")
	})

	t.Run("does not interfere with successful requests", func(t *testing.T) {
		// Arrange
		expectedStatus := http.StatusCreated
		expectedBody := `{"status":"success"}`

		handler := Recovery(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(expectedStatus)
			w.Write([]byte(expectedBody))
		}))

		req := httptest.NewRequest(http.MethodPost, "/test", nil)
		rec := httptest.NewRecorder()

		// Act
		handler.ServeHTTP(rec, req)

		// Assert
		assert.Equal(t, expectedStatus, rec.Code,
			"Should preserve original status code")
		assert.Equal(t, expectedBody, rec.Body.String(),
			"Should preserve original response body")
	})

	t.Run("recovers from different panic types", func(t *testing.T) {
		testCases := []struct {
			name       string
			panicValue interface{}
		}{
			{"string panic", "error message"},
			{"error panic", assert.AnError},
			{"integer panic", 42},
			{"nil panic", nil},
			{"struct panic", struct{ Message string }{"test error"}},
		}

		for _, tc := range testCases {
			t.Run(tc.name, func(t *testing.T) {
				// Arrange
				handler := Recovery(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
					panic(tc.panicValue)
				}))

				req := httptest.NewRequest(http.MethodGet, "/test", nil)
				rec := httptest.NewRecorder()

				// Act
				handler.ServeHTTP(rec, req)

				// Assert
				assert.Equal(t, http.StatusInternalServerError, rec.Code,
					"Should return 500 for all panic types")

				var response map[string]interface{}
				err := json.Unmarshal(rec.Body.Bytes(), &response)
				require.NoError(t, err, "Should return valid JSON")

				assert.Equal(t, "Internal server error", response["error"])
				assert.Equal(t, "INTERNAL_SERVER_ERROR", response["code"])
			})
		}
	})

	t.Run("recovers from panic in different HTTP methods", func(t *testing.T) {
		methods := []string{
			http.MethodGet,
			http.MethodPost,
			http.MethodPut,
			http.MethodPatch,
			http.MethodDelete,
		}

		for _, method := range methods {
			t.Run(method, func(t *testing.T) {
				// Arrange
				handler := Recovery(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
					panic("test panic")
				}))

				req := httptest.NewRequest(method, "/test", nil)
				rec := httptest.NewRecorder()

				// Act
				handler.ServeHTTP(rec, req)

				// Assert
				assert.Equal(t, http.StatusInternalServerError, rec.Code)
			})
		}
	})

	t.Run("recovers from panic at different paths", func(t *testing.T) {
		paths := []string{
			"/",
			"/api/v1/users",
			"/api/v1/configs",
			"/health",
		}

		for _, path := range paths {
			t.Run(path, func(t *testing.T) {
				// Arrange
				handler := Recovery(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
					panic("test panic")
				}))

				req := httptest.NewRequest(http.MethodGet, path, nil)
				rec := httptest.NewRecorder()

				// Act
				handler.ServeHTTP(rec, req)

				// Assert
				assert.Equal(t, http.StatusInternalServerError, rec.Code)
			})
		}
	})
}

func TestRecovery_ErrorResponse(t *testing.T) {
	t.Run("returns proper error structure", func(t *testing.T) {
		// Arrange
		handler := Recovery(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			panic("test panic")
		}))

		req := httptest.NewRequest(http.MethodGet, "/test", nil)
		rec := httptest.NewRecorder()

		// Act
		handler.ServeHTTP(rec, req)

		// Assert
		var response map[string]interface{}
		err := json.Unmarshal(rec.Body.Bytes(), &response)
		require.NoError(t, err)

		// Verify all required fields are present
		assert.Contains(t, response, "error", "Should contain error field")
		assert.Contains(t, response, "code", "Should contain code field")
		assert.Contains(t, response, "request_id", "Should contain request_id field")

		// Verify field types
		assert.IsType(t, "", response["error"], "error should be string")
		assert.IsType(t, "", response["code"], "code should be string")
		assert.IsType(t, "", response["request_id"], "request_id should be string")

		// Note: request_id will be empty if RequestID middleware is not used
		// This is expected behavior - the field is present but may be empty
	})

	t.Run("error message is user-friendly", func(t *testing.T) {
		// Arrange
		handler := Recovery(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			panic("internal database connection failed")
		}))

		req := httptest.NewRequest(http.MethodGet, "/test", nil)
		rec := httptest.NewRecorder()

		// Act
		handler.ServeHTTP(rec, req)

		// Assert
		var response map[string]interface{}
		err := json.Unmarshal(rec.Body.Bytes(), &response)
		require.NoError(t, err)

		// Should return generic error message, not internal details
		assert.Equal(t, "Internal server error", response["error"],
			"Should not expose internal error details")
	})
}

func TestRecovery_Integration(t *testing.T) {
	t.Run("works with request ID middleware", func(t *testing.T) {
		// Arrange - Chain both middlewares
		handler := RequestID(Recovery(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			panic("test panic")
		})))

		req := httptest.NewRequest(http.MethodGet, "/test", nil)
		rec := httptest.NewRecorder()

		// Act
		handler.ServeHTTP(rec, req)

		// Assert
		assert.Equal(t, http.StatusInternalServerError, rec.Code)

		var response map[string]interface{}
		err := json.Unmarshal(rec.Body.Bytes(), &response)
		require.NoError(t, err)

		// Request ID should be present and valid (UUID format)
		requestID, ok := response["request_id"].(string)
		require.True(t, ok, "request_id should be a string")
		assert.NotEmpty(t, requestID, "request_id should not be empty")
		assert.Len(t, requestID, 36, "request_id should be UUID format (36 chars)")
	})

	t.Run("works with security headers middleware", func(t *testing.T) {
		// Arrange - Chain both middlewares
		handler := SecurityHeaders(Recovery(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			panic("test panic")
		})))

		req := httptest.NewRequest(http.MethodGet, "/test", nil)
		rec := httptest.NewRecorder()

		// Act
		handler.ServeHTTP(rec, req)

		// Assert
		assert.Equal(t, http.StatusInternalServerError, rec.Code)

		// Security headers should still be set even after panic
		assert.NotEmpty(t, rec.Header().Get("X-Content-Type-Options"),
			"Security headers should be set even after panic")
	})

	t.Run("panic after partial response write", func(t *testing.T) {
		// Arrange
		handler := Recovery(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusOK)
			w.Write([]byte("partial response"))
			panic("panic after write")
		}))

		req := httptest.NewRequest(http.MethodGet, "/test", nil)
		rec := httptest.NewRecorder()

		// Act
		handler.ServeHTTP(rec, req)

		// Assert
		// Note: Status code cannot be changed after WriteHeader is called
		// This test documents the behavior when panic occurs after response starts
		assert.Equal(t, http.StatusOK, rec.Code,
			"Cannot change status after WriteHeader")
	})
}

func TestRecovery_RealWorldScenarios(t *testing.T) {
	t.Run("null pointer dereference", func(t *testing.T) {
		// Arrange
		handler := Recovery(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			var ptr *string
			_ = *ptr // This will panic
		}))

		req := httptest.NewRequest(http.MethodGet, "/test", nil)
		rec := httptest.NewRecorder()

		// Act
		handler.ServeHTTP(rec, req)

		// Assert
		assert.Equal(t, http.StatusInternalServerError, rec.Code,
			"Should recover from nil pointer panic")
	})

	t.Run("slice out of bounds", func(t *testing.T) {
		// Arrange
		handler := Recovery(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			slice := []int{1, 2, 3}
			_ = slice[10] // This will panic
		}))

		req := httptest.NewRequest(http.MethodGet, "/test", nil)
		rec := httptest.NewRecorder()

		// Act
		handler.ServeHTTP(rec, req)

		// Assert
		assert.Equal(t, http.StatusInternalServerError, rec.Code,
			"Should recover from index out of range panic")
	})

	t.Run("division by zero", func(t *testing.T) {
		// Arrange
		handler := Recovery(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			x := 10
			y := 0
			_ = x / y // This will panic
		}))

		req := httptest.NewRequest(http.MethodGet, "/test", nil)
		rec := httptest.NewRecorder()

		// Act
		handler.ServeHTTP(rec, req)

		// Assert
		assert.Equal(t, http.StatusInternalServerError, rec.Code,
			"Should recover from division by zero panic")
	})

	t.Run("type assertion failure with panic", func(t *testing.T) {
		// Arrange
		handler := Recovery(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			var i interface{} = "string"
			_ = i.(int) // This will panic
		}))

		req := httptest.NewRequest(http.MethodGet, "/test", nil)
		rec := httptest.NewRecorder()

		// Act
		handler.ServeHTTP(rec, req)

		// Assert
		assert.Equal(t, http.StatusInternalServerError, rec.Code,
			"Should recover from type assertion panic")
	})
}
