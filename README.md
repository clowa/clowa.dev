# Overview

This is my personal website project. It is built using [Astro](https://astro.build/) and hosted as a monolithic container on [Scaleway Cloud](https://www.scaleway.com). Go and check it out at [https://clowa.dev](https://clowa.dev).

## Repository Structure

- `swa/` — Frontend code of the website (Astro)
- `api/` — Go backend service (serves `/api/*`, e.g. `GET /api/quote`)
- `docker/` - Configuration of the monolithic docker container

## Getting Started

[![Open in GitHub Codespaces](https://github.com/codespaces/badge.svg)](https://codespaces.new/clowa/clowa.dev?devcontainer_path=.devcontainer/devcontainer.json)

To run the project locally, make sure you have [Node.js](https://nodejs.org/) installed — or simply use the provided DevContainer.

For simplicity you can use the provided [Taskfile](https://taskfile.dev/) to run common tasks.

**`task init-devcontainer`**

- Sets up the project by setting permissions and installing dependencies
- Should be run once after starting a new devcontainer or codespaces instance

**`task dev`**

- Starts Astro's development server with live reloading
- Available at `http://localhost:4321`

**`task build`**

- Builds the static site to `swa/dist/`

## Architecture

The project ships as a single container image made up of **two technical components** that together deliver **three logical components**.

**Logical components** — what the site is made of:

| # | Component | What it is | Lives in |
| - | --------- | ---------- | -------- |
| 1 | Static website | The [Astro](https://astro.build/) site (pages, assets), pre-built to static files | [`swa/`](swa/) |
| 2 | REST API | A [Gin](https://gin-gonic.com/) + Go service serving dynamic content (`GET /api/quote`) | [`api/`](api/) |
| 3 | Redirects | Domain → domain 301 rules (one file per source host) | [`docker/caddy/redirects/`](docker/caddy/redirects/) |

**Technical components** — the processes that run inside the container:

- **[Caddy](https://caddyserver.com/)** — the web server. Serves the static files, reverse-proxies the API, and performs the redirects. It runs as the container's primary process.
- **Go REST API** — a single static binary that serves the dynamic content. It is served through caddy.

## Lessons Learned

See [LESSONS_LEARNED.md](LESSONS_LEARNED.md).

## Sources

- [Serverless v3 OOS alternative](https://github.com/oss-serverless/serverless)
- [Scaleway - Serverless Framework Plugin](https://github.com/scaleway/serverless-scaleway-functions)
- [Scaleway - Serverless Container Custom Domain](https://www.scaleway.com/en/docs/serverless-containers/how-to/add-a-custom-domain-to-a-container/)
- [Scaleway - Serverless Container Limitations](https://www.scaleway.com/en/docs/serverless-containers/reference-content/containers-limitations/)
- [S6-overlay - container native process manager](https://github.com/just-containers/s6-overlay#verifying-downloads)
