package server_test

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"

	"github.com/clowa/clowa.dev/api/internal/quote"
	"github.com/clowa/clowa.dev/api/internal/server"
)

// The assembled engine must route the quote endpoint end to end.
func TestNewRoutesQuoteEndpoint(t *testing.T) {
	gin.SetMode(gin.TestMode)
	engine := server.New(quote.NewStaticRepository())

	req := httptest.NewRequest(http.MethodGet, "/api/quote", nil)
	rec := httptest.NewRecorder()
	engine.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("GET /api/quote through server.New() = %d, want %d", rec.Code, http.StatusOK)
	}
}

// Unknown paths must 404 rather than being silently swallowed — a guard that
// the engine has no unexpected catch-all.
func TestNewUnknownPathIs404(t *testing.T) {
	gin.SetMode(gin.TestMode)
	engine := server.New(quote.NewStaticRepository())

	req := httptest.NewRequest(http.MethodGet, "/api/does-not-exist", nil)
	rec := httptest.NewRecorder()
	engine.ServeHTTP(rec, req)

	if rec.Code != http.StatusNotFound {
		t.Fatalf("GET /api/does-not-exist = %d, want %d", rec.Code, http.StatusNotFound)
	}
}
