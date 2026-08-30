# KeepItSecret API Architecture

## Scope

This document describes the initial Go API vertical slice. Kubernetes manifests and Helm packaging are intentionally out of scope for now.

## Dependency direction

The API follows a small clean-architecture boundary:

```text
HTTP delivery → use case → domain interfaces ← repository adapter
```

- `internal/domain` contains entities, errors, and repository interfaces.
- `internal/usecase` owns validation and secret lifecycle rules.
- `internal/repository` contains storage implementations.
- `internal/delivery/http` maps HTTP requests and responses to use-case values.
- `cmd/api` wires the application and starts the server.

The first repository is in-memory so the API can be exercised without infrastructure. It is explicitly a development/test adapter; a transactional persistent adapter must be added before production deployment.

## One-time claim invariant

For a secret ID, at most one claim may return ciphertext. The repository owns the atomic read-and-delete operation. A future SQL implementation must enforce the same invariant in one transaction, for example with a conditional delete or `SELECT ... FOR UPDATE` followed by delete.

An expired record is not claimable and is removed by the in-memory adapter during the claim operation.

## HTTP contract

- `POST /api/v1/secrets` stores encrypted content and returns the secret ID.
- `POST /api/v1/secrets/{secretID}/claim` atomically consumes the secret and returns encrypted content.
- `GET /api/v1/health/live` reports process liveness.
- `GET /api/v1/health/ready` reports readiness.

The API accepts ciphertext only. It does not implement a plaintext decryption path.

## Resilience and privacy decisions

- Request bodies are bounded before JSON decoding.
- Request contexts are propagated through the use-case and repository boundaries.
- No automatic retries are performed for create or claim operations because claim is state-changing and a timeout can make its outcome ambiguous.
- Logs use `log/slog` and omit request bodies, ciphertext, keys, and complete URLs.
- HTTP responses use `Cache-Control: no-store` for secret operations.
