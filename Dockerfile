# syntax=docker/dockerfile:1

# ==============================================================================
# Stage 1: Build Astro static assets
# ==============================================================================
FROM node:22-alpine AS builder

# brotli needed for pre-compression of static assets
RUN apk add --no-cache brotli

WORKDIR /app

RUN --mount=type=bind,source=swa/package.json,target=package.json \
    --mount=type=bind,source=swa/package-lock.json,target=package-lock.json \
    --mount=type=cache,target=/root/.npm \
    npm ci --no-audit --no-fund

ARG PUBLIC_RYBBIT_SITE_ID
ENV PUBLIC_RYBBIT_SITE_ID=${PUBLIC_RYBBIT_SITE_ID}

RUN --mount=type=bind,source=swa/src,target=src \
    --mount=type=bind,source=swa/public,target=public \
    --mount=type=bind,source=swa/astro.config.mjs,target=astro.config.mjs \
    --mount=type=bind,source=swa/tsconfig.json,target=tsconfig.json \
    --mount=type=bind,source=swa/package.json,target=package.json \
    npm run build

# Pre-compress text-based assets; originals are kept as fallback for clients
# that don't support compression. Binary formats (images, fonts) are skipped
# since they are already compressed and would only grow in size.
RUN find /app/dist -type f \( \
        -name "*.html" -o -name "*.css" -o -name "*.js"  -o \
        -name "*.svg"  -o -name "*.xml" -o -name "*.json" -o \
        -name "*.txt"  -o -name "*.webmanifest" \
    \) | xargs -P4 -I{} sh -c 'brotli --best --keep "$1" && gzip -9 -k "$1"' _ {}

# ==============================================================================
# Stage 2: Build the Go API binary
#
# Runs on $BUILDPLATFORM (the host arch) and cross-compiles to $TARGETARCH via
# Go's native GOOS/GOARCH, so an arm64 laptop still emits the amd64 binary
# Scaleway needs — no QEMU. CGO is disabled so the result is a static binary
# that runs on the musl-based caddy:alpine runtime with no libc dependency.
# ==============================================================================
FROM --platform=$BUILDPLATFORM golang:1.26-alpine AS api-builder

WORKDIR /src

# Download modules first, keyed only on go.mod/go.sum, so the (cached) module
# download is reused whenever only Go source — not dependencies — changes.
RUN --mount=type=bind,source=api/go.mod,target=go.mod \
    --mount=type=bind,source=api/go.sum,target=go.sum \
    --mount=type=cache,target=/go/pkg/mod \
    go mod download

# TARGETOS/TARGETARCH are provided by BuildKit; they drive cross-compilation.
ARG TARGETOS
ARG TARGETARCH

# -trimpath strips local paths from the binary; -s -w drops the symbol table and
# DWARF debug info to shrink it. The source is bind-mounted read-only; the module
# and build caches are reused across builds.
RUN --mount=type=bind,source=api,target=/src \
    --mount=type=cache,target=/go/pkg/mod \
    --mount=type=cache,target=/root/.cache/go-build \
    CGO_ENABLED=0 GOOS=${TARGETOS} GOARCH=${TARGETARCH} \
    go build -trimpath -ldflags="-s -w" -o /out/api .

# ==============================================================================
# Stage 3: Build a minimal OpenTelemetry Collector
#
# The OpenTelemetry Collector Builder (ocb) compiles a collector containing only
# the components this container needs (see docker/otel-collector/builder-config.yaml)
# instead of shipping the ~200 MB contrib distribution. Cross-compiled like the api
# stage: runs on $BUILDPLATFORM, emits a static $TARGETARCH binary, CGO disabled.
# ==============================================================================
FROM --platform=$BUILDPLATFORM golang:1.26-alpine AS otelcol-builder

# Keep ocb pinned to the collector release the manifest's components target.
ARG OCB_VERSION=0.157.0

WORKDIR /build

