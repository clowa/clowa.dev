// Process logging setup for the api.
//
// Application logs go through log/slog to two sinks at once: structured JSON on
// stdout (captured by the api's s6-log pipeline and handy in local runs) and the
// OpenTelemetry log bridge, which ships to the collector when telemetry is enabled
// and is a no-op otherwise. Keeping stdout as JSON satisfies the rule that any log
// not emitted natively as an OTEL log is written in JSON.
package main

import (
	"context"
	"errors"
	"io"
	"log/slog"
	"os"

	"go.opentelemetry.io/contrib/bridges/otelslog"
)

// logScopeName is the OpenTelemetry instrumentation scope for the OTLP log bridge.
// It mirrors the server package's tracer scope so logs and spans share one scope.
const logScopeName = "github.com/clowa/clowa.dev/api"

// newLogger builds the process logger: JSON stdout fanned out to the OTLP bridge.
func newLogger() *slog.Logger {
	return newLoggerTo(os.Stdout)
}

// newLoggerTo is newLogger with an injectable stdout writer, for tests.
func newLoggerTo(w io.Writer) *slog.Logger {
	return slog.New(fanoutHandler{handlers: []slog.Handler{
		slog.NewJSONHandler(w, nil),
		otelslog.NewHandler(logScopeName),
	}})
}

// fanoutHandler delivers every record to each wrapped handler, so a single log
// call reaches both the JSON stdout sink and the OTLP bridge.
type fanoutHandler struct {
	handlers []slog.Handler
}

func (f fanoutHandler) Enabled(ctx context.Context, level slog.Level) bool {
	for _, h := range f.handlers {
		if h.Enabled(ctx, level) {
			return true
		}
	}
	return false
}

func (f fanoutHandler) Handle(ctx context.Context, r slog.Record) error {
	var errs error
	for _, h := range f.handlers {
		if !h.Enabled(ctx, r.Level) {
			continue
		}
		// Clone so a handler that retains or mutates the record cannot disturb the
		// copy handed to the next sink.
		if err := h.Handle(ctx, r.Clone()); err != nil {
			errs = errors.Join(errs, err)
		}
	}
	return errs
}

func (f fanoutHandler) WithAttrs(attrs []slog.Attr) slog.Handler {
	next := make([]slog.Handler, len(f.handlers))
	for i, h := range f.handlers {
		next[i] = h.WithAttrs(attrs)
	}
	return fanoutHandler{handlers: next}
}

func (f fanoutHandler) WithGroup(name string) slog.Handler {
	if name == "" {
		return f
	}
	next := make([]slog.Handler, len(f.handlers))
	for i, h := range f.handlers {
		next[i] = h.WithGroup(name)
	}
	return fanoutHandler{handlers: next}
}
