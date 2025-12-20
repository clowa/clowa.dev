# Overview

This directory contains the APIs of the static web app. Each API is implemented as a managed function.
For more information, see the [Static Web Apps API documentation](https://learn.microsoft.com/en-us/azure/static-web-apps/apis-overview).

## Getting Started

To run the APIs locally, you need to have the [Azure Static Web Apps CLI](https://learn.microsoft.com/en-us/azure/static-web-apps/local-development) installed.

Setup local development configuration by adding a `api/local.settings.json` file in this directory with the following content (example):

```json
{
  "IsEncrypted": false,
  "Values": {
    "FUNCTIONS_WORKER_RUNTIME": "node",
    "AzureWebJobsStorage": "UseDevelopmentStorage=true"
  }
}
```

After setting things up, you can start the local development server by running the following command in the root directory of the project:

```bash
task run-devserver
```

There are two main ways to run the project locally:

### `task run-devserver`

- Provides **live reloading** when you save files
- Uses Astro's development server for fast iteration
- **Limitation**: API endpoints don't work (known issue with SWA CLI)
- **Best for**: Frontend-only development

### `task run`

- Builds a static version of the entire app
- Spins up both the web app AND function APIs together
- **API endpoints work correctly**
- **Limitation**: No live reloading - requires rebuilding to see changes
- **Best for**: Testing the complete application including API functionality

> [!TIP]
> Choose `task run-devserver` for rapid frontend development, or `task run` when you need to test API endpoints and the full integrated experience.
