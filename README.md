# Overview

This is my personal website. It is built using [Astro](https://astro.build/) and hosted as a container. Go and check it out at [https://clowa.dev](https://clowa.dev).

## Repository Structure

- `swa/` — Frontend code of the website (Astro)

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

## Lessons Learned

See [LESSONS_LEARNED.md](LESSONS_LEARNED.md).

## Sources

- [Serverless v3 OOS alternative](https://github.com/oss-serverless/serverless)
- [Scaleway - Serverless Framework Plugin](https://github.com/scaleway/serverless-scaleway-functions)
- [Scaleway - Serverless Container Custom Domain](https://www.scaleway.com/en/docs/serverless-containers/how-to/add-a-custom-domain-to-a-container/)
- [Scaleway - Serverless Container Limitations](https://www.scaleway.com/en/docs/serverless-containers/reference-content/containers-limitations/)
