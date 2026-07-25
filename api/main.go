// Command api is the clowa.dev backend: a small gin HTTP server that currently
// serves a single static quote at GET /api/quote. It listens on loopback only
// and runs behind Caddy, which reverse-proxies /api/* to it. s6-overlay
// supervises the process inside the container (see docker/s6-overlay).
package main

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/clowa/clowa.dev/api/internal/quote"
	"github.com/clowa/clowa.dev/api/internal/server"
)

// defaultAddr binds loopback only: Caddy is the sole client and reaches the api
// over 127.0.0.1. Override with API_ADDR (e.g. for local testing).
const defaultAddr = "127.0.0.1:8080"

// shutdownTimeout bounds how long in-flight requests may drain on SIGTERM.
const shutdownTimeout = 10 * time.Second

func listenAddr() string {
	if addr := os.Getenv("API_ADDR"); addr != "" {
		return addr
	}
	return defaultAddr
}

func main() {
	// All the real work lives in run so its deferred cleanup — notably the
	// telemetry flush — actually executes. os.Exit skips defers, so the top level
	// only reports a fatal error run itself could not.
	if err := run(); err != nil {
		slog.Error("api exited with error", "err", err)
		os.Exit(1)
	}
}

// run wires up logging, telemetry and the HTTP server, blocks until a shutdown
// signal or a fatal server error, then drains in-flight requests and flushes
// telemetry. It returns an error instead of exiting so its defers (the OTel
// flush) run on every exit path.
func run() error {
	// Structured JSON logs to stdout + the OTLP bridge, installed first so every
	// subsequent line — including setup errors — is structured. See logging.go.
	slog.SetDefault(newLogger())

	// Telemetry next, so the server and its otelgin middleware export from the
	// first request. Driven entirely by OTEL_* env vars; a no-op when the
	// endpoint is unset (local dev). See otel.go.
	otelShutdown, err := setupOTel(context.Background())
	if err != nil {
		return fmt.Errorf("otel setup: %w", err)
	}
	// Flush and release the providers on the way out. Uses a fresh timeout so the
	// export still gets a chance even when run returns because of an error.
	defer func() {
		ctx, cancel := context.WithTimeout(context.Background(), shutdownTimeout)
		defer cancel()
		if err := otelShutdown(ctx); err != nil {
			slog.Error("otel shutdown error", "err", err)
		}
	}()

	// StaticRepository today; swap for a PostgresRepository here once the
	// database exists — the rest of the wiring stays the same.
	repo := quote.NewStaticRepository()

	srv := &http.Server{
		Addr:    listenAddr(),
		Handler: server.New(repo),
		// Guard against slow-loris style stalls on the header read.
		ReadHeaderTimeout: 5 * time.Second,
	}

	// Serve in the background so run can block on shutdown signals.
	serverErr := make(chan error, 1)
	go func() {
		slog.Info("api listening", "addr", srv.Addr)
		if err := srv.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			serverErr <- err
		}
	}()

	// s6-overlay sends SIGTERM on shutdown; also honour SIGINT for local runs.
	stop := make(chan os.Signal, 1)
	signal.Notify(stop, syscall.SIGINT, syscall.SIGTERM)

	select {
	case err := <-serverErr:
		return fmt.Errorf("server error: %w", err)
	case sig := <-stop:
		slog.Info("shutting down", "signal", sig.String())
	}

	ctx, cancel := context.WithTimeout(context.Background(), shutdownTimeout)
	defer cancel()
	if err := srv.Shutdown(ctx); err != nil {
		return fmt.Errorf("shutdown: %w", err)
	}
	slog.Info("api stopped")
	return nil
}
