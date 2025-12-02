package telemetry

import (
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

// PrometheusMetrics holds Prometheus-specific metrics
type PrometheusMetrics struct {
	// HTTP Metrics
	HTTPRequestsTotal *prometheus.CounterVec
	HTTPRequestDuration *prometheus.HistogramVec
	HTTPRequestsInFlight prometheus.Gauge
	
	// Config Metrics
	ConfigOperationsTotal *prometheus.CounterVec
	ConfigVersionConflicts prometheus.Counter
	ConfigValidationErrors prometheus.Counter
	
	// Database Metrics
	DBConnectionsOpen prometheus.Gauge
	DBConnectionsInUse prometheus.Gauge
	DBConnectionsIdle prometheus.Gauge
	DBQueryDuration *prometheus.HistogramVec
	
	// Raft Metrics
	RaftState prometheus.Gauge
	RaftLeaderChanges prometheus.Counter
	RaftCommits prometheus.Counter
	RaftSnapshots prometheus.Counter
	RaftApplyDuration prometheus.Histogram
}

// NewPrometheusMetrics creates and registers Prometheus metrics
func NewPrometheusMetrics(namespace string) *PrometheusMetrics {
	return &PrometheusMetrics{
		// HTTP Metrics
		HTTPRequestsTotal: promauto.NewCounterVec(
			prometheus.CounterOpts{
				Namespace: namespace,
				Name:      "http_requests_total",
				Help:      "Total number of HTTP requests",
			},
			[]string{"method", "path", "status"},
		),
		HTTPRequestDuration: promauto.NewHistogramVec(
			prometheus.HistogramOpts{
				Namespace: namespace,
				Name:      "http_request_duration_seconds",
				Help:      "HTTP request duration in seconds",
				Buckets:   []float64{.001, .005, .01, .025, .05, .1, .25, .5, 1, 2.5, 5, 10},
			},
			[]string{"method", "path"},
		),
		HTTPRequestsInFlight: promauto.NewGauge(
			prometheus.GaugeOpts{
				Namespace: namespace,
				Name:      "http_requests_in_flight",
				Help:      "Current number of HTTP requests being served",
			},
		),
		
		// Config Metrics
		ConfigOperationsTotal: promauto.NewCounterVec(
			prometheus.CounterOpts{
				Namespace: namespace,
				Name:      "config_operations_total",
				Help:      "Total number of config operations",
			},
			[]string{"operation", "status"},
		),
		ConfigVersionConflicts: promauto.NewCounter(
			prometheus.CounterOpts{
				Namespace: namespace,
				Name:      "config_version_conflicts_total",
				Help:      "Total number of config version conflicts (optimistic locking)",
			},
		),
		ConfigValidationErrors: promauto.NewCounter(
			prometheus.CounterOpts{
				Namespace: namespace,
				Name:      "config_validation_errors_total",
				Help:      "Total number of config validation errors",
			},
		),
		
		// Database Metrics
		DBConnectionsOpen: promauto.NewGauge(
			prometheus.GaugeOpts{
				Namespace: namespace,
				Name:      "db_connections_open",
				Help:      "Current number of open database connections",
			},
		),
		DBConnectionsInUse: promauto.NewGauge(
			prometheus.GaugeOpts{
				Namespace: namespace,
				Name:      "db_connections_in_use",
				Help:      "Current number of database connections in use",
			},
		),
		DBConnectionsIdle: promauto.NewGauge(
			prometheus.GaugeOpts{
				Namespace: namespace,
				Name:      "db_connections_idle",
				Help:      "Current number of idle database connections",
			},
		),
		DBQueryDuration: promauto.NewHistogramVec(
			prometheus.HistogramOpts{
				Namespace: namespace,
				Name:      "db_query_duration_seconds",
				Help:      "Database query duration in seconds",
				Buckets:   []float64{.001, .005, .01, .025, .05, .1, .25, .5, 1},
			},
			[]string{"operation"},
		),
		
		// Raft Metrics
		RaftState: promauto.NewGauge(
			prometheus.GaugeOpts{
				Namespace: namespace,
				Name:      "raft_state",
				Help:      "Current Raft node state (0=Follower, 1=Candidate, 2=Leader)",
			},
		),
		RaftLeaderChanges: promauto.NewCounter(
			prometheus.CounterOpts{
				Namespace: namespace,
				Name:      "raft_leader_changes_total",
				Help:      "Total number of Raft leader changes",
			},
		),
		RaftCommits: promauto.NewCounter(
			prometheus.CounterOpts{
				Namespace: namespace,
				Name:      "raft_commits_total",
				Help:      "Total number of Raft log entries committed",
			},
		),
		RaftSnapshots: promauto.NewCounter(
			prometheus.CounterOpts{
				Namespace: namespace,
				Name:      "raft_snapshots_total",
				Help:      "Total number of Raft snapshots created",
			},
		),
		RaftApplyDuration: promauto.NewHistogram(
			prometheus.HistogramOpts{
				Namespace: namespace,
				Name:      "raft_apply_duration_seconds",
				Help:      "Raft FSM apply operation duration in seconds",
				Buckets:   []float64{.001, .005, .01, .025, .05, .1, .25, .5, 1},
			},
		),
	}
}

