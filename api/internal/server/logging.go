package server

import (
	"log/slog"
	"time"

	"github.com/gin-gonic/gin"
	"go.opentelemetry.io/otel/trace"
)

// requestLogger emits one structured line per request through the default slog
// logger (JSON to stdout + the OTLP bridge; see the api's logging setup). It runs
// after otelgin, so when telemetry is enabled it attaches the active span's
// trace/span id for log<->trace correlation; with telemetry off the span context
// is invalid and the ids are omitted. It replaces gin.Logger, whose plain-text
// line is neither structured nor trace-aware.
func requestLogger() gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()

		c.Next()

		ctx := c.Request.Context()
		attrs := []slog.Attr{
			slog.String("method", c.Request.Method),
			slog.String("path", c.Request.URL.Path),
			slog.Int("status", c.Writer.Status()),
			slog.Duration("latency", time.Since(start)),
			slog.String("client_ip", c.ClientIP()),
		}
		// Passing ctx lets the OTLP bridge derive the trace context natively; the
		// explicit ids make the JSON stdout line correlate too.
		if sc := trace.SpanContextFromContext(ctx); sc.IsValid() {
			attrs = append(attrs,
				slog.String("trace_id", sc.TraceID().String()),
				slog.String("span_id", sc.SpanID().String()),
			)
		}
		slog.LogAttrs(ctx, slog.LevelInfo, "request", attrs...)
	}
}
