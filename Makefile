NIX ?= nix

.DEFAULT_GOAL := help

.PHONY: help develop test test-integration test-race vet fmt lint build docker-build check update-lock

help:
	@printf '%s\n' \
		'make develop          Enter the Nix development shell' \
		'make test             Run unit tests' \
		'make test-integration Run integration tests' \
		'make test-race        Run tests with the race detector' \
		'make vet              Run go vet' \
		'make fmt              Format the API' \
		'make lint             Run golangci-lint' \
		'make build            Build the API with Nix' \
		'make docker-build     Build the API container with the Nix package' \
		'make check            Validate the flake' \
		'make update-lock      Update the pinned nixpkgs revision'

develop:
	$(NIX) develop

test:
	$(NIX) develop --command go -C api test ./...

test-integration:
	$(NIX) develop --command go -C api test -tags=integration ./...

test-race:
	$(NIX) develop --command go -C api test -race ./...

vet:
	$(NIX) develop --command go -C api vet ./...

fmt:
	$(NIX) develop --command gofmt -w $$(find api -type f -name '*.go')

lint:
	$(NIX) develop --command golangci-lint run ./api/...

build:
	$(NIX) build .#api

docker-build:
	docker build -f api/Dockerfile -t keepitsecret-api .

check:
	$(NIX) flake check

update-lock:
	$(NIX) flake lock --update-input nixpkgs
