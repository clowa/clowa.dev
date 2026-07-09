# Lessons Learned

## BuildKit not available via `serverless-scaleway-functions`

The plugin builds Docker images using `dockerode` (Docker API directly), which does not read `DOCKER_BUILDKIT=1` from the environment and does not enable BuildKit by default. Dockerfile syntax requiring BuildKit (`RUN --mount=type=bind`, `RUN --mount=type=cache`) will fail with:

```
the --mount option requires BuildKit
```

**Fix:** Use standard `COPY` instructions instead. Copy package files before source files to preserve equivalent layer caching (`npm ci` is only re-run when `package.json`/`package-lock.json` change).

## ARM64 host produces wrong architecture for Scaleway

The `serverless-scaleway-functions` plugin calls `docker.buildImage()` via `dockerode` without specifying a `platform`, so the image is built for the host's native architecture (e.g., `arm64` on Apple Silicon). Scaleway Serverless Containers require `linux/amd64`.

**Fix:** Don't let the serverless plugin build the image at all. Build and push via `docker buildx build --platform linux/amd64` separately (locally via `task docker-build`, in CI via the `build` job). Pass the fully-qualified image reference to the serverless plugin via the `registryImage` env var (`REGISTRY_IMAGE`) so it only handles deployment, not the image build.
