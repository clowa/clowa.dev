// Package server assembles the gin engine: middleware plus the application's
// route handlers. Keeping this apart from main separates process concerns
// (config, signal handling, the listener) from HTTP wiring, and lets the router
// be exercised in tests without binding a socket.
package server

import (
	"github.com/gin-gonic/gin"
	"go.opentelemetry.io/contrib/instrumentation/github.com/gin-gonic/gin/otelgin"

	"github.com/clowa/clowa.dev/api/internal/quote"
)

// scopeName is the OpenTelemetry instrumentation scope name attached to the
// spans and metrics otelgin emits. It identifies the instrumentation, not the
// service — the service.name resource attribute comes from OTEL_SERVICE_NAME
// (see otel.go / serverless.yml).
const scopeName = "github.com/clowa/clowa.dev/api"

// New builds the HTTP router with all routes registered against the given quote
// repository. gin's mode (debug/release) is controlled out of band via the
// GIN_MODE environment variable, which gin reads on import.
func New(quotes quote.Repository) *gin.Engine {
	engine := gin.New()
	// otelgin first, so its span wraps the whole request (including the request
	// logger and Recovery below). It uses the global tracer/meter providers and
	// propagator set in setupOTel; with telemetry disabled those are no-ops, so
	// this is safe in tests and local runs. requestLogger then emits a structured
	// JSON line per request (to stdout and, via the OTLP bridge, to the collector),
	// correlated to the otelgin span; Recovery turns a panicking handler into a
	// 500 instead of crashing the process.
	engine.Use(otelgin.Middleware(scopeName), requestLogger(), gin.Recovery())

	quote.NewHandler(quotes).RegisterRoutes(engine)

	return engine
}