# Install the builder once; its own module downloads share the build cache.
RUN --mount=type=cache,target=/go/pkg/mod \
    --mount=type=cache,target=/root/.cache/go-build \
    go install "go.opentelemetry.io/collector/cmd/builder@v${OCB_VERSION}"

# TARGETOS/TARGETARCH drive cross-compilation of the generated collector. -s -w
# strips the symbol table and DWARF info to shrink the binary (as in the api stage).
# The manifest's output_path is ./_build, so the binary lands at /build/_build/otelcol.
ARG TARGETOS
ARG TARGETARCH
RUN --mount=type=bind,source=docker/otel-collector/builder-config.yaml,target=/build/builder-config.yaml \
    --mount=type=cache,target=/go/pkg/mod \
    --mount=type=cache,target=/root/.cache/go-build \
    CGO_ENABLED=0 GOOS=${TARGETOS} GOARCH=${TARGETARCH} \
    builder --config /build/builder-config.yaml --ldflags="-s -w"

# ==============================================================================
# Stage 4: Runtime — Caddy (the CMD) + the s6-overlay-supervised api & collector
#
# Several processes share one container, so the container runtime cannot restart
# a dead process for us. Caddy runs as the CMD (its exit stops the container);
# s6-overlay supervises the api and the otel-collector, restarting supporting
# ones (the collector) and — via per-service finish scripts — halting the whole
# container when a CRITICAL supervised service (the api) dies.
#
# 2.11.3+ is required for native OTLP metrics push (`metrics { otlp }`, PR #7664);
# tracing has been available since 2.5. Both are configured via OTEL_* env vars.
# ==============================================================================
FROM caddy:2.11.4-alpine

# --- s6-overlay installation -------------------------------------------------
# Pin the version; see https://github.com/just-containers/s6-overlay/releases.
ARG S6_OVERLAY_VERSION=3.2.3.2
# Provided automatically by BuildKit (amd64 in CI/prod, arm64 for local builds).
ARG TARGETARCH

# Install the noarch overlay + the arch-specific binaries, verifying checksums.
# `xz` is only needed to unpack the .tar.xz archives, so it is added as a
# virtual package and removed again in the same layer to keep the image small.
RUN set -eux; \
    apk add --no-cache --virtual .s6-deps xz; \
    case "${TARGETARCH}" in \
      amd64) s6_arch=x86_64 ;; \
      arm64) s6_arch=aarch64 ;; \
      arm)   s6_arch=arm ;; \
      386)   s6_arch=i686 ;; \
      *) echo "unsupported TARGETARCH: ${TARGETARCH}" >&2; exit 1 ;; \
    esac; \
    base="https://github.com/just-containers/s6-overlay/releases/download/v${S6_OVERLAY_VERSION}"; \
    cd /tmp; \
    for f in "s6-overlay-noarch.tar.xz" "s6-overlay-${s6_arch}.tar.xz"; do \
      wget -q -O "${f}"        "${base}/${f}"; \
      wget -q -O "${f}.sha256" "${base}/${f}.sha256"; \
      sha256sum -c "${f}.sha256"; \
      tar -C / -Jxpf "${f}"; \
      rm "${f}" "${f}.sha256"; \
    done; \
    apk del .s6-deps

# --- s6-overlay behaviour ----------------------------------------------------
# Halt the container if a oneshot/init step fails during startup (fail fast).
ENV S6_BEHAVIOUR_IF_STAGE2_FAILS=2
# Deliver `docker stop`'s SIGTERM straight to Caddy (the CMD) so it shuts down
# gracefully; s6-overlay then tears down the rest of the tree.
ENV S6_CMD_RECEIVE_SIGNALS=1
# S6_CMD_WAIT_FOR_SERVICES* are left at their defaults: Caddy need not wait for
# the api to be up before starting — until the api is ready Caddy just returns
# 502 for /api/*, which the website's loadQuote.ts already handles gracefully.

