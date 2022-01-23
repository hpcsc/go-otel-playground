package main

import (
	"context"
	"fmt"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/hpcsc/go-otel-playground/internal/config"
	"github.com/hpcsc/go-otel-playground/internal/handlers"
	internalMetric "github.com/hpcsc/go-otel-playground/internal/metric"
	"github.com/hpcsc/go-otel-playground/internal/tracer"
	"github.com/riandyrn/otelchi"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/metric/global"
	"go.opentelemetry.io/otel/propagation"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"
)

func main() {
	ctx := context.Background()
	c, err := config.Load()
	if err != nil {
		log.Fatal(err)
	}

	traceProvider, err := tracer.NewProvider(ctx, c.CollectorAddr)
	if err != nil {
		log.Fatal(err)
	}
	// set global tracer provider
	otel.SetTracerProvider(traceProvider)
	// set text propagators:
	// - propagation.TraceContext: W3C Trace Context format
	// - propagation.Baggage: W3C Baggage format
	otel.SetTextMapPropagator(propagation.NewCompositeTextMapPropagator(propagation.TraceContext{}, propagation.Baggage{}))

	metricProvider, err := internalMetric.NewProvider(ctx, c.CollectorAddr)
	if err != nil {
		log.Fatal(err)
	}
	global.SetMeterProvider(metricProvider)
	if err = metricProvider.Start(ctx); err != nil {
		log.Fatalf("failed to start metric provider: %v", err)
	}

	addr := fmt.Sprintf("0.0.0.0:%d", c.Port)
	server := &http.Server{Addr: addr, Handler: newRouter()}

	// Server run context
	serverCtx, shutdownDone := context.WithCancel(ctx)

	// Listen for syscall signals for process to interrupt/quit
	sig := make(chan os.Signal, 1)
	signal.Notify(sig, syscall.SIGHUP, syscall.SIGINT, syscall.SIGTERM, syscall.SIGQUIT)
	go func() {
		<-sig

		// Shutdown signal with grace period of 30 seconds
		shutdownCtx, _ := context.WithTimeout(serverCtx, 30*time.Second)

		go func() {
			<-shutdownCtx.Done()
			if shutdownCtx.Err() == context.DeadlineExceeded {
				log.Fatal("graceful shutdown timed out.. forcing exit.")
			}
		}()

		// shutdown tracer
		if err := traceProvider.Shutdown(ctx); err != nil {
			otel.Handle(err)
		}

		// stop metrics
		if err := metricProvider.Stop(ctx); err != nil {
			otel.Handle(err)
		}

		// shutdown server
		err := server.Shutdown(shutdownCtx)
		if err != nil {
			log.Fatal(err)
		}

		// signal to main goroutine that shutdown is done
		shutdownDone()
	}()

	// Run the server
	log.Printf("listening at %s\n", addr)
	err = server.ListenAndServe()
	if err != nil && err != http.ErrServerClosed {
		log.Fatal(err)
	}

	// Wait for server context to be stopped
	<-serverCtx.Done()
}

func newRouter() http.Handler {
	r := chi.NewRouter()

	r.Use(middleware.RequestID)
	r.Use(middleware.Logger)
	r.Use(otelchi.Middleware("go-otel-playground", otelchi.WithChiRoutes(r)))

	r.Get("/", handlers.Home)
	r.Get("/manual", handlers.ManualTrace)
	r.Get("/metric", handlers.Metric)

	return r
}
