package middleware

import (
	"net/http"
	"strconv"
	"time"

	"github.com/vlone310/cfguardian/internal/infrastructure/telemetry"
)

// Metrics middleware records metrics for HTTP requests
// Note: This middleware should be added BEFORE the Logging middleware
// so it can wrap the same responseWriter
func Metrics(metrics *telemetry.PrometheusMetrics) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if metrics == nil {
				next.ServeHTTP(w, r)
				return
			}
			
			// Track in-flight requests
			metrics.HTTPRequestsInFlight.Inc()
			defer metrics.HTTPRequestsInFlight.Dec()
			
			// Record start time
			start := time.Now()
			
			// Wrap response writer to capture status code (reuse existing type)
			rw := &responseWriter{
				ResponseWriter: w,
				status:         http.StatusOK,
			}
			
			// Serve the request
			next.ServeHTTP(rw, r)
			
			// Record metrics
			duration := time.Since(start).Seconds()
			status := strconv.Itoa(rw.status)
			
			// Record request count
			metrics.HTTPRequestsTotal.WithLabelValues(
				r.Method,
				r.URL.Path,
				status,
			).Inc()
			
			// Record request duration
			metrics.HTTPRequestDuration.WithLabelValues(
				r.Method,
				r.URL.Path,
			).Observe(duration)
		})
	}
}

