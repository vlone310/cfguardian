package telemetry

import (
	"context"
	"fmt"
	"time"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/exporters/otlp/otlpmetric/otlpmetrichttp"
	"go.opentelemetry.io/otel/metric"
	sdkmetric "go.opentelemetry.io/otel/sdk/metric"
	"go.opentelemetry.io/otel/sdk/resource"
	semconv "go.opentelemetry.io/otel/semconv/v1.17.0"
)

// MetricsConfig holds metrics configuration
type MetricsConfig struct {
	ServiceName     string
	ServiceVersion  string
	Endpoint        string
	MetricsInterval time.Duration
	Enabled         bool
}

// SetupMetrics initializes OpenTelemetry metrics
func SetupMetrics(ctx context.Context, cfg MetricsConfig) (metric.MeterProvider, func(context.Context) error, error) {
	if !cfg.Enabled {
		// Return a no-op meter provider
		mp := metric.NewNoopMeterProvider()
		return mp, func(context.Context) error { return nil }, nil
	}

	// Create OTLP metrics exporter
	exporter, err := otlpmetrichttp.New(ctx,
		otlpmetrichttp.WithEndpoint(cfg.Endpoint),
		otlpmetrichttp.WithInsecure(), // Use WithTLSClientConfig for production
	)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to create metrics exporter: %w", err)
	}

	// Create resource with service information
	res, err := resource.New(ctx,
		resource.WithAttributes(
			semconv.ServiceName(cfg.ServiceName),
			semconv.ServiceVersion(cfg.ServiceVersion),
		),
	)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to create resource: %w", err)
	}

	// Create meter provider
	mp := sdkmetric.NewMeterProvider(
		sdkmetric.WithResource(res),
		sdkmetric.WithReader(sdkmetric.NewPeriodicReader(
			exporter,
			sdkmetric.WithInterval(cfg.MetricsInterval),
		)),
	)

	// Set global meter provider
	otel.SetMeterProvider(mp)

	// Return shutdown function
	shutdown := func(ctx context.Context) error {
		return mp.Shutdown(ctx)
	}

	return mp, shutdown, nil
}

// GetMeter returns a meter for the given name
func GetMeter(name string) metric.Meter {
	return otel.Meter(name)
}

// Metrics holds application metrics
type Metrics struct {
	// HTTP Metrics
	HTTPRequestsTotal    metric.Int64Counter
	HTTPRequestDuration  metric.Float64Histogram
	HTTPRequestsInFlight metric.Int64UpDownCounter
	
	// Config Metrics
	ConfigOperationsTotal metric.Int64Counter
	ConfigVersionMismatches metric.Int64Counter
	ConfigValidationErrors metric.Int64Counter
	
	// Cache Metrics
	CacheHits   metric.Int64Counter
	CacheMisses metric.Int64Counter
	
	// Raft Metrics
	RaftLeaderChanges metric.Int64Counter
	RaftSnapshotCount metric.Int64Counter
}

// InitMetrics initializes application metrics
func InitMetrics(meterName string) (*Metrics, error) {
	meter := GetMeter(meterName)
	
	httpRequestsTotal, err := meter.Int64Counter(
		"http_requests_total",
		metric.WithDescription("Total number of HTTP requests"),
		metric.WithUnit("{request}"),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create http_requests_total: %w", err)
	}
	
	httpRequestDuration, err := meter.Float64Histogram(
		"http_request_duration_seconds",
		metric.WithDescription("HTTP request duration in seconds"),
		metric.WithUnit("s"),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create http_request_duration_seconds: %w", err)
	}
	
	httpRequestsInFlight, err := meter.Int64UpDownCounter(
		"http_requests_in_flight",
		metric.WithDescription("Current number of HTTP requests being served"),
		metric.WithUnit("{request}"),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create http_requests_in_flight: %w", err)
	}
	
	configOperationsTotal, err := meter.Int64Counter(
		"config_operations_total",
		metric.WithDescription("Total number of config operations"),
		metric.WithUnit("{operation}"),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create config_operations_total: %w", err)
	}
	
	configVersionMismatches, err := meter.Int64Counter(
		"config_version_mismatches_total",
		metric.WithDescription("Total number of config version mismatches"),
		metric.WithUnit("{mismatch}"),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create config_version_mismatches_total: %w", err)
	}
	
	configValidationErrors, err := meter.Int64Counter(
		"config_validation_errors_total",
		metric.WithDescription("Total number of config validation errors"),
		metric.WithUnit("{error}"),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create config_validation_errors_total: %w", err)
	}
	
	cacheHits, err := meter.Int64Counter(
		"cache_hits_total",
		metric.WithDescription("Total number of cache hits"),
		metric.WithUnit("{hit}"),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create cache_hits_total: %w", err)
	}
	
	cacheMisses, err := meter.Int64Counter(
		"cache_misses_total",
		metric.WithDescription("Total number of cache misses"),
		metric.WithUnit("{miss}"),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create cache_misses_total: %w", err)
	}
	
	raftLeaderChanges, err := meter.Int64Counter(
		"raft_leader_changes_total",
		metric.WithDescription("Total number of Raft leader changes"),
		metric.WithUnit("{change}"),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create raft_leader_changes_total: %w", err)
	}
	
	raftSnapshotCount, err := meter.Int64Counter(
		"raft_snapshots_total",
		metric.WithDescription("Total number of Raft snapshots created"),
		metric.WithUnit("{snapshot}"),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create raft_snapshots_total: %w", err)
	}
	
	return &Metrics{
		HTTPRequestsTotal:       httpRequestsTotal,
		HTTPRequestDuration:     httpRequestDuration,
		HTTPRequestsInFlight:    httpRequestsInFlight,
		ConfigOperationsTotal:   configOperationsTotal,
		ConfigVersionMismatches: configVersionMismatches,
		ConfigValidationErrors:  configValidationErrors,
		CacheHits:               cacheHits,
		CacheMisses:             cacheMisses,
		RaftLeaderChanges:       raftLeaderChanges,
		RaftSnapshotCount:       raftSnapshotCount,
	}, nil
}

