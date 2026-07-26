package server_test

import (
	"bytes"
	"encoding/json"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"go.opentelemetry.io/otel"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	nooptrace "go.opentelemetry.io/otel/trace/noop"

	"github.com/clowa/clowa.dev/api/internal/quote"
	"github.com/clowa/clowa.dev/api/internal/server"
)

// captureLogs points the default slog logger at a buffer for the test's duration
// and restores the previous logger afterwards.
func captureLogs(t *testing.T) *bytes.Buffer {
	t.Helper()
	var buf bytes.Buffer
	prev := slog.Default()
	slog.SetDefault(slog.New(slog.NewJSONHandler(&buf, nil)))
	t.Cleanup(func() { slog.SetDefault(prev) })
	return &buf
}

// findRequestLog returns the first JSON log line whose msg is "request".
func findRequestLog(t *testing.T, buf *bytes.Buffer) map[string]any {
	t.Helper()
	for _, line := range bytes.Split(bytes.TrimSpace(buf.Bytes()), []byte("\n")) {
		if len(line) == 0 {
			continue
		}
		var rec map[string]any
		if err := json.Unmarshal(line, &rec); err != nil {
			continue
		}
		if rec["msg"] == "request" {
			return rec
		}
	}
	t.Fatalf("no \"request\" log line in output: %q", buf.String())
	return nil
}

// With a real tracer active, otelgin opens a recording span and the request
// logger must stamp each access log with the trace and span id so logs correlate
// to traces in Middleware.
func TestRequestLogIncludesTraceContext(t *testing.T) {
	otel.SetTracerProvider(sdktrace.NewTracerProvider(sdktrace.WithSampler(sdktrace.AlwaysSample())))
	t.Cleanup(func() { otel.SetTracerProvider(nooptrace.NewTracerProvider()) })

	buf := captureLogs(t)
	gin.SetMode(gin.TestMode)
	engine := server.New(quote.NewStaticRepository())

	engine.ServeHTTP(httptest.NewRecorder(), httptest.NewRequest(http.MethodGet, "/api/quote", nil))

	rec := findRequestLog(t, buf)
	if rec["method"] != http.MethodGet {
		t.Errorf("method = %v, want GET", rec["method"])
	}
	if got, ok := rec["status"].(float64); !ok || int(got) != http.StatusOK {
		t.Errorf("status = %v, want 200", rec["status"])
	}
	for _, k := range []string{"trace_id", "span_id"} {
		v, ok := rec[k].(string)
		if !ok || v == "" || allZero(v) {
			t.Errorf("%s = %v, want a non-zero id", k, rec[k])
		}
	}
}

// Without a tracer (telemetry off) there is no valid span, so the logger must omit
// the trace fields rather than emit all-zero ids.
func TestRequestLogOmitsTraceWhenNoSpan(t *testing.T) {
	otel.SetTracerProvider(nooptrace.NewTracerProvider())
	buf := captureLogs(t)
	gin.SetMode(gin.TestMode)
	engine := server.New(quote.NewStaticRepository())

	engine.ServeHTTP(httptest.NewRecorder(), httptest.NewRequest(http.MethodGet, "/api/quote", nil))

	rec := findRequestLog(t, buf)
	if _, ok := rec["trace_id"]; ok {
		t.Errorf("trace_id present without an active span: %v", rec["trace_id"])
	}
}

// With trusted proxies restricted to loopback (as main configures them), a
// client-supplied X-Forwarded-For arriving from an untrusted peer must be ignored:
// the logged client_ip is the real socket peer, not the spoofed header value.
func TestRequestLogClientIPIgnoresSpoofedForwardedFor(t *testing.T) {
	otel.SetTracerProvider(nooptrace.NewTracerProvider())
	buf := captureLogs(t)
	gin.SetMode(gin.TestMode)
	engine := server.New(quote.NewStaticRepository())
	if err := engine.SetTrustedProxies([]string{"127.0.0.1"}); err != nil {
		t.Fatalf("SetTrustedProxies: %v", err)
	}

	req := httptest.NewRequest(http.MethodGet, "/api/quote", nil)
	req.RemoteAddr = "203.0.113.7:54321" // untrusted peer
	req.Header.Set("X-Forwarded-For", "9.9.9.9")
	engine.ServeHTTP(httptest.NewRecorder(), req)

	rec := findRequestLog(t, buf)
	if got := rec["client_ip"]; got != "203.0.113.7" {
		t.Errorf("client_ip = %v, want 203.0.113.7 (spoofed X-Forwarded-For must be ignored)", got)
	}
}

func allZero(id string) bool {
	for _, c := range id {
		if c != '0' {
			return false
		}
	}
	return true
}
