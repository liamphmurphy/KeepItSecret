# KeepItSecret

KeepItSecret is a one-time secret-sharing application. The Go API is under
`api/`; the Next.js UI will live alongside it in this repository.

## Nix development environment

The root `flake.nix` is the shared development and build definition. Its
`flake.lock` pins the nixpkgs revision, including the Go toolchain and Go
development tools.

Install Nix with flakes enabled, then enter the project shell:

```sh
nix develop
```

The shell provides the pinned Go toolchain, `gopls`, `gotools`, and
`golangci-lint`. It also sets project-local Go caches and `GOTOOLCHAIN=local`
so Go does not silently download a different toolchain.

Run the API checks from the repository root:

```sh
nix develop --command go -C api test ./...
nix develop --command go -C api test -tags=integration ./...
nix develop --command go -C api test -race ./...
nix develop --command go -C api vet ./...
```

Build the API with the same Nix-pinned Go toolchain:

```sh
nix build .#api
./result/bin/keepitsecret-api
```

Validate the flake and all declared outputs with:

```sh
nix flake check
```

Update the pinned nixpkgs revision deliberately, then review the resulting
lockfile and rerun the checks:

```sh
nix flake lock --update-input nixpkgs
nix flake check
```

The current `api/Dockerfile` remains a conventional multi-stage Docker build.
The Nix build is the canonical reproducible Go build; a later container-build
step can package its output directly so Docker and CI do not maintain a second
Go-version declaration.
