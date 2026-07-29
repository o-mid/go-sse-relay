#!/usr/bin/env bash
set -euo pipefail

ROOT="$(cd "$(dirname "$0")/.." && pwd)"
cd "$ROOT"

ADDR="${LISTEN_ADDR:-127.0.0.1:8080}"
BASE="http://${ADDR}"
CHANNEL="${CHANNEL:-demo_user}"

echo "==> health"
curl -fsS "$BASE/healthz"
echo
echo

echo "==> publish event to channel=${CHANNEL}"
curl -fsS -X POST "$BASE/v1/hooks/${CHANNEL}" \
  -H 'Content-Type: application/json' \
  -d '{"type":"payment.updated","data":{"status":"settled","amount":"12.50"}}'
echo
echo
echo "Browser demo: ${BASE}/  (connect channel=${CHANNEL})"
echo "Stream here:  curl -N \"${BASE}/v1/stream?channel=${CHANNEL}\""
