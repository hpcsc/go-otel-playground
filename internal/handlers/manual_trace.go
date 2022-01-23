package handlers

import (
	"context"
	"fmt"
	"github.com/hpcsc/go-otel-playground/internal/tracer"
	"net/http"
	"time"
)

func ManualTrace(w http.ResponseWriter, r *http.Request) {
	doWork(r.Context())

	w.Write([]byte(fmt.Sprintf("all done.\n")))
}

func doWork(ctx context.Context) {
	_, span := tracer.StartSpan(ctx, "doWork")
	defer span.End()

	time.Sleep(1 * time.Second)
}
