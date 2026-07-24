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
└── s6-overlay/s6-rc.d/        -> /etc/s6-overlay/s6-rc.d/   (supporting services)
    └── api/ + api-log/        the future Go API + its log pipeline (DISABLED)
```

Caddy is **not** an s6-rc service — it is the `CMD`, so there is no `caddy/`
service directory. Autostart of supervised services is governed by s6-overlay's
default (empty) `user` bundle; nothing autostarts today (the only defined
service, `api`, ships disabled).

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
- **Supporting supervised services** (e.g. a future otel-collector) — ship **without** a `finish` script, so s6-supervise just **restarts** them in place.

## Logging

- **Caddy** logs to **stdout/stderr** → `docker logs` / Scaleway Cockpit (the standard `CMD` pattern).
- **Supervised services** each have a dedicated `s6-log` pipeline (`producer-for` / `consumer-for`, grouped into a `*-pipeline` bundle via `pipeline-name`) that writes a rotating `current` file as `nobody`:

| Service | Log file               |
| ------- | ---------------------- |
| api     | `/var/log/api/current` |

Rotation keeps 20 archives, rolling at ~1 MiB (`n20 s1000000`). The directory is created + chowned in the Dockerfile.

## Resource limits (prepared, disabled)

- **Caddy (`CMD`)** — bounded by the container/cgroup limits (`cpuLimit` / `memoryLimit` in [`serverless.yml`](../serverless.yml)). To apply per-process rlimits instead, wrap the `CMD` with `s6-softlimit` (a commented example sits next to the `CMD` in the Dockerfile).
- **Supervised services** — each `run` script has a commented `s6-softlimit` wrapper (POSIX rlimits — effective for memory on single-process daemons; there is no rlimit for CPU share, so use container limits for CPU).

## Enabling the `api` (when the binary exists)

The `api` is defined but **not** started — it is absent from the `user` bundle. (A `down` file in the source dir does **not** work; s6-rc-compile ignores it.) To enable it:

1. Add the binary to the image at `/usr/local/bin/api`.
2. Create the autostart marker and ship it: add `docker/s6-overlay/user-bundles.d/user/type` (containing `bundle`) + an empty `docker/s6-overlay/user-bundles.d/user/contents.d/api-pipeline`, and `COPY docker/s6-overlay/user-bundles.d/ /etc/s6-overlay/user-bundles.d/` in the Dockerfile. (For a quick runtime test: `s6-rc -u change api-pipeline`.)
3. Add a `reverse_proxy 127.0.0.1:8080` route in the Caddy config.

## Adding a supporting (restart-only) service

Create `s6-rc.d/<name>/` with `type` (`longrun`) + `run`, **omit** `finish`, add a log pipeline like `api-log`, and register it for autostart via `user-bundles.d/user/contents.d/<name>` (use `<name>-pipeline` if the service is logged) — see the `api` enable steps for the bundle wiring.
