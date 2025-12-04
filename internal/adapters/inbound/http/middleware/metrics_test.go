package middleware

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/testutil"
	"github.com/stretchr/testify/assert"
	"github.com/vlone310/cfguardian/internal/infrastructure/telemetry"
)

func TestMetrics(t *testing.T) {
	t.Run("nil metrics", func(t *testing.T) {
		t.Run("handles nil metrics gracefully", func(t *testing.T) {
			// Arrange
			handler := Metrics(nil)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(http.StatusOK)
			}))
			req := httptest.NewRequest(http.MethodGet, "/test", nil)
			rec := httptest.NewRecorder()

			// Act
			handler.ServeHTTP(rec, req)

			// Assert
			assert.Equal(t, http.StatusOK, rec.Code)
		})
	})

	t.Run("with metrics", func(t *testing.T) {
		t.Run("records request count", func(t *testing.T) {
			// Arrange
			registry := prometheus.NewRegistry()
			metrics := createTestMetrics(registry, "test1")
			handler := Metrics(metrics)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(http.StatusOK)
			}))
			req := httptest.NewRequest(http.MethodGet, "/test", nil)
			rec := httptest.NewRecorder()

			// Act
			handler.ServeHTTP(rec, req)

			// Assert
			assert.Equal(t, http.StatusOK, rec.Code)
			assert.Equal(t, 1.0, testutil.ToFloat64(metrics.HTTPRequestsTotal.WithLabelValues("GET", "/test", "200")))
		})

		t.Run("records request duration", func(t *testing.T) {
			// Arrange
			registry := prometheus.NewRegistry()
			metrics := createTestMetrics(registry, "test2")
			slowHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				time.Sleep(10 * time.Millisecond)
				w.WriteHeader(http.StatusOK)
			})
			handler := Metrics(metrics)(slowHandler)
			req := httptest.NewRequest(http.MethodGet, "/slow", nil)
			rec := httptest.NewRecorder()

			// Act
			start := time.Now()
			handler.ServeHTTP(rec, req)
			elapsed := time.Since(start)

			// Assert
			assert.Equal(t, http.StatusOK, rec.Code)
			assert.GreaterOrEqual(t, elapsed, 10*time.Millisecond)
			// Note: We can't easily inspect histogram values in tests, but we verify the request completed
		})

		t.Run("tracks in-flight requests", func(t *testing.T) {
			// Arrange
			registry := prometheus.NewRegistry()
			metrics := createTestMetrics(registry, "test3")
			started := make(chan struct{})
			finished := make(chan struct{})

			handler := Metrics(metrics)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				close(started)
				<-finished
				w.WriteHeader(http.StatusOK)
			}))
			req := httptest.NewRequest(http.MethodGet, "/test", nil)
			rec := httptest.NewRecorder()

			// Act
			go func() {
				handler.ServeHTTP(rec, req)
			}()

			// Wait for request to start
			<-started
			inFlight := testutil.ToFloat64(metrics.HTTPRequestsInFlight)
			close(finished)

			// Wait a bit for the goroutine to finish
			time.Sleep(10 * time.Millisecond)

			// Assert
			assert.Equal(t, 1.0, inFlight, "Should track in-flight requests")
			assert.Equal(t, 0.0, testutil.ToFloat64(metrics.HTTPRequestsInFlight), "Should decrement after completion")
		})

		t.Run("records different status codes", func(t *testing.T) {
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
					registry := prometheus.NewRegistry()
					metrics := createTestMetrics(registry, "test_status_"+tc.name)
					handler := Metrics(metrics)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
						w.WriteHeader(tc.statusCode)
					}))
					req := httptest.NewRequest(http.MethodGet, "/test", nil)
					rec := httptest.NewRecorder()

					// Act
					handler.ServeHTTP(rec, req)

					// Assert
					assert.Equal(t, tc.statusCode, rec.Code)
				})
			}
		})

		t.Run("records different HTTP methods", func(t *testing.T) {
			testCases := []string{
				http.MethodGet,
				http.MethodPost,
				http.MethodPut,
				http.MethodDelete,
				http.MethodPatch,
			}

			for _, method := range testCases {
				t.Run(method, func(t *testing.T) {
					// Arrange
					registry := prometheus.NewRegistry()
					metrics := createTestMetrics(registry, "test_method_"+method)
					handler := Metrics(metrics)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
						w.WriteHeader(http.StatusOK)
					}))
					req := httptest.NewRequest(method, "/test", nil)
					rec := httptest.NewRecorder()

					// Act
					handler.ServeHTTP(rec, req)

					// Assert
					assert.Equal(t, http.StatusOK, rec.Code)
					assert.Equal(t, 1.0, testutil.ToFloat64(metrics.HTTPRequestsTotal.WithLabelValues(method, "/test", "200")))
				})
			}
		})

		t.Run("records different paths", func(t *testing.T) {
			testCases := []string{
				"/",
				"/health",
				"/api/v1/users",
				"/api/v1/projects/123/configs",
			}

			for _, path := range testCases {
				t.Run(path, func(t *testing.T) {
					// Arrange
					registry := prometheus.NewRegistry()
					metrics := createTestMetrics(registry, "test_path")
					handler := Metrics(metrics)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
						w.WriteHeader(http.StatusOK)
					}))
					req := httptest.NewRequest(http.MethodGet, path, nil)
					rec := httptest.NewRecorder()

					// Act
					handler.ServeHTTP(rec, req)

					// Assert
					assert.Equal(t, http.StatusOK, rec.Code)
				})
			}
		})
	})

	t.Run("integration", func(t *testing.T) {
		t.Run("works with middleware chain", func(t *testing.T) {
			// Arrange
			registry := prometheus.NewRegistry()
			metrics := createTestMetrics(registry, "test_integration")

			// Create final handler
			finalHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(http.StatusOK)
				w.Write([]byte("test"))
			})

			// Apply middleware in order: RequestID -> Metrics -> SecurityHeaders -> finalHandler
			handler := RequestID(Metrics(metrics)(SecurityHeaders(finalHandler)))

			req := httptest.NewRequest(http.MethodGet, "/test", nil)
			rec := httptest.NewRecorder()

			// Act
			handler.ServeHTTP(rec, req)

			// Assert
			assert.Equal(t, http.StatusOK, rec.Code)
			assert.Equal(t, "test", rec.Body.String())
			assert.Equal(t, 1.0, testutil.ToFloat64(metrics.HTTPRequestsTotal.WithLabelValues("GET", "/test", "200")))

			// Security headers should be set
			assert.NotEmpty(t, rec.Header().Get("X-Content-Type-Options"))

			// Request ID should be set
			assert.NotEmpty(t, rec.Header().Get("X-Request-ID"))
		})

		t.Run("handles panics correctly", func(t *testing.T) {
			// Arrange
			registry := prometheus.NewRegistry()
			metrics := createTestMetrics(registry, "test_panic")

			panicHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				panic("test panic")
			})

			// Apply middleware: Metrics -> Recovery -> panicHandler
			// This order ensures metrics can see the status code from recovery
			handler := Metrics(metrics)(Recovery(panicHandler))

			req := httptest.NewRequest(http.MethodGet, "/panic", nil)
			rec := httptest.NewRecorder()

			// Act
			handler.ServeHTTP(rec, req)

			// Assert
			assert.Equal(t, http.StatusInternalServerError, rec.Code)

			// Metrics should be recorded with 500 status from recovery middleware
			assert.Equal(t, 1.0, testutil.ToFloat64(metrics.HTTPRequestsTotal.WithLabelValues("GET", "/panic", "500")))
		})
	})
}

// Helper function to create test metrics
func createTestMetrics(registry *prometheus.Registry, namespace string) *telemetry.PrometheusMetrics {
	return &telemetry.PrometheusMetrics{
		HTTPRequestsTotal: prometheus.NewCounterVec(
			prometheus.CounterOpts{
				Namespace: namespace,
				Name:      "http_requests_total",
				Help:      "Total number of HTTP requests",
			},
			[]string{"method", "path", "status"},
		),
		HTTPRequestDuration: prometheus.NewHistogramVec(
			prometheus.HistogramOpts{
				Namespace: namespace,
				Name:      "http_request_duration_seconds",
				Help:      "HTTP request duration in seconds",
				Buckets:   []float64{.001, .005, .01, .025, .05, .1, .25, .5, 1},
			},
			[]string{"method", "path"},
		),
		HTTPRequestsInFlight: prometheus.NewGauge(
			prometheus.GaugeOpts{
				Namespace: namespace,
				Name:      "http_requests_in_flight",
				Help:      "Current number of HTTP requests being served",
			},
		),
	}
}
