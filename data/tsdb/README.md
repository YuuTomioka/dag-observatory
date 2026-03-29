# TSDB Data Assets

This directory defines the repository's time-series data assets.

## Role

- schema and migration assets
- query assets
- data design notes for analysis-oriented storage

## Data Source Of Truth

SQL is the single source of truth for data design in this repository.

That includes:

- schema definition
- migrations
- query specification
- TimescaleDB-specific operations

Application code is a user of SQL assets, not the owner of schema truth.

## Design Rules

- DDL changes belong in migration SQL
- query logic should remain in shared SQL assets
- generated code may wrap SQL, but must not redefine the schema
- ORM models must not become the primary schema definition

## Storage Role Split

The current repository separates storage responsibilities as follows:

- TimescaleDB: primary durable store for relational and time-series data
- MinIO: durable blob storage for large artifacts and backup bodies
- StateStore: runtime engine state, separate from application data persistence
- Kafka/Redpanda: event transport and event history layer

## Role Across Implementations

- Go may generate code from shared SQL assets
- Python may use shared SQL assets without owning schema truth
- both implementations should converge on the same data model

## Data Asset Categories

- `schema/`: DDL and migration assets
- `query/`: shared query specification
- `seed/`: optional initialization data
- notes: rationale for time-series-oriented design decisions

Schema-specific structure and migration filename rules are defined in `data/tsdb/schema/README.md`.

## Minimal Data Model Direction

The current repository direction centers on:

- `task_attempts` as the primary execution record
- `processed_events` for idempotency protection
- `outbox_events` for transactional publication
- `artifacts` for artifact metadata
- hypertables such as `measurements` for time-series data

Supporting concepts such as upload sessions and GC are operational extensions around these core records.

## Artifact And Backup Direction

MinIO-backed object storage is treated as the body store for:

- worker result artifacts
- database backup artifacts

The database should store metadata and references, while file transfer should be handled through presigned access patterns rather than application-body proxying.

## Current Layout

```txt
data/tsdb/
├─ README.md
├─ schema/
│  ├─ README.md
│  ├─ migrate/
│  └─ seed/
└─ query/
```

## Migrator Usage

The migrator is treated as an execution helper for data assets.

- SQL itself belongs to `data/tsdb/`
- the runnable container and compose entrypoint may remain outside `data/` while repository restructuring is in progress
- its role is to apply data-owned migration assets, not to define schema truth

### Docker Compose Example

```sh
export TSDB_URL="postgres://user:pass@host:5432/dbname?sslmode=disable"
docker compose -f deployments/compose/docker-compose.migrator.yml run --rm tsdb-migrator
```

### Compose Startup Order

```sh
# 1) Start infrastructure
docker compose -f deployments/compose/docker-compose.app.dev.yml up -d timescaledb minio redpanda

# 1.5) Start observability stack if needed
docker compose -f platform/observability/compose/docker-compose.observability.yml up -d

# 2) Apply migrations
export TSDB_URL="postgres://dag:dag@localhost:5432/dag?sslmode=disable"
docker compose -f deployments/compose/docker-compose.migrator.yml run --rm tsdb-migrator

# 3) Start applications
docker compose -f deployments/compose/docker-compose.app.dev.yml up -d dag-core-api worker-py worker-py-outbox worker-py-gc
```

### CI Example

```yaml
name: tsdb-migrate
on:
  workflow_dispatch:
jobs:
  migrate:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v4
      - name: Run migrator
        run: |
          docker build -f deployments/migrator/Dockerfile -t tsdb-migrator .
          docker run --rm -e TSDB_URL="${{ secrets.TSDB_URL }}" tsdb-migrator
```

## Boundary Note

`deployments/migrator/Dockerfile` and `deployments/compose/docker-compose.migrator.yml` are currently kept in `deployments/` as operational wrappers.

They should be read as execution shells around `data/tsdb/` assets rather than as the owner of database design.

## Migration Notes

This document is the durable home for TSDB data asset guidance and the old migration-era store design notes.
