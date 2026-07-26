package quote_test

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"

	"github.com/clowa/clowa.dev/api/internal/quote"
)

func newTestRouter(repo quote.Repository) *gin.Engine {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	quote.NewHandler(repo).RegisterRoutes(r)
	return r
}

func doGet(r *gin.Engine, target string) *httptest.ResponseRecorder {
	req := httptest.NewRequest(http.MethodGet, target, nil)
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)
	return rec
}

// Acceptance test: GET /api/quote returns 200 with the default quote as JSON.
func TestGetQuoteReturnsStaticQuote(t *testing.T) {
	rec := doGet(newTestRouter(quote.NewStaticRepository()), "/api/quote")

	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d", rec.Code, http.StatusOK)
	}
	if ct := rec.Header().Get("Content-Type"); ct != "application/json; charset=utf-8" {
		t.Errorf("Content-Type = %q, want application/json; charset=utf-8", ct)
	}

	var got quote.Quote
	if err := json.Unmarshal(rec.Body.Bytes(), &got); err != nil {
		t.Fatalf("response body is not valid JSON: %v", err)
	}
	want := quote.Quote{
		Content: "A computer can never be held accountable, therefore a computer must never make a management decision. 🤖",
		Author:  "IBM Training Manual",
		Year:    "1979",
	}
	if got != want {
		t.Errorf("body = %+v, want %+v", got, want)
	}
}

// The JSON keys are the frontend's contract (loadQuote.ts reads content/author;
// year is reserved for future use). Assert the exact key set so a struct-tag
// change cannot silently break the website.
func TestGetQuoteJSONKeys(t *testing.T) {
	rec := doGet(newTestRouter(quote.NewStaticRepository()), "/api/quote")

	var raw map[string]json.RawMessage
	if err := json.Unmarshal(rec.Body.Bytes(), &raw); err != nil {
		t.Fatalf("invalid JSON: %v", err)
	}
	for _, key := range []string{"content", "author", "year"} {
		if _, ok := raw[key]; !ok {
			t.Errorf("response missing key %q; got %v", key, keysOf(raw))
		}
	}
	if len(raw) != 3 {
		t.Errorf("response has %d keys %v, want exactly 3", len(raw), keysOf(raw))
	}
}

// A repository failure must surface as HTTP 500 — not a panic and not an empty
// 200. The database-backed repository will depend on this behaviour.
func TestGetQuoteRepositoryError(t *testing.T) {
	rec := doGet(newTestRouter(failingRepo{}), "/api/quote")

	if rec.Code != http.StatusInternalServerError {
		t.Fatalf("status = %d, want %d", rec.Code, http.StatusInternalServerError)
	}
}

type failingRepo struct{}

func (failingRepo) Random(context.Context) (quote.Quote, error) {
	return quote.Quote{}, errors.New("boom")
}

func keysOf(m map[string]json.RawMessage) []string {
	ks := make([]string, 0, len(m))
	for k := range m {
		ks = append(ks, k)
	}
	return ks
}
