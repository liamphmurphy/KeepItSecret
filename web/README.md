# KeepItSecret web

The Next.js application and backend-for-frontend live here. The browser encrypts secrets with AES-256-GCM; Route Handlers forward encrypted payloads to the Go API.

```sh
cp .env.example .env.local
npm install
npm run dev
```

Set `API_BASE_URL` to the Go API address. In the repository's local Kubernetes setup, port-forward the API to `http://127.0.0.1:8080`.

Useful checks:

```sh
npm run typecheck
npm run lint
npm test
npm run build
```
