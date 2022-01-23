package tracer

import (
	"context"
	"fmt"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracegrpc"
	"go.opentelemetry.io/otel/exporters/stdout/stdouttrace"
	"go.opentelemetry.io/otel/sdk/resource"
	sdk "go.opentelemetry.io/otel/sdk/trace"
	semconv "go.opentelemetry.io/otel/semconv/v1.7.0"
	"go.opentelemetry.io/otel/trace"
	"google.golang.org/grpc"
)

const tracerName = "github.com/hpcsc/go-otel-playground"

func StartSpan(ctx context.Context, name string) (context.Context, trace.Span) {
	return otel.Tracer(tracerName).Start(ctx, name)
}

func NewProvider(ctx context.Context, collectorAddr string) (*sdk.TracerProvider, error) {
	exporter, err := newOTLPExporter(ctx, collectorAddr)
	if err != nil {
		return nil, fmt.Errorf("failed to create trace exporter: %v", err)
	}

	res, err := newResource()
	if err != nil {
		return nil, fmt.Errorf("failed to create trace resource: %v", err)
	}

	provider := sdk.NewTracerProvider(
		sdk.WithBatcher(exporter),
		sdk.WithResource(res),
	)

	return provider, nil
}

func newOTLPExporter(ctx context.Context, collectorAddr string) (sdk.SpanExporter, error) {
	traceClient := otlptracegrpc.NewClient(
		otlptracegrpc.WithInsecure(),
		otlptracegrpc.WithEndpoint(collectorAddr),
		otlptracegrpc.WithDialOption(
			grpc.WithBlock(),
			grpc.FailOnNonTempDialError(true),
		))
	return otlptrace.New(ctx, traceClient)
}

func newConsoleExporter() (*stdouttrace.Exporter, error) {
	return stdouttrace.New(
		// Use human-readable output.
		stdouttrace.WithPrettyPrint(),
		// Do not print timestamps for the demo.
		stdouttrace.WithoutTimestamps(),
	)
}

func newResource() (*resource.Resource, error) {
	return resource.Merge(
		resource.Default(),
		resource.NewWithAttributes(
			semconv.SchemaURL,
			semconv.ServiceNameKey.String("go-otel-playground"),
			semconv.ServiceVersionKey.String("v0.1.0"),
			attribute.String("environment", "local"),
		),
	)
}
