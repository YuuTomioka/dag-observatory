#!/usr/bin/env bash
set -euo pipefail

docker compose -f platform/observability/compose/docker-compose.observability.yml up -d
cd deployments/compose
docker compose -f ../../platform/observability/compose/docker-compose.observability.yml -f docker-compose.app.dev.yml up -d --build
