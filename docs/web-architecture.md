# KeepItSecret Web Architecture

```text
Browser client components
  ├─ Web Crypto API (AES-256-GCM, key stays in URL fragment)
  └─ fetch('/api/v1/...')
          │
          ▼
Next.js Route Handlers (BFF)
  └─ fetch(API_BASE_URL + '/api/v1/...')
          │
          ▼
Go API
```

The BFF is a routing and policy boundary, not a cryptographic boundary. It forwards encrypted payloads and secret identifiers only. `API_BASE_URL` is read in server-side route handlers and is never exposed through `NEXT_PUBLIC_*` configuration.

The URL fragment is read only by the recipient browser and is never included in the claim request. The create flow generates the secret identifier and AES key locally, uses the identifier/version as AES-GCM additional authenticated data, and puts only the key in the fragment.

The first implementation uses plain CSS and small client components rather than a UI framework. This keeps the visual system explicit and avoids adding a component abstraction before the product language is established.
