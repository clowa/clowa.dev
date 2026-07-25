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
├── otel.go                     OpenTelemetry SDK bootstrap (traces, metrics, logs)
├── logging.go                  slog logger: JSON stdout + OTLP bridge
└── internal/
    ├── quote/                  the quote domain
    │   ├── quote.go            Quote model + Repository interface + StaticRepository
    │   └── handler.go          gin handler for /api/quote
    └── server/
        ├── server.go           assembles the gin engine (middleware + routes)
        └── logging.go          slog request-logging middleware (trace-correlated)
```

The `Repository` interface is the seam for the planned move to PostgreSQL: it
takes a `context.Context` and returns an `error`, so a `PostgresRepository` can
replace `StaticRepository` in `main.go` with no other changes. The `Quote`
fields mirror the future `quotes` table columns.

## Configuration

| Variable                      | Default             | Purpose                                                |
| ----------------------------- | ------------------- | ------------------------------------------------------ |
| `API_ADDR`                    | `127.0.0.1:8080`    | Listen address.                                        |
| `GIN_MODE`                    | `release` (image)   | gin mode; set `debug` for local dev.                   |
| `OTEL_EXPORTER_OTLP_ENDPOINT` | _unset_             | OTLP endpoint; **unset disables all telemetry.**       |
| `OTEL_SERVICE_NAME`           | `quote-api` (image) | Service name reported in telemetry.                    |

Telemetry is configured entirely from the standard `OTEL_*` env vars (endpoint,
protocol, headers, resource attributes, sampler) — see the repo root `AGENTS.md`
→ Observability for the full picture.

## Observability

The api is instrumented with OpenTelemetry (see [`otel.go`](otel.go) and
[`logging.go`](logging.go)): request traces + HTTP metrics via `otelgin`, Go
runtime metrics, and structured logs through `log/slog`. Logs go to two sinks at
once — **JSON on stdout** and the **OTLP log bridge** — and each request log is
correlated to its trace (`trace_id` / `span_id`). Everything is driven by the
standard `OTEL_*` env vars and is a **no-op when `OTEL_EXPORTER_OTLP_ENDPOINT` is
unset**, so local runs stay quiet. In the container these point at the
in-container otel-collector, which forwards to Middleware.

## Development

```sh
cd api
go test ./...        # run the test suite
go run .             # serve on 127.0.0.1:8080
curl 127.0.0.1:8080/api/quote
```

**Test-driven development.** Tests are written first as acceptance criteria, then
the implementation is written against them; the HTTP handler is exercised with
`net/http/httptest`, so no socket is bound during tests. New behaviour — including
the OpenTelemetry wiring — follows the same red→green flow (e.g. `otel_test.go`,
`logging_test.go`).
