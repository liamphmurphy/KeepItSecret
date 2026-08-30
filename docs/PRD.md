# KeepItSecret Product Requirements Document

## 1. Document status

- Status: Draft
- Version: 0.1
- Date: 2026-08-30
- Product: KeepItSecret

## 2. Product summary

KeepItSecret is a web application for sharing sensitive text through links that can be opened only once. The sender enters a secret, receives a shareable URL, and gives that URL to one or more intended recipients. The first successful recipient retrieves and decrypts the secret; subsequent retrieval attempts fail.

The application will consist of:

- A Next.js web application and backend-for-frontend (BFF).
- A Go HTTP API responsible for secret lifecycle, retrieval, and policy enforcement.
- Persistent storage defined as infrastructure in a Helm chart.
- Kubernetes deployment targeted initially at DigitalOcean Kubernetes.

## 3. Goals

### 3.1 Primary goals

1. Allow a user to create a one-time secret-sharing link.
2. Ensure the secret plaintext is encrypted in the sender's browser before it leaves the browser.
3. Ensure the Go API and its database store ciphertext, not plaintext or decryption keys.
4. Ensure retrieval is atomic: exactly one request can successfully consume a secret.
5. Provide sensible expiration, size, abuse-prevention, and operational controls.
6. Make the application straightforward to deploy and operate on Kubernetes.

### 3.2 Non-goals for the initial release

- User accounts or contact management.
- Guaranteed protection against a compromised recipient device.
- Guaranteed protection against a compromised frontend deployment.
- File attachments.
- End-to-end encrypted multi-recipient groups.
- Password-protected secrets, unless added in a later iteration.
- Custom cryptography or a proprietary encryption format.

## 4. Security and privacy model

### 4.1 Security statement

The product provides client-side authenticated encryption with a server-blind storage layer. The backend receives and stores encrypted content but must not receive the plaintext or the decryption key.

This protects the secret against database exposure and unauthorized inspection of stored records. It does not protect against malicious JavaScript served by a compromised frontend, browser extensions, malware, screenshots, or a recipient who intentionally copies the decrypted secret.

### 4.2 Share URL format

The share URL will use a random, unguessable secret identifier and a URL fragment containing the decryption key:

```text
https://example.com/s/<secret-id>#<decryption-key>
```

The URL fragment is not sent in HTTP requests. The browser sends only the secret ID to the server, while frontend code reads the key locally and decrypts the returned ciphertext.

The complete URL is a bearer credential: anyone who obtains it can attempt to retrieve and decrypt the secret. The product UI must communicate this clearly.

### 4.3 Cryptography requirements

- Use AES-256-GCM through the browser Web Crypto API.
- Generate a new 256-bit encryption key for every secret using a cryptographically secure random generator.
- Generate a fresh random 96-bit nonce for every encryption operation.
- Store the nonce with the ciphertext; it is not secret.
- Preserve and verify the authenticated encryption tag supplied by AES-GCM.
- Use authenticated additional data (AAD) to bind the ciphertext to the secret ID and encryption-format version.
- Encode binary values using base64url without unnecessary padding.
- Never implement cryptographic primitives or random-number generation manually.
- Reject modified, malformed, expired, or otherwise unauthentic ciphertext.

The initial wire format should be versioned, for example:

```json
{
  "version": 1,
  "secretId": "...",
  "ciphertext": "...",
  "nonce": "...",
  "expiresAt": "..."
}
```

The exact API representation may differ, but the format must be documented and versioned so it can evolve safely.

### 4.4 Transport and browser protections

- Serve all production traffic over HTTPS.
- Prefer TLS 1.3 and support TLS 1.2 only where necessary.
- Enable HSTS after validating the production domain and certificate setup.
- Send `Cache-Control: no-store` for secret creation, retrieval, and display responses.
- Do not send plaintext, keys, or complete share URLs to application logs, analytics, error trackers, or tracing systems.
- Use a restrictive Content Security Policy and avoid third-party scripts on secret pages.
- Set an appropriate `Referrer-Policy` such as `no-referrer`.
- Clear the URL fragment from the address bar after extracting it when doing so does not harm usability.
- Do not persist plaintext in localStorage, sessionStorage, IndexedDB, cookies, or service-worker caches.

