FROM docker.io/nixos/nix:2.28.3 AS build

WORKDIR /src
COPY flake.nix flake.lock ./
COPY api ./api
RUN nix --extra-experimental-features 'nix-command flakes' \
    build .#api --out-link /tmp/keepitsecret-api \
    && mkdir -p /out \
    && cp -L /tmp/keepitsecret-api/bin/keepitsecret-api /out/keepitsecret-api

FROM gcr.io/distroless/static-debian12:nonroot
COPY --from=build /out/keepitsecret-api /keepitsecret-api
EXPOSE 8080
USER nonroot:nonroot
ENTRYPOINT ["/keepitsecret-api"]
