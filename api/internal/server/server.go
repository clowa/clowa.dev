// Package server assembles the gin engine: middleware plus the application's
// route handlers. Keeping this apart from main separates process concerns
// (config, signal handling, the listener) from HTTP wiring, and lets the router
// be exercised in tests without binding a socket.
package server

import (
	"github.com/gin-gonic/gin"

	"github.com/clowa/clowa.dev/api/internal/quote"
)

// New builds the HTTP router with all routes registered against the given quote
// repository. gin's mode (debug/release) is controlled out of band via the
// GIN_MODE environment variable, which gin reads on import.
func New(quotes quote.Repository) *gin.Engine {
	engine := gin.New()
	// Logger writes request lines to stdout (captured by the api's s6-log
	// pipeline); Recovery turns a panicking handler into a 500 instead of
	// crashing the process.
	engine.Use(gin.Logger(), gin.Recovery())

	quote.NewHandler(quotes).RegisterRoutes(engine)

	return engine
}