## 5. User experience

### 5.1 Create a secret

The sender can:

1. Open the create-secret page.
2. Enter or paste text into a protected input area.
3. Optionally choose an expiration period from supported presets.
4. Submit the secret.
5. See a generated share URL and a clear warning that anyone with the URL can use it.
6. Copy the URL to the clipboard.

The UI should not display the plaintext after successful submission unless the sender explicitly chooses to keep it visible.

### 5.2 Retrieve a secret

The recipient can:

1. Open the share URL.
2. See a loading state while the browser claims the ciphertext.
3. See the decrypted secret if retrieval succeeds.
4. See a clear consumed, expired, invalid, or unavailable state otherwise.
5. Copy the plaintext deliberately, with a warning that copying may expose it to the clipboard.

Retrieval must not be triggered by page crawlers or link previews. Loading the informational page and consuming the secret should be separate operations, with consumption initiated by the application only after the recipient page has loaded.

### 5.3 Error behavior

Responses should avoid revealing unnecessary information. For example, an invalid ID and an already-consumed ID may share the same user-facing message. Internal logs may retain a safe reason code without the secret, key, or full URL.

## 6. Functional requirements

### 6.1 Secret creation

- The browser generates the encryption key and encrypts the plaintext locally.
- The browser submits ciphertext, nonce, version, expiration, and a random secret ID to the API through the Next.js BFF.
- The API validates payload size, format, expiration, and identifier entropy.
- The API stores the encrypted record and returns success.
- The browser constructs the share URL using the returned or pre-generated ID and the local key.

### 6.2 Secret retrieval

- The browser extracts the key from the URL fragment locally.
- The browser requests a one-time claim for the secret ID.
- The API atomically reads and consumes the record.
- A successful response contains ciphertext and the metadata needed for local verification and decryption.
- The browser verifies and decrypts the ciphertext locally.
- Any authentication failure must result in no plaintext display.

### 6.3 Expiration and limits

- Every secret has an expiration timestamp.
- The API rejects creation with an expiration outside configured policy limits.
- Expired records cannot be claimed.
- The initial release must define maximum plaintext/ciphertext size, minimum and maximum expiration, and per-client creation/retrieval rate limits.
- Cleanup of expired records must not be required for correctness, but a background cleanup mechanism should reclaim storage.

### 6.4 One-time semantics

The consume operation must be implemented as an atomic database transaction or equivalent compare-and-delete operation. Concurrent requests must produce one success and failures for all later attempts.

The system should record a consumed timestamp and safe operational metadata, subject to the privacy policy.

## 7. Proposed architecture

```text
Browser
  │ HTTPS
  ▼
Next.js application / BFF
  │ private Kubernetes service
  ▼
Go API
  │ private Kubernetes service
  ▼
Database or key-value store
```

### 7.1 Next.js BFF

- Provides the public web experience.
- Proxies only encrypted payloads and secret identifiers to the Go API.
- Keeps the Go API service address private from ordinary browser clients.
- Must not be treated as a cryptographic trust boundary.
- Must not log request bodies, URL fragments, or sensitive headers.

The BFF can reduce API exposure, but it cannot make browser network activity invisible. The browser necessarily receives JavaScript and makes requests.

### 7.2 Go API

Responsibilities include:

- Payload validation.
- Secret creation.
- Atomic one-time claim.
- Expiration enforcement.
- Rate limiting and abuse controls.
- Safe structured logging and health endpoints.

The API should not contain a plaintext decryption path in the initial design.

### 7.3 Storage

The initial deployment may use PostgreSQL or a compatible transactional database. The schema should support atomic claim semantics, expiration queries, and indexed lookup by secret ID.

Storage must use encrypted volumes and protected backups. Database credentials must be supplied through Kubernetes secret management, not committed configuration.

## 8. Kubernetes and DigitalOcean requirements

- Package application components in a Helm chart.
- Define the Next.js app and Go API as separate Deployments and Services.
- Keep the database private; use a managed DigitalOcean database or a separately evaluated stateful deployment.
- Provide configurable resource requests/limits, replica counts, image references, and environment settings.
- Provide readiness and liveness probes for application pods.
- Support TLS termination through the cluster ingress/controller.
- Add NetworkPolicies where supported.
- Use non-root containers, read-only filesystems where practical, and minimal production images.
- Store no persistent secrets in images, Helm values committed to source, or logs.
- Document database migration and backup/restore procedures.

