# Stage 1: Build Astro static assets
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

# Stage 2: Serve static assets via Caddy
FROM caddy:2.11.2-alpine

COPY Caddyfile /etc/caddy/Caddyfile
COPY --from=builder /app/dist /srv

EXPOSE 80
