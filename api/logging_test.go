package main

import (
	"bytes"
	"context"
	"encoding/json"
	"log/slog"
	"sync"
	"testing"

	"go.opentelemetry.io/contrib/bridges/otelslog"
	sdklog "go.opentelemetry.io/otel/sdk/log"
	"go.opentelemetry.io/otel/trace"
)

// The stdout sink must be JSON (decision 5: non-OTLP-native logs are JSON) and
// carry the structured attributes it was given.
func TestNewLoggerToWritesJSON(t *testing.T) {
	var buf bytes.Buffer
	logger := newLoggerTo(&buf)

	logger.Info("hello", slog.String("key", "val"))

	var rec map[string]any
	if err := json.Unmarshal(bytes.TrimSpace(buf.Bytes()), &rec); err != nil {
		t.Fatalf("stdout line is not valid JSON: %v\nline: %q", err, buf.String())
	}
	if rec["msg"] != "hello" {
		t.Errorf("msg = %v, want hello", rec["msg"])
	}
	if rec["level"] != "INFO" {
		t.Errorf("level = %v, want INFO", rec["level"])
	}
	if rec["key"] != "val" {
		t.Errorf("key = %v, want val", rec["key"])
	}
}

// The fanout handler must deliver each record to every sink, so the stdout mirror
// and the OTLP bridge both see the same log line.
func TestFanoutHandlerEmitsToEverySink(t *testing.T) {
	var a, b bytes.Buffer
	logger := slog.New(fanoutHandler{handlers: []slog.Handler{
		slog.NewJSONHandler(&a, nil),
		slog.NewJSONHandler(&b, nil),
	}})

	logger.Info("fan")

	for name, buf := range map[string]*bytes.Buffer{"a": &a, "b": &b} {
		if !bytes.Contains(buf.Bytes(), []byte(`"msg":"fan"`)) {
			t.Errorf("sink %s missing the record: %q", name, buf.String())
		}
	}
}

// A log emitted with a span-carrying context must reach the OTLP pipeline as one
// record stamped with that trace/span id — the correlation Middleware relies on.
func TestOTLPLogCarriesTraceContext(t *testing.T) {
	exp := &memLogExporter{}
	lp := sdklog.NewLoggerProvider(sdklog.WithProcessor(sdklog.NewBatchProcessor(exp)))
	t.Cleanup(func() { _ = lp.Shutdown(context.Background()) })

	logger := slog.New(otelslog.NewHandler("test", otelslog.WithLoggerProvider(lp)))

	tid, _ := trace.TraceIDFromHex("0123456789abcdef0123456789abcdef")
	sid, _ := trace.SpanIDFromHex("0123456789abcdef")
	sc := trace.NewSpanContext(trace.SpanContextConfig{
		TraceID: tid, SpanID: sid, TraceFlags: trace.FlagsSampled,
	})
	ctx := trace.ContextWithSpanContext(context.Background(), sc)

	logger.InfoContext(ctx, "correlated")
	if err := lp.ForceFlush(context.Background()); err != nil {
		t.Fatalf("force flush: %v", err)
	}

	if got := exp.count(); got != 1 {
		t.Fatalf("exported records = %d, want 1", got)
	}
	rec := exp.first()
	if got := rec.TraceID(); got != tid {
		t.Errorf("record trace id = %v, want %v", got, tid)
	}
}

// memLogExporter is an in-memory sdklog.Exporter that records everything it is
// handed, so tests can assert on the emitted log records.
type memLogExporter struct {
	mu      sync.Mutex
	records []sdklog.Record
}

func (e *memLogExporter) Export(_ context.Context, recs []sdklog.Record) error {
	e.mu.Lock()
	defer e.mu.Unlock()
	e.records = append(e.records, recs...)
	return nil
}

func (e *memLogExporter) Shutdown(context.Context) error   { return nil }
func (e *memLogExporter) ForceFlush(context.Context) error { return nil }

func (e *memLogExporter) count() int {
	e.mu.Lock()
	defer e.mu.Unlock()
	return len(e.records)
}

func (e *memLogExporter) first() sdklog.Record {
	e.mu.Lock()
	defer e.mu.Unlock()
	return e.records[0]
}
