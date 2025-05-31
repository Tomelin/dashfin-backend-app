package telemetry

import (
	"context"
	"log"
	"time"

	"example.com/profile-service/internal/config"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracegrpc"
	"go.opentelemetry.io/otel/exporters/prometheus"
	"go.opentelemetry.io/otel/propagation"
	"go.opentelemetry.io/otel/sdk/metric"
	"go.opentelemetry.io/otel/sdk/resource"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	semconv "go.opentelemetry.io/otel/semconv/v1.24.0" // Use a recent stable version
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure" // For OTLP exporter
)

// InitTracerProvider initializes and registers a global TracerProvider.
func InitTracerProvider(cfg *config.OpenTelemetryConfig) (*sdktrace.TracerProvider, error) {
	ctx := context.Background()

	res, err := resource.New(ctx,
		resource.WithAttributes(
			semconv.ServiceNameKey.String(cfg.ServiceName),
		),
	)
	if err != nil {
		return nil, err
	}

	// Setup OTLP gRPC exporter for traces
	// Ensure the OTLP endpoint is accessible and correctly configured.
    // For production, secure transport (TLS) should be configured.
    conn, err := grpc.DialContext(ctx, cfg.ExporterEndpoint,
        grpc.WithTransportCredentials(insecure.NewCredentials()), // Use insecure for now
        grpc.WithBlock(),
    )
    if err != nil {
        log.Printf("Failed to dial OTLP exporter: %v. Tracing will be impacted.", err)
        // Depending on policy, you might return the error or allow the app to start without tracing.
        // For now, returning error to make it explicit.
        return nil, err
    }

	traceExporter, err := otlptracegrpc.New(ctx, otlptracegrpc.WithGRPCConn(conn))
	if err != nil {
        log.Printf("Failed to create OTLP trace exporter: %v. Tracing will be impacted.", err)
		return nil, err
	}

	// Configure sampler
	sampler := sdktrace.ParentBased(sdktrace.TraceIDRatioBased(cfg.SampleRate))

	tp := sdktrace.NewTracerProvider(
		sdktrace.WithBatcher(traceExporter, sdktrace.WithBatchTimeout(time.Second*5)), // Batching options
		sdktrace.WithResource(res),
		sdktrace.WithSampler(sampler),
	)
	otel.SetTracerProvider(tp)
	log.Println("Global TracerProvider initialized.")
	return tp, nil
}

// InitMeterProvider initializes and registers a global MeterProvider with Prometheus exporter.
// The actual Prometheus HTTP endpoint needs to be exposed separately.
func InitMeterProvider(cfg *config.OpenTelemetryConfig) (*metric.MeterProvider, error) {
	// The prometheus exporter will be used to export metrics to Prometheus.
	// It typically needs an HTTP endpoint to be scraped.
	exporter, err := prometheus.New()
	if err != nil {
        log.Printf("Failed to create Prometheus exporter: %v. Metrics will be impacted.", err)
		return nil, err
	}

    // Resource for metrics, similar to traces
    res, err := resource.New(context.Background(),
		resource.WithAttributes(
			semconv.ServiceNameKey.String(cfg.ServiceName),
		),
	)
    if err != nil {
        log.Printf("Failed to create resource for metrics: %v. Metrics will be impacted.", err)
		return nil, err
	}

	meterProvider := metric.NewMeterProvider(
        metric.WithReader(exporter),
        metric.WithResource(res),
    )
	otel.SetMeterProvider(meterProvider)
	log.Println("Global MeterProvider initialized.")
	return meterProvider, nil
}

// InitGlobalPropagators sets up global text map propagators.
func InitGlobalPropagators() {
	// W3C Trace Context and Baggage propagators are standard.
	// The custom X-TRACE-ID will be handled by attempting to read it if no traceparent is found,
	// or by linking it to the current trace.
	otel.SetTextMapPropagator(propagation.NewCompositeTextMapPropagator(
		propagation.TraceContext{}, // Handles 'traceparent' and 'tracestate'
		propagation.Baggage{},    // Handles 'baggage'
	))
	log.Println("Global TextMapPropagator initialized (TraceContext & Baggage).")
}
