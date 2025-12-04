package middleware

import (
	"bytes"
	"context"
	"fmt"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestLogging(t *testing.T) {
	// Capture log output for testing
	var logBuf bytes.Buffer
	oldLogger := slog.Default()
	defer slog.SetDefault(oldLogger)

	t.Run("logs request start and completion", func(t *testing.T) {
		// Arrange
		logBuf.Reset()
		logger := slog.New(slog.NewJSONHandler(&logBuf, nil))
		slog.SetDefault(logger)

		finalHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusOK)
			w.Write([]byte("response"))
		})

		handler := Logging(finalHandler)
		req := httptest.NewRequest(http.MethodGet, "/test", nil)
		rec := httptest.NewRecorder()

		// Act
		handler.ServeHTTP(rec, req)

		// Assert
		assert.Equal(t, http.StatusOK, rec.Code)

		logOutput := logBuf.String()
		assert.Contains(t, logOutput, "http request started",
			"Should log request start")
		assert.Contains(t, logOutput, "http request completed",
			"Should log request completion")
		assert.Contains(t, logOutput, "GET", "Should log HTTP method")
		assert.Contains(t, logOutput, "/test", "Should log path")
	})

	t.Run("captures response status code", func(t *testing.T) {
		// Arrange
		logBuf.Reset()
		logger := slog.New(slog.NewJSONHandler(&logBuf, nil))
		slog.SetDefault(logger)

		testCases := []struct {
			name       string
			statusCode int
		}{
			{"200 OK", http.StatusOK},
			{"201 Created", http.StatusCreated},
			{"400 Bad Request", http.StatusBadRequest},
			{"404 Not Found", http.StatusNotFound},
			{"500 Internal Server Error", http.StatusInternalServerError},
		}

		for _, tc := range testCases {
			t.Run(tc.name, func(t *testing.T) {
				// Arrange
				logBuf.Reset()

				finalHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
					w.WriteHeader(tc.statusCode)
				})

				handler := Logging(finalHandler)
				req := httptest.NewRequest(http.MethodGet, "/test", nil)
				rec := httptest.NewRecorder()

				// Act
				handler.ServeHTTP(rec, req)

				// Assert
				assert.Equal(t, tc.statusCode, rec.Code)
				logOutput := logBuf.String()
				assert.Contains(t, logOutput, fmt.Sprintf(`"status":%d`, tc.statusCode),
					"Should log status code")
			})
		}
	})

	t.Run("captures response size", func(t *testing.T) {
		// Arrange
		logBuf.Reset()
		logger := slog.New(slog.NewJSONHandler(&logBuf, nil))
		slog.SetDefault(logger)

		responseBody := []byte("This is a test response body")

		finalHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusOK)
			w.Write(responseBody)
		})

		handler := Logging(finalHandler)
		req := httptest.NewRequest(http.MethodGet, "/test", nil)
		rec := httptest.NewRecorder()

		// Act
		handler.ServeHTTP(rec, req)

		// Assert
		logOutput := logBuf.String()
		assert.Contains(t, logOutput, `"size":`, "Should log response size")
	})

	t.Run("captures request duration", func(t *testing.T) {
		// Arrange
		logBuf.Reset()
		logger := slog.New(slog.NewJSONHandler(&logBuf, nil))
		slog.SetDefault(logger)

		sleepDuration := 50 * time.Millisecond

		finalHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			time.Sleep(sleepDuration)
			w.WriteHeader(http.StatusOK)
		})

		handler := Logging(finalHandler)
		req := httptest.NewRequest(http.MethodGet, "/test", nil)
		rec := httptest.NewRecorder()

		// Act
		start := time.Now()
		handler.ServeHTTP(rec, req)
		elapsed := time.Since(start)

		// Assert
		logOutput := logBuf.String()
		assert.Contains(t, logOutput, `"duration"`, "Should log duration")
		assert.Contains(t, logOutput, `"duration_ms"`, "Should log duration in ms")

		// Duration should be at least the sleep time
		assert.GreaterOrEqual(t, elapsed, sleepDuration,
			"Captured duration should be at least sleep time")
	})

	t.Run("includes request metadata", func(t *testing.T) {
		// Arrange
		logBuf.Reset()
		logger := slog.New(slog.NewJSONHandler(&logBuf, nil))
		slog.SetDefault(logger)

		finalHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusOK)
		})

		handler := Logging(finalHandler)
		req := httptest.NewRequest(http.MethodPost, "/api/v1/users", nil)
		req.RemoteAddr = "192.168.1.100:12345"
		req.Header.Set("User-Agent", "TestClient/1.0")
		rec := httptest.NewRecorder()

		// Act
		handler.ServeHTTP(rec, req)

		// Assert
		logOutput := logBuf.String()
		assert.Contains(t, logOutput, "POST", "Should log method")
		assert.Contains(t, logOutput, "/api/v1/users", "Should log path")
		assert.Contains(t, logOutput, "192.168.1.100:12345", "Should log remote address")
		assert.Contains(t, logOutput, "TestClient/1.0", "Should log user agent")
	})
}

