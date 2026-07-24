// Package quote holds the quote domain model and the storage abstraction used
// to retrieve quotes. Keeping the model and the Repository interface here (and
// free of any HTTP or gin types) lets the HTTP layer and a future
// database-backed store depend on the same contract.
package quote

import "context"

// Quote is a single quotation as served by the API.
//
// The JSON tags are the public API contract consumed by the website
// (swa/src/scripts/loadQuote.ts). The fields also mirror the columns of the
// planned `quotes` PostgreSQL table, so the same struct can be scanned straight
// out of a row once a database-backed Repository lands.
type Quote struct {
	Content string `json:"content"`
	Author  string `json:"author"`
	Year    string `json:"year"`
}

// Repository is the source of quotes. StaticRepository is the only
// implementation today; a PostgresRepository will satisfy the same interface
// later. Every method takes a context.Context (for request cancellation and
// query deadlines) and returns an error (for I/O failures) so that swapping in
// a database implementation is a drop-in change — even though the static store
// can never block or fail.
type Repository interface {
	// Random returns a single quote to display. StaticRepository always
	// returns the same quote; a database-backed implementation will select a
	// random row (e.g. `ORDER BY random() LIMIT 1`).
	Random(ctx context.Context) (Quote, error)
}

// StaticRepository is a Repository that always returns one hard-coded quote.
// It lets the API ship before the database exists and can later serve as an
// offline fallback.
type StaticRepository struct {
	quote Quote
}

// NewStaticRepository returns a StaticRepository seeded with the default quote.
func NewStaticRepository() *StaticRepository {
	return &StaticRepository{
		quote: Quote{
			Content: "A computer can never be held accountable, therefore a computer must never make a management decision. 🤖",
			Author:  "IBM Training Manual",
			Year:    "1979",
		},
	}
}

// Random returns the single static quote. It never returns an error; the
// context is accepted only to satisfy the Repository contract.
func (r *StaticRepository) Random(context.Context) (Quote, error) {
	return r.quote, nil
}
