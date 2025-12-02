package handlers

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
)

// HealthChecker defines interface for health checking components
type HealthChecker interface {
	Check(ctx context.Context) error
}

// HealthHandler handles health check endpoints
type HealthHandler struct {
	dbPool *pgxpool.Pool
	// raftStore could be added here for Raft health checks
}

// NewHealthHandler creates a new HealthHandler
func NewHealthHandler(dbPool *pgxpool.Pool) *HealthHandler {
	return &HealthHandler{
		dbPool: dbPool,
	}
}

// HealthResponse represents health check response
type HealthResponse struct {
	Status string `json:"status"`
}

// ReadinessResponse represents readiness check response
type ReadinessResponse struct {
	Status    string                  `json:"status"`
	Checks    map[string]CheckResult `json:"checks"`
	Timestamp string                 `json:"timestamp"`
}

// LivenessResponse represents liveness check response
type LivenessResponse struct {
	Status    string `json:"status"`
	Timestamp string `json:"timestamp"`
	Uptime    string `json:"uptime"`
}

// CheckResult represents individual health check result
type CheckResult struct {
	Status  string `json:"status"`
	Message string `json:"message,omitempty"`
}

var startTime = time.Now()

// Health handles basic health check
// GET /health
func (h *HealthHandler) Health(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(HealthResponse{
		Status: "healthy",
	})
}

// Readiness handles readiness probe (K8s-style)
// GET /ready
// Returns 200 if service is ready to accept traffic, 503 otherwise
func (h *HealthHandler) Readiness(w http.ResponseWriter, r *http.Request) {
	ctx, cancel := context.WithTimeout(r.Context(), 2*time.Second)
	defer cancel()
	
	checks := make(map[string]CheckResult)
	allHealthy := true
	
	// Check database connection
	if err := h.checkDatabase(ctx); err != nil {
		checks["database"] = CheckResult{
			Status:  "unhealthy",
			Message: err.Error(),
		}
		allHealthy = false
	} else {
		checks["database"] = CheckResult{
			Status: "healthy",
		}
	}
	
	// TODO: Add Raft health check
	// For now, assume Raft is healthy
	checks["raft"] = CheckResult{
		Status: "healthy",
	}
	
	response := ReadinessResponse{
		Status:    "healthy",
		Checks:    checks,
		Timestamp: time.Now().UTC().Format(time.RFC3339),
	}
	
	statusCode := http.StatusOK
	if !allHealthy {
		response.Status = "unhealthy"
		statusCode = http.StatusServiceUnavailable
	}
	
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	json.NewEncoder(w).Encode(response)
}

// Liveness handles liveness probe (K8s-style)
// GET /live
// Returns 200 if service is alive, 503 if it should be restarted
func (h *HealthHandler) Liveness(w http.ResponseWriter, r *http.Request) {
	// Liveness is simpler - just check if the process is responsive
	// If we can respond, we're alive
	
	uptime := time.Since(startTime)
	
	response := LivenessResponse{
		Status:    "alive",
		Timestamp: time.Now().UTC().Format(time.RFC3339),
		Uptime:    uptime.String(),
	}
	
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(response)
}

// checkDatabase verifies database connectivity
func (h *HealthHandler) checkDatabase(ctx context.Context) error {
	if h.dbPool == nil {
		return errors.New("database pool not initialized")
	}
	
	// Ping database
	return h.dbPool.Ping(ctx)
}

// checkRaft verifies Raft cluster health
// TODO: Implement this when raftStore is added to handler
func (h *HealthHandler) checkRaft(ctx context.Context) error {
	// Check if Raft is running
	// Check if node has a leader
	// Check if cluster has quorum
	return nil
}

