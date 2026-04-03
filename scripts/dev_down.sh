#!/usr/bin/env bash
set -euo pipefail

ROOT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
export REPO_ROOT="$ROOT_DIR"

docker compose \
  -f "$ROOT_DIR/platform/observability/compose/docker-compose.observability.yml" \
  -f "$ROOT_DIR/deployments/compose/docker-compose.app.dev.yml" \
  down --remove-orphans
