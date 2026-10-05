---
name: run-container-locally
description: Build and run the full clowa.dev container locally, including the OpenTelemetry collector, with the Middleware API key pulled from 1Password. Use when running, debugging, or smoke-testing the container image on a dev machine.
---

# Run the container locally

With `OTEL_EXPORTER_OTLP_ENDPOINT` unset, the api's telemetry is a no-op and Caddy's export fails quietly in the background; the collector only runs in the image. To run the full container (including the collector) locally:

```sh
docker buildx build -t clowa/clowa.dev .; docker run -it --rm \
  --name clowa-web \
  -p 8080:80  \
  -e MIDDLEWARE_OTLP_ENDPOINT=https://cedric-ahlers.middleware.io:443 \
  -e MIDDLEWARE_API_KEY="$(op read 'op://Informatik/Middleware/bjohkfrgxqr5axh7kbjbikizpm/clowa-dev')" \
  -e OTEL_EXPORTER_OTLP_ENDPOINT=http://localhost:4317 \
  -e OTEL_EXPORTER_OTLP_PROTOCOL=grpc \
  -e OTEL_SERVICE_NAME=web-proxy \
  -e OTEL_RESOURCE_ATTRIBUTES=service.namespace=clowa,deployment.environment.name=local \
  clowa/clowa.dev
```

The site is then served on `http://localhost:8080`.
