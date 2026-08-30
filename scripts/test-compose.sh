#!/usr/bin/env bash
set -euo pipefail

compose=(docker compose -f compose.integration.yaml)

cleanup() {
  "${compose[@]}" down --remove-orphans
}
trap cleanup EXIT

"${compose[@]}" up --build -d

for attempt in {1..30}; do
  if curl --fail --silent http://127.0.0.1:18080/api/v1/health/ready >/dev/null \
    && curl --fail --silent http://127.0.0.1:13000 >/dev/null; then
    break
  fi
  if [[ "$attempt" == 30 ]]; then
    "${compose[@]}" logs
    exit 1
  fi
  sleep 2
done

node scripts/compose-integration.mjs
