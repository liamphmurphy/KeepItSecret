# KeepItSecret

KeepItSecret is a one-time secret-sharing application. The Go API is under
`api/`; the Next.js UI will live alongside it in this repository.

## Nix development environment

The root `flake.nix` is the shared development and build definition. Its
`flake.lock` pins the nixpkgs revision, including the Go toolchain and Go
development tools.

Install Nix with flakes enabled, then enter the project shell:

```sh
make develop
```

The shell provides the pinned Go toolchain, `gopls`, `gotools`, and
`golangci-lint`. It also sets project-local Go caches and `GOTOOLCHAIN=local`
so Go does not silently download a different toolchain.

Run the API checks from the repository root:

```sh
make test
make test-integration
make test-race
make vet
```

Run the API and Next.js services together in Docker Compose and exercise their
HTTP lifecycle:

```sh
make test-compose
```

This builds the production API and Next.js images, waits for both services,
creates and claims a secret through the Next.js proxy, verifies decryption,
and confirms that a second claim is rejected.

Run the stack interactively for local development:

```sh
make run
```

The web app is available at <http://127.0.0.1:13000>. Stop the services with
`Ctrl-C`, then clean up the Compose resources with:

```sh
make stop
```

Build the API with the same Nix-pinned Go toolchain:

```sh
make build
./result/bin/keepitsecret-api
```

Validate the flake and all declared outputs with:

```sh
make check
```

Update the pinned nixpkgs revision deliberately, then review the resulting
lockfile and rerun the checks:

```sh
make update-lock
make check
```

Run `make help` to see all available development targets.

The `api/Dockerfile` packages the same Nix-built API artifact into the minimal
runtime image, so Docker and standalone Nix builds use the same Go toolchain and
build flags. Build it from the repository root with:

```sh
make docker-build
```

## Run locally in Kubernetes with kind

Install [kind](https://kind.sigs.k8s.io/), `kubectl`, Helm, Docker, and Nix.
Then run the following commands from the repository root:

```sh
kind create cluster --name keepitsecret
docker build -f api/Dockerfile -t keepitsecret-api:local .
kind load docker-image keepitsecret-api:local --name keepitsecret

helm upgrade --install keepitsecret-api ./api/charts/keepitsecret-api \
  --namespace keepitsecret \
  --create-namespace \
  --set image.repository=keepitsecret-api \
  --set image.tag=local \
  --set image.pullPolicy=IfNotPresent

kubectl --namespace keepitsecret rollout status deployment/keepitsecret-api
```

Expose the API on `localhost:8080` and verify its health endpoint:

```sh
kubectl --namespace keepitsecret port-forward service/keepitsecret-api 8080:8080
```

In another terminal:

```sh
curl http://127.0.0.1:8080/api/v1/health/ready
```

When finished, remove the local cluster:

```sh
kind delete cluster --name keepitsecret
```
