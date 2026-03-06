package tracing

import (
	"context"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracegrpc"
	"go.opentelemetry.io/otel/propagation"
	"go.opentelemetry.io/otel/sdk/resource"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	semconv "go.opentelemetry.io/otel/semconv/v1.24.0"
)

// InitTracer initializes an OTLP exporter, and configures the corresponding trace and
// metric providers.
func InitTracer() (*sdktrace.TracerProvider, error) {
	ctx := context.Background()

	// Configure an OTLP gRPC Exporter pointing to localhost:4317
	exporter, err := otlptracegrpc.New(ctx,
		otlptracegrpc.WithEndpoint("localhost:4317"),
		otlptracegrpc.WithInsecure(),
	)
	if err != nil {
		return nil, err
	}

	// Set up a TracerProvider with a Resource specifying the service name
	res, err := resource.New(ctx,
		resource.WithAttributes(
			semconv.ServiceNameKey.String("api-gateway"),
		),
	)
	if err != nil {
		return nil, err
	}

	// Configures a Sampling strategy: TraceIDRatioBased(1.0) for testing
	tp := sdktrace.NewTracerProvider(
		sdktrace.WithSampler(sdktrace.TraceIDRatioBased(1.0)),
		sdktrace.WithBatcher(exporter),
		sdktrace.WithResource(res),
	)

	// Sets the global TracerProvider and global TextMapPropagator
	otel.SetTracerProvider(tp)
	otel.SetTextMapPropagator(propagation.NewCompositeTextMapPropagator(propagation.TraceContext{}, propagation.Baggage{}))

	return tp, nil
}
