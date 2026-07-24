package quote

import (
	"context"
	"testing"
)

// The static repository must return exactly the documented default quote —
// this is the acceptance criterion for the minimal API.
func TestStaticRepositoryRandom(t *testing.T) {
	repo := NewStaticRepository()

	got, err := repo.Random(context.Background())
	if err != nil {
		t.Fatalf("Random() returned unexpected error: %v", err)
	}

	want := Quote{
		Content: "A computer can never be held accountable, therefore a computer must never make a management decision. 🤖",
		Author:  "IBM Training Manual",
		Year:    "1979",
	}
	if got != want {
		t.Errorf("Random() = %+v, want %+v", got, want)
	}
}

// Unlike the future database-backed store, the static repository must be
// deterministic: repeated calls return the identical quote.
func TestStaticRepositoryRandomIsStable(t *testing.T) {
	repo := NewStaticRepository()

	first, err := repo.Random(context.Background())
	if err != nil {
		t.Fatalf("first Random() error: %v", err)
	}
	second, err := repo.Random(context.Background())
	if err != nil {
		t.Fatalf("second Random() error: %v", err)
	}

	if first != second {
		t.Errorf("Random() is not stable: %+v != %+v", first, second)
	}
}

// StaticRepository must satisfy the Repository interface; a compile-time check
// guards against the signature drifting away from the contract.
var _ Repository = (*StaticRepository)(nil)