# gin runs in release mode in the image (no debug logging / startup warnings);
# with-contenv makes this visible to the supervised api process. Override for
# local debugging by setting GIN_MODE=debug.
ENV GIN_MODE=release

# --- Caddy configuration (split by purpose) ----------------------------------
# sites/     -> hosted content (the Astro site on clowa.dev)
# redirects/ -> domain->domain / path 301 redirects
# Kept in separate directories so the two concerns never bleed into each other.
COPY docker/caddy/ /etc/caddy/

# Astro static site, served from a per-site sub-directory so additional sites can
# live side-by-side under /srv.
COPY --from=builder /app/dist /srv/clowa.dev

# The Go api binary, supervised by s6-overlay and reverse-proxied by Caddy on
# 127.0.0.1:8080. Referenced by docker/s6-overlay/s6-rc.d/api/run.
COPY --from=api-builder /out/api /usr/local/bin/api

# The minimal otel-collector binary + its config, supervised by s6-overlay as a
# supporting (restart-only) service. It is the single telemetry egress to
# Middleware. Referenced by docker/s6-overlay/s6-rc.d/otel-collector/run.
COPY --from=otelcol-builder /build/_build/otelcol /usr/local/bin/otelcol
COPY docker/otel-collector/config.yaml /etc/otel-collector/config.yaml

# --- s6-overlay service definitions ------------------------------------------
# run/finish scripts must be executable; --chmod=0755 covers the whole tree
# (the data files it also touches don't mind being executable).
COPY --chmod=0755 docker/s6-overlay/s6-rc.d/ /etc/s6-overlay/s6-rc.d/

# User bundle: the set of services s6-overlay autostarts. It lists the
# api-pipeline (api + its log consumer) and the otel-collector-pipeline (collector
# + its log consumer), so both start with the container. Caddy is the CMD (not a
# supervised service), so it is not listed here.
COPY docker/s6-overlay/user-bundles.d/ /etc/s6-overlay/user-bundles.d/

# Log directory for the supervised api: s6-log runs as `nobody`, so it must own
# it; 0755 keeps rotated logs world-readable for the otel-collector's filelog
# receiver.
RUN mkdir -p /var/log/api \
    && chown nobody:nobody /var/log/api \
    && chmod 0755 /var/log/api

# Log directory for Caddy's JSON access/runtime logs (see docker/caddy/Caddyfile).
# Caddy runs as root (the CMD), so root ownership is fine; the otel-collector
# (also root) tails these files and forwards them.
RUN mkdir -p /var/log/caddy \
    && chmod 0755 /var/log/caddy

# Log directory for the supervised otel-collector: its s6-log runs as `nobody`.
RUN mkdir -p /var/log/otel-collector \
    && chown nobody:nobody /var/log/otel-collector \
    && chmod 0755 /var/log/otel-collector

EXPOSE 80

# /init is s6-overlay's PID 1: it reaps zombies, runs the supervision tree, and
# performs an orderly shutdown. Must be exec form.
ENTRYPOINT ["/init"]

# Caddy is the container's primary process (s6-overlay "CMD" pattern): /init
# starts the supervision tree, then launches this. If Caddy exits, s6-overlay
# brings the whole container down with Caddy's exit code — no finish script
# needed. Caddy logs to stdout/stderr (docker logs / Scaleway).
#
# `with-contenv` is required: s6-overlay resets the environment before the CMD,
# so without it Caddy would not see the runtime OTEL_* vars (set in
# serverless.yml) it needs to export telemetry.
#
# To cap Caddy with per-process rlimits instead of the container's cpu/memory
# limits, insert s6-softlimit after with-contenv, e.g.:
#   CMD ["/command/with-contenv","s6-softlimit","-m","134217728","caddy","run","--config","/etc/caddy/Caddyfile","--adapter","caddyfile"]
CMD ["/command/with-contenv", "caddy", "run", "--config", "/etc/caddy/Caddyfile", "--adapter", "caddyfile"]
