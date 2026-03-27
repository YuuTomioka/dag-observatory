# Deployments

This directory holds operational wrappers for local execution.

## Role

- app-development compose assets that wire implementations and infrastructure together
- cross-cutting execution wrappers that do not own the underlying design truth

## Current Contents

- `compose/docker-compose.app.dev.yml`: local app-development stack composition
- `compose/docker-compose.migrator.yml`: wrapper for running the TSDB migrator container
- `migrator/Dockerfile`: build wrapper for migration execution

## Ownership Boundaries

- observability topology and its main compose entrypoint belong to `platform/observability/`
- time-series schema, migrations, and queries belong to `data/tsdb/`
- application implementation details belong to `implementations/dag-core/` and `implementations/worker-py/`

`deployments/` should therefore be read as an execution-layer helper, not as the source of truth for architecture, data design, or implementation structure.

## Current As-Is Note

Observability backend configuration has been moved under `platform/observability/`.

`deployments/` now remains focused on local execution wrappers rather than on permanent platform configuration.
