FROM docker.io/library/golang:1.26 AS build

WORKDIR /src
COPY go.mod ./
COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -trimpath -ldflags='-s -w' -o /out/keepitsecret-api ./cmd/api

FROM gcr.io/distroless/static-debian12:nonroot
COPY --from=build /out/keepitsecret-api /keepitsecret-api
EXPOSE 8080
USER nonroot:nonroot
ENTRYPOINT ["/keepitsecret-api"]
