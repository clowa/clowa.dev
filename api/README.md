# api

The clowa.dev backend: a small [gin](https://gin-gonic.com/) HTTP server. It
listens on loopback only (`127.0.0.1:8080`) and runs behind Caddy, which
reverse-proxies `/api/*` to it (see
[`docker/caddy/sites/clowa.dev.caddy`](../docker/caddy/sites/clowa.dev.caddy)).
Inside the container it is a **critical, supervised** s6-overlay service (see
[`docker/README.md`](../docker/README.md)).

## Endpoints

| Method | Path         | Description                     |
| ------ | ------------ | ------------------------------- |
| GET    | `/api/quote` | Returns a single quote as JSON. |

`GET /api/quote` response:

```json
{
  "content": "A computer can never be held accountable, therefore a computer must never make a management decision. 🤖",
  "author": "IBM Training Manual",
  "year": "1979"
}
```

`content` and `author` are consumed by the website
([`swa/src/scripts/loadQuote.ts`](../swa/src/scripts/loadQuote.ts)); `year` is
reserved for future use.

## Layout

```text
api/
├── main.go                     process entrypoint: config, signals, listener
└── internal/
    ├── quote/                  the quote domain
    │   ├── quote.go            Quote model + Repository interface + StaticRepository
    │   └── handler.go          gin handler for /api/quote
    └── server/
        └── server.go           assembles the gin engine (middleware + routes)
```

The `Repository` interface is the seam for the planned move to PostgreSQL: it
takes a `context.Context` and returns an `error`, so a `PostgresRepository` can
replace `StaticRepository` in `main.go` with no other changes. The `Quote`
fields mirror the future `quotes` table columns.

## Configuration

| Variable   | Default           | Purpose                                  |
| ---------- | ----------------- | ---------------------------------------- |
| `API_ADDR` | `127.0.0.1:8080`  | Listen address.                          |
| `GIN_MODE` | `release` (image) | gin mode; set `debug` for local dev.     |

## Development

```sh
cd api
go test ./...        # run the test suite
go run .             # serve on 127.0.0.1:8080
curl 127.0.0.1:8080/api/quote
```

Tests are written first as acceptance criteria (TDD); the HTTP handler is
exercised with `net/http/httptest`, so no socket is bound during tests.
