# Overview

This directory contains the APIs of the static web app. Each API is implemented as a managed function.
For more information, see the [Static Web Apps API documentation](https://learn.microsoft.com/en-us/azure/static-web-apps/apis-overview).

## Getting Started

To run the APIs locally, you need to have the [Azure Static Web Apps CLI](https://learn.microsoft.com/en-us/azure/static-web-apps/local-development) or at least the [Azure Functions Core Tools](https://learn.microsoft.com/en-us/azure/azure-functions/functions-run-local) installed.

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

You can then start the function host with the following command:

```sh
cd api
func start
```

This will start the function host and expose the API endpoints at `http://localhost:7071/api/{functionName}`.
For a better local development experience, consider using the `swa start` command from the [Azure Static Web Apps CLI](https://learn.microsoft.com/en-us/azure/static-web-apps/local-development) to run both the frontend and backend together.
