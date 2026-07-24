// Command api is the clowa.dev backend: a small gin HTTP server that currently
// serves a single static quote at GET /api/quote. It listens on loopback only
// and runs behind Caddy, which reverse-proxies /api/* to it. s6-overlay
// supervises the process inside the container (see docker/s6-overlay).
package main

import (
	"context"
	"errors"
	"log"
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
	// StaticRepository today; swap for a PostgresRepository here once the
	// database exists — the rest of the wiring stays the same.
	repo := quote.NewStaticRepository()

	srv := &http.Server{
		Addr:    listenAddr(),
		Handler: server.New(repo),
		// Guard against slow-loris style stalls on the header read.
		ReadHeaderTimeout: 5 * time.Second,
	}

	// Serve in the background so main can block on shutdown signals.
	serverErr := make(chan error, 1)
	go func() {
		log.Printf("api listening on %s", srv.Addr)
		if err := srv.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			serverErr <- err
		}
	}()

	// s6-overlay sends SIGTERM on shutdown; also honour SIGINT for local runs.
	stop := make(chan os.Signal, 1)
	signal.Notify(stop, syscall.SIGINT, syscall.SIGTERM)

	select {
	case err := <-serverErr:
		log.Fatalf("api server error: %v", err)
	case sig := <-stop:
		log.Printf("api received %s, shutting down", sig)
	}

	ctx, cancel := context.WithTimeout(context.Background(), shutdownTimeout)
	defer cancel()
	if err := srv.Shutdown(ctx); err != nil {
		log.Fatalf("api shutdown error: %v", err)
	}
	log.Println("api stopped")
}
