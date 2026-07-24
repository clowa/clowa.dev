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
# Stage 2: Runtime — s6-overlay supervising Caddy (and, later, the api)
#
# Multiple processes now share one container, so the container runtime can no
# longer restart a dead process for us. s6-overlay takes that role: it
# supervises each service, restarts supporting ones, and (via per-service finish
# scripts) tears the whole container down when a CRITICAL service dies.
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
# S6_CMD_WAIT_FOR_SERVICES* are left at their defaults: no supervised longrun has
# to be up before Caddy starts (the only one, api, ships disabled).

# --- Caddy configuration (split by purpose) ----------------------------------
# sites/     -> hosted content (the Astro site on clowa.dev)
# redirects/ -> domain->domain / path 301 redirects
# Kept in separate directories so the two concerns never bleed into each other.
COPY docker/caddy/ /etc/caddy/

# Astro static site, served from a per-site sub-directory so more sites (or the
# api's assets) can live side-by-side under /srv later.
COPY --from=builder /app/dist /srv/clowa.dev

# --- s6-overlay service definitions ------------------------------------------
# run/finish scripts must be executable; --chmod=0755 covers the whole tree
# (the data files it also touches don't mind being executable).
COPY --chmod=0755 docker/s6-overlay/s6-rc.d/ /etc/s6-overlay/s6-rc.d/

# No custom user bundle is shipped: Caddy is the CMD (not a supervised service)
# and the only defined service (api) ships DISABLED, so s6-overlay's default
# empty `user` bundle autostarts nothing. See docker/README.md to enable a
# supervised service later.

# Log directory for supervised services: s6-log runs as `nobody`, so it must own
# it; 0755 keeps rotated logs world-readable for a future monitoring agent.
# Caddy, being the CMD, logs to stdout/stderr instead.
RUN mkdir -p /var/log/api \
    && chown nobody:nobody /var/log/api \
    && chmod 0755 /var/log/api

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