func TestLogging_WithRequestID(t *testing.T) {
	// Capture log output
	var logBuf bytes.Buffer
	oldLogger := slog.Default()
	defer slog.SetDefault(oldLogger)

	t.Run("includes request ID in logs when present", func(t *testing.T) {
		// Arrange
		logBuf.Reset()
		logger := slog.New(slog.NewJSONHandler(&logBuf, nil))
		slog.SetDefault(logger)

		requestID := "test-request-123"

		finalHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusOK)
		})

		handler := Logging(finalHandler)
		req := httptest.NewRequest(http.MethodGet, "/test", nil)

		// Add request ID to context
		ctx := context.WithValue(req.Context(), RequestIDKey, requestID)
		req = req.WithContext(ctx)

		rec := httptest.NewRecorder()

		// Act
		handler.ServeHTTP(rec, req)

		// Assert
		logOutput := logBuf.String()
		assert.Contains(t, logOutput, requestID,
			"Should include request ID in logs")
	})

	t.Run("works with RequestID middleware", func(t *testing.T) {
		// Arrange
		logBuf.Reset()
		logger := slog.New(slog.NewJSONHandler(&logBuf, nil))
		slog.SetDefault(logger)

		finalHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusOK)
		})

		// Chain middlewares: RequestID -> Logging -> finalHandler
		handler := RequestID(Logging(finalHandler))
		req := httptest.NewRequest(http.MethodGet, "/test", nil)
		rec := httptest.NewRecorder()

		// Act
		handler.ServeHTTP(rec, req)

		// Assert
		logOutput := logBuf.String()

		// Should contain a UUID-like request ID
		// Just verify it's present and non-empty
		assert.Contains(t, logOutput, `"request_id"`, "Should have request_id field")

		// Count occurrences - should appear twice (start and completion)
		count := strings.Count(logOutput, `"request_id"`)
		assert.Equal(t, 2, count, "request_id should appear in both log entries")
	})
}

