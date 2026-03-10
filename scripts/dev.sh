#!/usr/bin/env bash
set -euo pipefail

ROOT_DIR=$(cd "$(dirname "$0")/.." && pwd)

(
  cd "$ROOT_DIR/backend"
  go run ./cmd/server
) &
BACKEND_PID=$!

(
  cd "$ROOT_DIR/frontend"
  npm run dev
) &
FRONTEND_PID=$!

cleanup() {
  kill "$BACKEND_PID" "$FRONTEND_PID" 2>/dev/null || true
}

trap cleanup EXIT INT TERM
wait
