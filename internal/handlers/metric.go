package handlers

import (
	"fmt"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/metric"
	"go.opentelemetry.io/otel/metric/global"
	"net/http"
)

func Metric(w http.ResponseWriter, r *http.Request) {
	meter := global.Meter("go-otel-playground")

	commonLabels := []attribute.KeyValue{
		attribute.String("method", "GET"),
	}

	requestCount := metric.Must(meter).
		NewInt64Counter(
			"request_counts",
			metric.WithDescription("The number of requests processed"),
		)

	meter.RecordBatch(
		r.Context(),
		commonLabels,
		requestCount.Measurement(1),
	)

	w.Write([]byte(fmt.Sprintf("metric sent.\n")))
}
