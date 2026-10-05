# AGENTS.md

Mono-repo for the personal website **clowa.dev**, built and shipped as a **single container** on Scaleway Serverless Containers. Deployed with `osls` (an open-source Serverless Framework v3 fork) + the `serverless-scaleway-functions` plugin.

## Architecture

The container runs **Caddy as its `CMD`** (the primary process) under **s6-overlay** (`/init`, PID 1), which also supervises the supporting services. Following the [s6-overlay "Usage" pattern](https://github.com/just-containers/s6-overlay#usage), the main daemon runs as `CMD`: when it exits, s6-overlay tears the whole container down and the container exits with Caddy's exit code.

- **caddy** — the container's `CMD`/primary process: serves the Astro static site (`root * /srv/clowa.dev`), performs 301 redirects, and exports OpenTelemetry data (see Observability). Logs to stdout/stderr (`docker logs` / Scaleway).
- **api** — Go backend (gin), source in [`api/`](./api/). A **supervised, critical** s6-rc service, autostarted via the s6 `user` bundle. Listens on loopback (`127.0.0.1:8080`); Caddy `reverse_proxy`es `/api/*` to it. Serves `GET /api/quote` (a static quote) today; quotes will move to PostgreSQL behind the existing `quote.Repository` interface. See [`api/README.md`](./api/README.md).

**Failure policy**: **Caddy** is the `CMD`, so if it dies the container exits with its code and the orchestrator reschedules it. Among the **supervised** services, critical ones (`api`) carry a `finish` script that halts the whole container on exit; supporting ones (e.g. a future otel-collector) omit `finish` and are restarted in place. Supervised services log to their own rotating file (`/var/log/<svc>/current`) via an `s6-log` pipeline; Caddy logs to stdout.

Caddy config lives in `docker/caddy/`: `Caddyfile` imports `sites/*.caddy` (hosted content, e.g. `clowa.dev`) and `redirects/*.caddy` (domain→domain 301s). Redirect source hosts must also be attached as `custom_domains` in `serverless.yml` to reach Caddy. See [`docker/README.md`](./docker/README.md) for the full container-runtime layout.

## Layout

- `swa/` — the Astro app. **All npm/site work happens here**, not the repo root.
- `api/` — self-contained Go module; `go test ./...` / `go run .` run from inside `api/`. See [`api/README.md`](./api/README.md).
- `docker/` — container runtime config. See [`docker/README.md`](./docker/README.md).

## Commands (prefer the Taskfile — `taskfile.yaml`)

- `task dev` — Astro dev server on `http://localhost:4321` (runs in `swa/`).
- `task build` — `npm ci` + `npm run build` in `swa/` → `swa/dist/`.
- `task docker-build` — `docker buildx build --platform linux/amd64` + push to the Scaleway registry (tags: short SHA + `latest`).
- `task deploy` — runs `docker-build`, then `serverless deploy` via 1Password (`op run --env-file ./.env.1password`), passing the image ref as `REGISTRY_IMAGE`.
- `task init-devcontainer` — run once in a fresh devcontainer/codespace (fixes `swa/node_modules` perms + `npm install`).
- In `swa/`: `npm run build` = `astro check && astro build` (the build is also the type-check).

## Development conventions

- **Markdown** is owned entirely by markdownlint (`.markdownlint.yaml`; MD013 off; `ul` markers 1-space single / 2-space multi).

## Observability (OpenTelemetry → Middleware)

Every process is instrumented with **vendor-neutral OpenTelemetry** and exports to **[Middleware](https://cedric-ahlers.middleware.io)**. Standard `OTEL_*` env vars keep the backend swappable.

**Egress model.** An **in-container OpenTelemetry Collector** (a supporting, restart-only s6 service — see [`docker/otel-collector/`](./docker/otel-collector/)) is the **single egress**: Caddy and the api send OTLP to it on loopback (`127.0.0.1:4317`, gRPC), and it forwards traces/metrics/logs to Middleware. The collector is the **only holder of the Middleware API key**, so the producers stay agentless and secret-free. It is built minimally with the OpenTelemetry Collector Builder ([`builder-config.yaml`](./docker/otel-collector/builder-config.yaml)) rather than shipping the contrib distribution.

**Services** (both under `service.namespace=clowa`, env in `deployment.environment.name`):

| Service | Process | Signals |
| --- | --- | --- |
| `web-proxy` | Caddy | Traces (`tracing` on every site + redirect), metrics (`metrics { otlp }`), JSON logs (file → collector `filelog`, since Caddy has no native OTLP log export) |
| `quote-api` | Go api | Traces + metrics (otelgin) + Go runtime metrics, and structured logs via `slog` → OTLP (see [`api/otel.go`](./api/otel.go), [`api/logging.go`](./api/logging.go)) |

**Config.** Non-secret `OTEL_*` + `MIDDLEWARE_OTLP_ENDPOINT` live in [`serverless.yml`](./serverless.yml); the api overrides `OTEL_SERVICE_NAME=quote-api` in its s6 `run`. The `MIDDLEWARE_API_KEY` secret is a Scaleway extra-secure env var, sourced from the `MIDDLEWARE_API_KEY` GitHub Actions secret (CI) or 1Password (`task deploy`). Any log not emitted natively as an OTEL log (Caddy's access + runtime logs) is written as **JSON** to files the collector tails.

**Local dev.** With `OTEL_EXPORTER_OTLP_ENDPOINT` unset, the api's telemetry is a no-op and Caddy's export fails quietly in the background; the collector only runs in the image. To run the full container locally, see the `run-container-locally` skill.

## Lessons Learned

Pitfalls and hard-won knowledge live in [LESSONS_LEARNED.md](./LESSONS_LEARNED.md).
