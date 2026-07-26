# Container runtime config

Everything under `docker/` is baked into the runtime image by the root
[`Dockerfile`](../Dockerfile). The image follows the
[s6-overlay "Usage" pattern](https://github.com/just-containers/s6-overlay#usage):
**Caddy runs as the container's `CMD`** (the primary process), while `/init`
(s6-overlay, PID 1) supervises any **supporting** services. Config is split
cleanly by purpose:

```text
docker/
├── caddy/                     -> /etc/caddy/            (Caddy configuration)
│   ├── Caddyfile              global options + imports of the two dirs below
│   ├── sites/                 HOSTED CONTENT — one file per served site
│   │   └── clowa.dev.caddy    the Astro static site (root /srv/clowa.dev)
│   └── redirects/             REDIRECTS — one file per source host
│       ├── qunis.day.caddy        301 qunis.day     -> qunisday.qunis.de
│       ├── me.clowa.de.caddy      301 me.clowa.de   -> clowa.dev
│       └── www.clowa.dev.caddy    301 www.clowa.dev -> clowa.dev
├── otel-collector/            -> /etc/otel-collector/   (OpenTelemetry Collector)
│   ├── config.yaml           receivers/processors/exporters (single egress to Middleware)
│   └── builder-config.yaml    ocb manifest — builds the minimal collector binary
└── s6-overlay/                -> /etc/s6-overlay/          (supervision tree)
    ├── s6-rc.d/               service definitions
    │   ├── api/ + api-log/            the Go API + its log pipeline
    │   └── otel-collector/ + …-log/   the collector + its log pipeline
    └── user-bundles.d/        autostart set — `api-pipeline` + `otel-collector-pipeline`
```

Caddy is **not** an s6-rc service — it is the `CMD`, so there is no `caddy/`
service directory. Autostart of supervised services is governed by the `user`
bundle (`user-bundles.d/user/contents.d/`); it lists `api-pipeline`, so the
`api` (and its log consumer) starts with the container.

## Caddy: sites vs. redirects

`Caddyfile` sets global options and then `import`s `sites/*.caddy` and `redirects/*.caddy`. The two directories keep the concerns separate on disk:

- **`sites/`** — serving hosted content. `clowa.dev.caddy` binds to `:80` with no host filter, so it is the **default** handler (answers `clowa.dev`, the raw Scaleway URL, the TCP health check and local `curl localhost`).
- **`redirects/`** — one file per redirect source host, each an explicit `<host>:80` block. Binding to port 80 (not a bare hostname, which would default to `:443` and be unreachable behind Scaleway) makes these host-specific blocks win over the `:80` catch-all. All redirects are **301 permanent** and preserve path + query via `{uri}`.

### Redirect kinds

- **Domain → domain** — the three files in `redirects/`.
- **Path** (same host, e.g. `/home` → `/home/my`) — belongs with the site it redirects within; add `redir` rules in `sites/clowa.dev.caddy` (commented example included there).
- **HTTP → HTTPS** — handled at the Scaleway edge (`httpOption: redirected` in `serverless.yml`); Caddy only ever speaks plain HTTP on `:80`, and `auto_https off` keeps it from attempting ACME behind the proxy.

> **Production routing:** a redirect source host only reaches Caddy if it is also attached as a `custom_domain` in [`serverless.yml`](../serverless.yml) with DNS pointed at Scaleway. See the commented entries there.

## Process model & failure policy

A single container can't rely on the runtime to restart a dead process, so the policy is encoded per process:

- **Caddy (`CMD`, primary)** — launched by `/init` after the supervision tree is up. If Caddy exits for any reason, s6-overlay tears the whole container down and the container exits **with Caddy's exit code**, so the orchestrator reschedules it. This is the native s6-overlay `CMD` behavior — no `finish` script needed.
- **Critical supervised services** (`api`) — a `finish` script writes the exit code to `/run/s6-linux-init-container-results/exitcode` and runs `/run/s6/basedir/bin/halt`, so **the whole container exits** on failure too.
- **Supporting supervised services** (the `otel-collector`) — ship **without** a `finish` script, so s6-supervise just **restarts** them in place. Telemetry is not container-critical, so a collector crash must not take the container down.

## Logging

All logs ultimately reach **Middleware** via the `otel-collector` (see the repo root `AGENTS.md` → Observability). How each process gets them there differs:

- **Caddy** has no native OTLP log export, so it writes **JSON** to files the collector tails via its `filelog` receiver: `/var/log/caddy/access.log` (per-site access logs, from the `(otel)` snippet) and `/var/log/caddy/runtime.log` (the default logger). This routes Caddy's logs **out of `docker logs`/Scaleway Cockpit**; the rolled files remain on disk as a local fallback.
- **api** emits logs **natively over OTLP** (via `slog`) straight to the collector, and mirrors them as **JSON on stdout** — captured by its `s6-log` pipeline as a fallback (not tailed by the collector, to avoid double-sending).
- **Supervised services** each have a dedicated `s6-log` pipeline (`producer-for` / `consumer-for`, grouped into a `*-pipeline` bundle via `pipeline-name`) that writes a rotating `current` file as `nobody`:

| Service          | s6-log file                      |
| ---------------- | -------------------------------- |
| api              | `/var/log/api/current`           |
| otel-collector   | `/var/log/otel-collector/current`|

Rotation keeps 20 archives, rolling at ~1 MiB (`n20 s1000000`). Log directories are created (and, for `nobody`-owned ones, chowned) in the Dockerfile.

## Resource limits (prepared, disabled)

- **Caddy (`CMD`)** — bounded by the container/cgroup limits (`cpuLimit` / `memoryLimit` in [`serverless.yml`](../serverless.yml)). To apply per-process rlimits instead, wrap the `CMD` with `s6-softlimit` (a commented example sits next to the `CMD` in the Dockerfile).
- **Supervised services** — each `run` script has a commented `s6-softlimit` wrapper (POSIX rlimits — effective for memory on single-process daemons; there is no rlimit for CPU share, so use container limits for CPU).

## Running as non-root (deferred)

The `api` and `otel-collector` longruns currently **run as root** (bare `exec` in their `run` scripts); only the `s6-log` sinks already drop to `nobody` (see [Logging](#logging)). Deferred because production runs in a Scaleway microVM (`sandbox: v2` in [`serverless.yml`](../serverless.yml)), which isolates the container regardless.

- **api** has no reason to be root — wrap its `run` with `s6-setuidgid nobody` to drop it.
- **otel-collector** runs as root specifically to read Caddy's root-owned `/var/log/caddy/*.log` (noted in its `run` header). Dropping it first requires making those files group-readable (adjust the `/var/log/caddy` `mkdir`/`chmod`/`chown` in the Dockerfile).

## The `api` service

The `api` (Go source in [`../api`](../api), binary at `/usr/local/bin/api`) is a
**critical, supervised** service. Its wiring:

1. The Dockerfile builds the binary in a cross-compiling Go stage and `COPY`s it
   to `/usr/local/bin/api`.
2. The `user` bundle (`user-bundles.d/user/`, a `bundle` with a
   `contents.d/api-pipeline` marker) is `COPY`ed to `/etc/s6-overlay/`, telling
   s6-overlay to start the `api-pipeline` (the `api` longrun + its `api-log`
   consumer) with the container. Bundle membership is the on/off switch — an
   s6-rc.d `down` file does **not** work, as s6-rc-compile ignores it.
3. `docker/caddy/sites/clowa.dev.caddy` `reverse_proxy`es `/api/*` to
   `127.0.0.1:8080`.

## Adding a supporting (restart-only) service

Create `s6-rc.d/<name>/` with `type` (`longrun`) + `run`, **omit** `finish`, add a log pipeline like `api-log`, and register it for autostart via `user-bundles.d/user/contents.d/<name>` (use `<name>-pipeline` if the service is logged) — see the `api` enable steps for the bundle wiring.
