package metric

import (
	"context"
	"fmt"
	"go.opentelemetry.io/otel/exporters/otlp/otlpmetric"
	"go.opentelemetry.io/otel/exporters/otlp/otlpmetric/otlpmetricgrpc"
	controller "go.opentelemetry.io/otel/sdk/metric/controller/basic"
	processor "go.opentelemetry.io/otel/sdk/metric/processor/basic"
	"go.opentelemetry.io/otel/sdk/metric/selector/simple"
	"time"
)

func NewProvider(ctx context.Context, collectorAddr string) (*controller.Controller, error) {
	exporter, err := newOTLPExporter(ctx, collectorAddr)
	if err != nil {
		return nil, fmt.Errorf("failed to create metric exporter: %v", err)
	}

	pusher := controller.New(
		processor.NewFactory(
			simple.NewWithHistogramDistribution(),
			exporter,
		),
		controller.WithExporter(exporter),
		controller.WithCollectPeriod(2*time.Second),
	)

	return pusher, nil
}

func newOTLPExporter(ctx context.Context, collectorAddr string) (*otlpmetric.Exporter, error) {
	metricClient := otlpmetricgrpc.NewClient(
		otlpmetricgrpc.WithInsecure(),
		otlpmetricgrpc.WithEndpoint(collectorAddr))
	return otlpmetric.New(ctx, metricClient)
}
