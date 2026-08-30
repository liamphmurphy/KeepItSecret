# KeepItSecret Web Requirements

## First web slice

- A sender can enter text, choose a supported expiration, encrypt it in the browser, and create a one-time link.
- A sender can copy the complete link without the plaintext or key being sent to the Go API.
- A recipient can open `/s/[secretId]`; the client extracts the fragment key, claims once through the Next.js BFF, authenticates/decrypts locally, and can deliberately copy the plaintext.
- Missing keys, malformed data, unavailable secrets, expired secrets, and authentication failures have clear safe states.
- Dark mode is the default. A user can toggle light mode, and the preference is retained without storing secret content.
- The browser talks to Next.js Route Handlers. The Go API address is server-only via `API_BASE_URL`.
- Secret operations use `Cache-Control: no-store`; sensitive values are not logged or persisted.

## Acceptance checks

- `npm run lint` and `npm run build` pass in `web/`.
- Crypto unit tests cover round trips, tampering, wrong keys, and base64url behavior.
- BFF routes reject malformed input and preserve the upstream status without exposing the Go service address to browser code.
- The UI is usable at narrow widths and has keyboard-visible focus states.