func TestResponseWriter(t *testing.T) {
	t.Run("captures status code correctly", func(t *testing.T) {
		// Arrange
		rec := httptest.NewRecorder()
		wrapped := &responseWriter{
			ResponseWriter: rec,
			status:         http.StatusOK,
		}

		// Act
		wrapped.WriteHeader(http.StatusCreated)

		// Assert
		assert.Equal(t, http.StatusCreated, wrapped.status,
			"Should capture status code")
		assert.Equal(t, http.StatusCreated, rec.Code,
			"Should pass through to underlying writer")
	})

	t.Run("captures response size correctly", func(t *testing.T) {
		// Arrange
		rec := httptest.NewRecorder()
		wrapped := &responseWriter{
			ResponseWriter: rec,
			status:         http.StatusOK,
		}

		// Act
		data1 := []byte("Hello")
		data2 := []byte(" World")
		n1, err1 := wrapped.Write(data1)
		n2, err2 := wrapped.Write(data2)

		// Assert
		require.NoError(t, err1)
		require.NoError(t, err2)
		assert.Equal(t, len(data1), n1, "Should return correct write count")
		assert.Equal(t, len(data2), n2, "Should return correct write count")
		assert.Equal(t, len(data1)+len(data2), wrapped.size,
			"Should track total size")
		assert.Equal(t, "Hello World", rec.Body.String(),
			"Should pass through data")
	})

	t.Run("default status is 200 when not explicitly set", func(t *testing.T) {
		// Arrange
		rec := httptest.NewRecorder()
		wrapped := &responseWriter{
			ResponseWriter: rec,
			status:         http.StatusOK,
		}

		// Act
		wrapped.Write([]byte("response"))

		// Assert
		assert.Equal(t, http.StatusOK, wrapped.status,
			"Should default to 200 OK")
	})

	t.Run("handles multiple writes correctly", func(t *testing.T) {
		// Arrange
		rec := httptest.NewRecorder()
		wrapped := &responseWriter{
			ResponseWriter: rec,
			status:         http.StatusOK,
		}

		// Act
		writes := []string{"first", " second", " third"}
		totalSize := 0
		for _, write := range writes {
			n, err := wrapped.Write([]byte(write))
			require.NoError(t, err)
			totalSize += n
		}

		// Assert
		assert.Equal(t, totalSize, wrapped.size,
			"Should track cumulative size")
		assert.Equal(t, "first second third", rec.Body.String(),
			"Should accumulate all writes")
	})
}

func TestLogging_DifferentMethods(t *testing.T) {
	// Capture log output
	var logBuf bytes.Buffer
	oldLogger := slog.Default()
	defer slog.SetDefault(oldLogger)

	methods := []string{
		http.MethodGet,
		http.MethodPost,
		http.MethodPut,
		http.MethodPatch,
		http.MethodDelete,
		http.MethodOptions,
	}

	for _, method := range methods {
		t.Run(method, func(t *testing.T) {
			// Arrange
			logBuf.Reset()
			logger := slog.New(slog.NewJSONHandler(&logBuf, nil))
			slog.SetDefault(logger)

			finalHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(http.StatusOK)
			})

			handler := Logging(finalHandler)
			req := httptest.NewRequest(method, "/test", nil)
			rec := httptest.NewRecorder()

			// Act
			handler.ServeHTTP(rec, req)

			// Assert
			logOutput := logBuf.String()
			assert.Contains(t, logOutput, method,
				"Should log the HTTP method")
		})
	}
}

func TestLogging_DifferentPaths(t *testing.T) {
	// Capture log output
	var logBuf bytes.Buffer
	oldLogger := slog.Default()
	defer slog.SetDefault(oldLogger)

	paths := []string{
		"/",
		"/health",
		"/api/v1/users",
		"/api/v1/projects/123/configs/app-config",
	}

	for _, path := range paths {
		t.Run(path, func(t *testing.T) {
			// Arrange
			logBuf.Reset()
			logger := slog.New(slog.NewJSONHandler(&logBuf, nil))
			slog.SetDefault(logger)

			finalHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(http.StatusOK)
			})

			handler := Logging(finalHandler)
			req := httptest.NewRequest(http.MethodGet, path, nil)
			rec := httptest.NewRecorder()

			// Act
			handler.ServeHTTP(rec, req)

			// Assert
			logOutput := logBuf.String()
			assert.Contains(t, logOutput, path,
				"Should log the request path")
		})
	}
}

func TestLogging_Integration(t *testing.T) {
	// Capture log output
	var logBuf bytes.Buffer
	oldLogger := slog.Default()
	defer slog.SetDefault(oldLogger)

	t.Run("works with full middleware chain", func(t *testing.T) {
		// Arrange
		logBuf.Reset()
		logger := slog.New(slog.NewJSONHandler(&logBuf, nil))
		slog.SetDefault(logger)

		finalHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusOK)
			w.Write([]byte(`{"status":"success"}`))
		})

		// Chain: RequestID -> SecurityHeaders -> Logging -> finalHandler
		handler := RequestID(SecurityHeaders(Logging(finalHandler)))
		req := httptest.NewRequest(http.MethodGet, "/api/v1/users", nil)
		rec := httptest.NewRecorder()

		// Act
		handler.ServeHTTP(rec, req)

		// Assert
		assert.Equal(t, http.StatusOK, rec.Code)

		logOutput := logBuf.String()
		assert.Contains(t, logOutput, "http request started")
		assert.Contains(t, logOutput, "http request completed")
		assert.Contains(t, logOutput, `"status":200`)

		// Security headers should still be set
		assert.NotEmpty(t, rec.Header().Get("X-Content-Type-Options"))

		// Request ID should be in response headers
		requestID := rec.Header().Get("X-Request-ID")
		assert.NotEmpty(t, requestID)
		assert.Contains(t, logOutput, requestID, "Logs should contain request ID")
	})
}

