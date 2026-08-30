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
