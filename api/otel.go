// OpenTelemetry SDK bootstrap for the api.
//
// The SDK is configured entirely from the standard OTEL_* environment variables
// (endpoint, protocol, headers, service name, resource attributes, sampler) — see
// serverless.yml for the values used in the container. Nothing is hard-coded
// here, so the same binary exports to the in-container otel-collector in
// production and stays silent during local `go run .`.
package main

import (
	"context"
	"errors"
	"log"
	"os"
	"time"

	"go.opentelemetry.io/contrib/exporters/autoexport"
	otelruntime "go.opentelemetry.io/contrib/instrumentation/runtime"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/propagation"
	sdkmetric "go.opentelemetry.io/otel/sdk/metric"
	"go.opentelemetry.io/otel/sdk/resource"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
)

// setupOTel installs the global tracer/meter providers and text-map propagator —
// all driven by OTEL_* env vars — plus Go runtime metrics, and returns a shutdown
// func that flushes and releases them.
//
// It is a no-op (telemetry disabled, no-op global providers left in place) when
// OTEL_EXPORTER_OTLP_ENDPOINT is unset. That keeps local development quiet while
// the container — which sets the endpoint to the loopback collector — exports
// normally. On a partial-setup error it tears down whatever was registered so it
// never leaks a half-initialised provider.
func setupOTel(ctx context.Context) (func(context.Context) error, error) {
	if os.Getenv("OTEL_EXPORTER_OTLP_ENDPOINT") == "" {
		log.Println("otel: disabled (OTEL_EXPORTER_OTLP_ENDPOINT unset)")
		return func(context.Context) error { return nil }, nil
	}

	var shutdownFns []func(context.Context) error
	shutdown := func(ctx context.Context) error {
		var errs error
		for _, fn := range shutdownFns {
			errs = errors.Join(errs, fn(ctx))
		}
		shutdownFns = nil
		return errs
	}

	// W3C trace context + baggage so spans stitch together across the
	// Caddy -> api hop (and any future service). otelgin reads these globals.
	otel.SetTextMapPropagator(propagation.NewCompositeTextMapPropagator(
		propagation.TraceContext{},
		propagation.Baggage{},
	))

	// Resource: service.name from OTEL_SERVICE_NAME, plus the attributes in
	// OTEL_RESOURCE_ATTRIBUTES (service.namespace, deployment.environment.name).
	res, err := resource.New(ctx,
		resource.WithFromEnv(),
		resource.WithTelemetrySDK(),
	)
	if err != nil {
		return shutdown, err
	}

	// Traces. autoexport builds the OTLP exporter from OTEL_EXPORTER_OTLP_*; the
	// sampler comes from OTEL_TRACES_SAMPLER[_ARG] because WithSampler is omitted.
	spanExporter, err := autoexport.NewSpanExporter(ctx)
	if err != nil {
		return shutdown, err
	}
	tracerProvider := sdktrace.NewTracerProvider(
		sdktrace.WithBatcher(spanExporter),
		sdktrace.WithResource(res),
	)
	shutdownFns = append(shutdownFns, tracerProvider.Shutdown)
	otel.SetTracerProvider(tracerProvider)

	// Metrics. autoexport builds the periodic reader/exporter from the same vars.
	metricReader, err := autoexport.NewMetricReader(ctx)
	if err != nil {
		return shutdown, errors.Join(err, shutdown(ctx))
	}
	meterProvider := sdkmetric.NewMeterProvider(
		sdkmetric.WithReader(metricReader),
		sdkmetric.WithResource(res),
	)
	shutdownFns = append(shutdownFns, meterProvider.Shutdown)
	otel.SetMeterProvider(meterProvider)

	// Go runtime metrics (goroutines, GC, heap) via the global meter provider.
	if err := otelruntime.Start(otelruntime.WithMinimumReadMemStatsInterval(time.Second)); err != nil {
		return shutdown, errors.Join(err, shutdown(ctx))
	}

	return shutdown, nil
}