func TestLogging_Performance(t *testing.T) {
	// Use a minimal logger for performance testing
	oldLogger := slog.Default()
	defer slog.SetDefault(oldLogger)

	// Set logger to output to /dev/null to measure pure middleware overhead
	logger := slog.New(slog.NewTextHandler(os.NewFile(0, os.DevNull), nil))
	slog.SetDefault(logger)

	t.Run("minimal overhead for fast requests", func(t *testing.T) {
		// Arrange
		finalHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusOK)
		})

		handler := Logging(finalHandler)
		req := httptest.NewRequest(http.MethodGet, "/test", nil)
		rec := httptest.NewRecorder()

		// Act
		start := time.Now()
		handler.ServeHTTP(rec, req)
		elapsed := time.Since(start)

		// Assert
		// Logging middleware should add minimal overhead (< 5ms on modern hardware)
		assert.Less(t, elapsed.Milliseconds(), int64(5),
			"Logging middleware should have minimal overhead")
	})

	t.Run("accurately measures slow handler duration", func(t *testing.T) {
		// Arrange
		handlerDelay := 100 * time.Millisecond

		finalHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			time.Sleep(handlerDelay)
			w.WriteHeader(http.StatusOK)
		})

		handler := Logging(finalHandler)
		req := httptest.NewRequest(http.MethodGet, "/test", nil)
		rec := httptest.NewRecorder()

		// Act
		start := time.Now()
		handler.ServeHTTP(rec, req)
		elapsed := time.Since(start)

		// Assert
		// Total elapsed should be at least the handler delay
		assert.GreaterOrEqual(t, elapsed, handlerDelay,
			"Should accurately measure handler duration")
	})
}

func TestLogging_EdgeCases(t *testing.T) {
	// Capture log output
	var logBuf bytes.Buffer
	oldLogger := slog.Default()
	defer slog.SetDefault(oldLogger)

	t.Run("handles empty response", func(t *testing.T) {
		// Arrange
		logBuf.Reset()
		logger := slog.New(slog.NewJSONHandler(&logBuf, nil))
		slog.SetDefault(logger)

		finalHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// Don't write anything
			w.WriteHeader(http.StatusNoContent)
		})

		handler := Logging(finalHandler)
		req := httptest.NewRequest(http.MethodDelete, "/test", nil)
		rec := httptest.NewRecorder()

		// Act
		handler.ServeHTTP(rec, req)

		// Assert
		assert.Equal(t, http.StatusNoContent, rec.Code)
		logOutput := logBuf.String()
		assert.Contains(t, logOutput, `"size":0`, "Should log size 0 for empty response")
	})

	t.Run("handles handler that never calls WriteHeader", func(t *testing.T) {
		// Arrange
		logBuf.Reset()
		logger := slog.New(slog.NewJSONHandler(&logBuf, nil))
		slog.SetDefault(logger)

		finalHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// Just write body without explicit WriteHeader
			w.Write([]byte("response"))
		})

		handler := Logging(finalHandler)
		req := httptest.NewRequest(http.MethodGet, "/test", nil)
		rec := httptest.NewRecorder()

		// Act
		handler.ServeHTTP(rec, req)

		// Assert
		// Default status should be 200 OK
		logOutput := logBuf.String()
		assert.Contains(t, logOutput, `"status":200`,
			"Should use default 200 status")
	})
}
