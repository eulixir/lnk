package opentelemetry

import (
	"context"
	"fmt"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracegrpc"
	"go.opentelemetry.io/otel/sdk/resource"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	semconv "go.opentelemetry.io/otel/semconv/v1.37.0"
	"go.opentelemetry.io/otel/trace"
)

var tracer trace.Tracer

func SetupOTelSDK(ctx context.Context, cfg *Config) (func(context.Context) error, error) {
	exp, err := newExporter(ctx, cfg.Endpoint)
	if err != nil {
		return nil, fmt.Errorf("failed to initialize exporter: %w", err)
	}

	tp := newTracerProvider(exp, cfg.ServiceName)

	defer func() { _ = tp.Shutdown(ctx) }()

	otel.SetTracerProvider(tp)

	tracer = tp.Tracer(cfg.ServiceName)

	return func(ctx context.Context) error {
		return tp.Shutdown(ctx)
	}, nil
}

func newExporter(ctx context.Context, endpoint string) (sdktrace.SpanExporter, error) {
	traceExporter, err := otlptracegrpc.New(ctx, otlptracegrpc.WithEndpoint(endpoint), otlptracegrpc.WithInsecure())
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
			semconv.ServiceName(serviceName),
		),
	)

	if err != nil {
		panic(err)
	}

	return sdktrace.NewTracerProvider(
		sdktrace.WithBatcher(exp),
		sdktrace.WithResource(r),
	)
}