## 9. API outline

The following is an initial contract, not a final implementation requirement:

```text
POST /api/v1/secrets
POST /api/v1/secrets/{secretId}/claim
GET  /api/v1/health/live
GET  /api/v1/health/ready
```

`POST /secrets` accepts encrypted content only. `POST /secrets/{id}/claim` is preferred over a retrieval `GET` because consuming a resource is a state-changing operation and should not be triggered by crawlers or prefetchers.

The claim endpoint must be idempotent only in the sense that repeated claims after the first successful claim fail safely; it must not return the ciphertext twice.

## 10. Threat model and mitigations

| Threat | Mitigation | Residual risk |
| --- | --- | --- |
| Database compromise | Client-side AES-GCM encryption | Attacker sees metadata and can perform offline analysis of any future password-based scheme |
| Guessing a secret ID | 128+ bits of cryptographically random ID; rate limits | Denial of service remains possible |
| Network interception | HTTPS, TLS configuration, HSTS | Compromised endpoint or trusted client remains in scope |
| Ciphertext tampering | AES-GCM authentication and AAD | Key holder can still intentionally alter local data |
| Double retrieval race | Atomic consume transaction | Operational outages may cause availability issues |
| URL leakage | Fragment key, no-referrer policy, no logs, no third-party scripts | User sharing, history, screenshots, extensions, and messaging systems may retain it |
| XSS or malicious frontend | CSP, dependency review, secure deployment pipeline | A server that serves malicious JavaScript can capture plaintext or keys |
| Database backup exposure | Encrypted backups and access controls | Authorized infrastructure operators may still access ciphertext and metadata |

## 11. Observability and privacy

Metrics may include request counts, latency, error classes, storage usage, and successful/failed claim counts. Metrics and logs must not contain plaintext, encryption keys, ciphertext bodies, full share URLs, or URL fragments.

The PRD must establish a retention policy for IP addresses, user agents, audit events, and consumed-record metadata before production launch.

## 12. Acceptance criteria for the first release

- A sender can create a secret and receive a complete share URL.
- The Go API database contains no sender plaintext or decryption key.
- A recipient with the complete URL can decrypt the secret in the browser.
- A second recipient cannot successfully claim the same secret.
- Concurrent claim requests result in exactly one successful claim.
- Expired secrets cannot be claimed.
- Tampered ciphertext fails authentication and displays no plaintext.
- Secret content and URL fragments are absent from application logs and telemetry.
- Link previews do not consume secrets.
- Production traffic is HTTPS-only with security headers enabled.
- The complete system can be deployed from the Helm chart to a supported Kubernetes cluster.
- Automated tests cover encryption/decryption, malformed payloads, expiration, atomic claim races, rate limits, and log redaction.

## 13. Future considerations

- Optional password-derived encryption keys using Argon2id.
- Secret deletion before retrieval through a separate capability.
- Encrypted file attachments and streaming encryption.
- Abuse reporting and administrative controls that do not expose plaintext.
- Managed key or secret services for infrastructure credentials.
- Formal security review and penetration testing before handling high-value secrets.

## 14. Reference guidance

- [OWASP Cryptographic Storage Cheat Sheet](https://cheatsheetseries.owasp.org/cheatsheets/Cryptographic_Storage_Cheat_Sheet.html)
- [OWASP Key Management Cheat Sheet](https://cheatsheetseries.owasp.org/cheatsheets/Key_Management_Cheat_Sheet.html)
- [OWASP Transport Layer Security Cheat Sheet](https://cheatsheetseries.owasp.org/cheatsheets/Transport_Layer_Security_Cheat_Sheet.html)
- [OWASP HTTP Strict Transport Security Cheat Sheet](https://cheatsheetseries.owasp.org/cheatsheets/HTTP_Strict_Transport_Security_Cheat_Sheet.html)
- [MDN: URI fragments](https://developer.mozilla.org/en-US/docs/Web/URI/Reference/Fragment)
