# Stage 1: Build Astro static assets
FROM node:22-alpine AS builder

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

# Stage 2: Serve static assets via Caddy
FROM caddy:2.11.2-alpine

COPY Caddyfile /etc/caddy/Caddyfile
COPY --from=builder /app/dist /srv

EXPOSE 80
