# Data Conventions

## Purpose

This document defines data design conventions for time-series and analysis assets.

## Principles

- time-series data is a first-class design asset
- schema, migration, and query responsibilities must remain explicit
- data assets should outlive a single implementation detail
- application code should not hide the authoritative data design

## SQL Single Source Of Truth

SQL is the single source of truth for repository-wide database design.

That includes:

- DDL
- migrations
- query specifications
- TimescaleDB-specific operations

Application code may consume these assets, generate wrappers around them, or map results from them, but must not silently replace them as the schema authority.

## Repository Placement Rules

Data assets should be separated by responsibility:

- schema and migration SQL
- seed SQL when needed
- shared query SQL
- design notes for time-series usage

This repository is moving those assets under `data/tsdb/` as a permanent home.

## Implementation Rules

- Go may use generated access code from shared SQL
- Python may use shared SQL without owning the schema definition
- ORM usage is allowed only as a query or mapping helper, not as schema truth
- automatic schema migration from application runtime is not the primary model

## Migration Rules

- schema changes should be forward-only migration assets
- application services should not be the primary DDL executor in long-lived environments
- operational migration paths should remain reproducible and auditable
- migration filename ordering must be deterministic and repository-wide rules should avoid mixed-width sequence formats

## Timeseries Design Rules

- hypertable, compression, retention, and continuous aggregate operations belong in SQL assets
- time-series storage design should be reviewable independently from application code
- data review should function as architecture review, not only implementation review

## Ownership Rules Across Services

- worker-side services may be the primary writers for durable execution records and artifact metadata
- API-side services should prefer read, coordination, and controlled access responsibilities over competing writes
- artifact bodies and metadata should be separated deliberately
- outbox-based publication is preferred when database confirmation and event publication must stay aligned

## Event And State Storage Rules

- event transport format should preserve an explicit event shape
- partitioning rules should be consistent with state scope
- commit and retry rules must be explicit for event-driven persistence
- poison message handling must be a deliberate operational rule, not an undefined edge case

## Migration Notes

This document is the durable home for repository-level data and TSDB conventions.
