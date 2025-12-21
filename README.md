[![Azure Static Web Apps CI/CD](https://github.com/clowa/me.clowa.de/actions/workflows/deployment.yaml/badge.svg)](https://github.com/clowa/me.clowa.de/actions/workflows/deployment.yaml)

# Overview

This is my personal website. It is built using [Astro](https://astro.build/) and hosted on [Azure Static Web Apps](https://azure.microsoft.com/de-de/products/app-service/static). Go and check it out at [https://clowa.dev](https://clowa.dev).

## Repository Structure

- `api/` hosts the Azure Functions for the APIs
- `swa/` hosts the Frontend Code of the Website

## Getting Started

[![Open in GitHub Codespaces](https://github.com/codespaces/badge.svg)](https://codespaces.new/clowa/clowa.dev?devcontainer_path=.devcontainer/devcontainer.json)

To run the project locally, make sure you have [Node.js](https://nodejs.org/) and the [Azure Static Web Apps CLI](https://learn.microsoft.com/en-us/azure/static-web-apps/local-development) installed - or simply use the provided DevContainer.

For simpliciy you can use the provided [Taskfile](https://taskfile.dev/) to run common tasks.

**`task init-devcontainer`**

- Sets up the project by setting up permissions and installing dependencies
- Should be run once after starting a new devcontainer or codespaces instance

>[!NOTE]
> There are two main ways to run the project locally.

**`task run-devserver`**

- Provides **live reloading** of the swa when you save files
- Uses Astro's development server for fast iteration
- **Limitation**: API endpoints don't support live-reloading
- **Best for**: Frontend-only development

**`task run`**

- Builds a static version of the entire app
- Spins up both the web app AND function APIs together
- **API endpoints work correctly**
- **Limitation**: No live reloading - requires rebuilding to see changes
- **Best for**: Testing the complete application including API functionality

> [!TIP]
> Choose `task run-devserver` for rapid frontend development, or `task run` when you need to test API endpoints and the full integrated experience.
