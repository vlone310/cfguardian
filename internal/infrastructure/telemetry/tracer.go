package telemetry

import (
	"context"
	"fmt"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracehttp"
	"go.opentelemetry.io/otel/propagation"
	"go.opentelemetry.io/otel/sdk/resource"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	semconv "go.opentelemetry.io/otel/semconv/v1.17.0"
	"go.opentelemetry.io/otel/trace"
)

// TracerConfig holds tracer configuration
type TracerConfig struct {
	ServiceName    string
	ServiceVersion string
	Endpoint       string
	TraceRatio     float64
	Enabled        bool
}

// SetupTracer initializes OpenTelemetry tracer
func SetupTracer(ctx context.Context, cfg TracerConfig) (trace.TracerProvider, func(context.Context) error, error) {
	if !cfg.Enabled {
		// Return a no-op tracer provider
		tp := trace.NewNoopTracerProvider()
		return tp, func(context.Context) error { return nil }, nil
	}

	// Create OTLP trace exporter
	client := otlptracehttp.NewClient(
		otlptracehttp.WithEndpoint(cfg.Endpoint),
		otlptracehttp.WithInsecure(), // Use WithTLSClientConfig for production
	)
	
	exporter, err := otlptrace.New(ctx, client)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to create trace exporter: %w", err)
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

	// Create trace provider
	tp := sdktrace.NewTracerProvider(
		sdktrace.WithBatcher(exporter),
		sdktrace.WithResource(res),
		sdktrace.WithSampler(sdktrace.TraceIDRatioBased(cfg.TraceRatio)),
	)

	// Set global tracer provider
	otel.SetTracerProvider(tp)

	// Set global propagator for W3C Trace Context propagation
	otel.SetTextMapPropagator(propagation.NewCompositeTextMapPropagator(
		propagation.TraceContext{},
		propagation.Baggage{},
	))

	// Return shutdown function
	shutdown := func(ctx context.Context) error {
		return tp.Shutdown(ctx)
	}

	return tp, shutdown, nil
}

// GetTracer returns a tracer for the given name
func GetTracer(name string) trace.Tracer {
	return otel.Tracer(name)
}

