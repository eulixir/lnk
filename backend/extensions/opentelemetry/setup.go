package opentelemetry

import (
	"context"
	"fmt"
	"os"
	"time"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracegrpc"
	"go.opentelemetry.io/otel/sdk/resource"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	semconv "go.opentelemetry.io/otel/semconv/v1.37.0"
	"go.opentelemetry.io/otel/trace"
)

var Tracer trace.Tracer

func SetupOTelSDK(ctx context.Context, cfg *Config) (func(context.Context) error, error) {
	exp, err := newExporter(ctx, cfg.Endpoint)
	if err != nil {
		return nil, fmt.Errorf("failed to initialize exporter: %w", err)
	}

	tp := newTracerProvider(exp, cfg.ServiceName)

	otel.SetTracerProvider(tp)
	
	fmt.Fprintf(os.Stdout, "[DEBUG] OpenTelemetry tracer initialized, exporting to: %s\n", cfg.Endpoint)
	os.Stdout.Sync()

	Tracer = tp.Tracer(cfg.ServiceName)

	return func(ctx context.Context) error {
		return tp.Shutdown(ctx)
	}, nil
}

func newExporter(ctx context.Context, endpoint string) (sdktrace.SpanExporter, error) {
	client := otlptracegrpc.NewClient(
		otlptracegrpc.WithEndpoint(endpoint),
		otlptracegrpc.WithInsecure(),
	)
	
	traceExporter, err := otlptrace.New(ctx, client)
	if err != nil {
		return nil, fmt.Errorf("failed to create trace exporter: %w", err)
	}
	return traceExporter, nil
}

func newTracerProvider(exp sdktrace.SpanExporter, serviceName string) *sdktrace.TracerProvider {
	r, err := resource.Merge(
		resource.Default(),
		resource.NewWithAttributes(
			semconv.SchemaURL,
			semconv.ServiceNameKey.String(serviceName),
		),
	)

	if err != nil {
		panic(err)
	}

	return sdktrace.NewTracerProvider(
		sdktrace.WithBatcher(exp,
			sdktrace.WithBatchTimeout(1*time.Second), // Flush batches every 1 second
			sdktrace.WithMaxExportBatchSize(512),     // Export up to 512 spans per batch
		),
		sdktrace.WithResource(r),
	)
}
