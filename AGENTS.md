# AGENTS.md

Mono-repo for the personal website **clowa.dev**, built and shipped as a **single container** on Scaleway Serverless Containers. Deployed with `osls` (an open-source Serverless Framework v3 fork) + the `serverless-scaleway-functions` plugin.

## Architecture

The container runs multiple processes under a **runit** supervision tree (`runsvdir` is PID 1 — see `Dockerfile` stage 2). Each service is `docker/runit/service/<name>/` with a `run` (foreground process, auto-restarted) and `log/run` (svlogd → `/var/log/<name>`). `STOPSIGNAL SIGHUP` lets `docker stop` tear the whole tree down cleanly.

- **caddy** — serves the Astro static site (`root * /srv/clowa.dev`) and performs 301 redirects. It also exports OpenTelemetry data (see Observability).
- **api** — Go backend, **not implemented yet**. Ships DISABLED via a `down` marker so runit won't autostart it; it will listen on loopback (`127.0.0.1:8080`) and Caddy will `reverse_proxy` to it once the binary exists. Will use PostgreSQL. No `api/` source dir yet.

Caddy config lives in `docker/caddy/`: `Caddyfile` imports `sites/*.caddy` (hosted content, e.g. `clowa.dev`) and `redirects/*.caddy` (domain→domain 301s). Redirect source hosts must also be attached as `custom_domains` in `serverless.yml` to reach Caddy. **The root `Caddyfile` is legacy and unused** — edit the ones under `docker/caddy/`.

## Layout

- `swa/` — the Astro app. **All npm/site work happens here**, not the repo root.
- `docker/` — container runtime config: `caddy/` (split Caddyfile + `sites/`/`redirects/`) and `runit/service/` (per-process supervision).
- `Dockerfile` (root) builds the image: an Astro build stage, then a runit + Caddy runtime stage. `serverless.yml` deploys the prebuilt image to Scaleway.
- Root `package.json` holds only the deploy tooling (`osls`, `serverless-scaleway-functions`) — no app code.

## Commands (prefer the Taskfile — `taskfile.yaml`)

- `task dev` — Astro dev server on `http://localhost:4321` (runs in `swa/`).
- `task build` — `npm ci` + `npm run build` in `swa/` → `swa/dist/`.
- `task docker-build` — `docker buildx build --platform linux/amd64` + push to the Scaleway registry (tags: short SHA + `latest`).
- `task deploy` — runs `docker-build`, then `serverless deploy` via 1Password (`op run --env-file ./.env.1password`), passing the image ref as `REGISTRY_IMAGE`.
- `task init-devcontainer` — run once in a fresh devcontainer/codespace (fixes `swa/node_modules` perms + `npm install`).
- In `swa/`: `npm run build` = `astro check && astro build` (build is also the type-check); `npm run typecheck` is `astro check` alone. Quality tooling: `npm run lint` (ESLint + Stylelint), `npm run lint:fix`, `npm run format` / `format:check` (Prettier), `npm run lint:md` (markdownlint, repo-wide).

## Conventions

- **Formatting** is owned by Prettier (`swa/.prettierrc.mjs`, `prettier-plugin-astro`) for `.astro`/`.ts`/`.css`/`.json`/`.yaml`. Prettier ignores `*.md`.
- **Linting** is correctness-only — stylistic rules are turned off via `eslint-config-prettier` (loaded last). ESLint flat config (`swa/eslint.config.mjs`): `typescript-eslint` + `eslint-plugin-astro`. Stylelint: `swa/.stylelintrc.json`. Astro **a11y rules are intentionally off** for now (blocked on `eslint-plugin-jsx-a11y` supporting ESLint 10).
- **Markdown** is owned entirely by markdownlint (`.markdownlint.yaml`; MD013 off; `ul` markers 1-space single / 2-space multi).
- **Pre-commit**: `lint-staged` config exists (`swa/.lintstagedrc.json`), but **no local git hook is currently wired** — `swa/package.json`'s `prepare` points at a not-yet-present `scripts/setup-hooks.mjs`, so it is a no-op. **CI is the enforced quality gate.**

## CI (`.github/workflows/`)

- `ci.yaml` (`quality` job; on PRs + push to `main`; Node 22): Prettier `--check`, ESLint + Stylelint, markdownlint, `astro check`.
- `deploy.yaml` (push to `main`): a `build` job pushes an `amd64` image via `docker/build-push-action` (GHA cache), then a `deploy` job runs `serverless deploy` (osls) with the image ref + Scaleway/Grafana secrets.
- `cleanup-registry.yaml` (weekly Sun 03:00 UTC + manual): prunes old registry tags via the `scaleway-registry-cleanup` composite action (keeps last 10; `latest` protected).

## Observability (OpenTelemetry → Grafana Cloud)

TODO

## Lessons Learned

Pitfalls and hard-won knowledge live in [LESSONS_LEARNED.md](./LESSONS_LEARNED.md).
