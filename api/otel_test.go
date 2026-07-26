package main

import (
	"context"
	"testing"
)

// With no OTLP endpoint configured (the local-dev / test default), setupOTel must
// be a quiet no-op: it returns a usable shutdown func (never nil) and no error, so
// callers can always defer the shutdown unconditionally. main relies on this
// contract to stay silent during `go run .` and `go test`.
func TestSetupOTelDisabledWhenEndpointUnset(t *testing.T) {
	// Empty value is treated as unset by setupOTel; t.Setenv also restores it.
	t.Setenv("OTEL_EXPORTER_OTLP_ENDPOINT", "")

	shutdown, err := setupOTel(context.Background())
	if err != nil {
		t.Fatalf("setupOTel() error = %v, want nil", err)
	}
	if shutdown == nil {
		t.Fatal("setupOTel() shutdown = nil, want a non-nil no-op func")
	}

	// The shutdown must be safe to call and idempotent — a second call (e.g. a
	// deferred flush after an explicit one) must not error or panic.
	if err := shutdown(context.Background()); err != nil {
		t.Fatalf("shutdown() error = %v, want nil", err)
	}
	if err := shutdown(context.Background()); err != nil {
		t.Fatalf("second shutdown() error = %v, want nil", err)
	}
}
