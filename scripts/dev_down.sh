#!/usr/bin/env bash
set -euo pipefail

cd deployments/compose
docker compose -f docker-compose.observability.yml -f docker-compose.app.dev.yml down --remove-orphans
