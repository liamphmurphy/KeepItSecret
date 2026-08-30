# keepitsecret-api Helm chart

This chart deploys the KeepItSecret backend API.

Set `image.repository` and, for immutable releases, `image.tag` to the image
built from `api/Dockerfile`:

```sh
helm upgrade --install keepitsecret-api ./charts/keepitsecret-api \
  --set image.repository=ghcr.io/example/keepitsecret-api \
  --set image.tag=sha-<commit>
```

The chart configures liveness and readiness probes against the API's existing
health endpoints. Ingress and CPU-based autoscaling are disabled by default.
