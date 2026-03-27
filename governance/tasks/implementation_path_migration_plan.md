# Implementation Path Migration Plan

## Background

An external decision has been made to adopt the structure policy that treats `implementations/` as the physical source of truth for reference implementations.

In this decision, moving `dag-core/` and `worker-py/` under `implementations/` is mandatory, not optional.

Current repository documents still include the old accepted decision (`ADR-0002`) that kept those directories at the top level.
That conflict must be resolved first at the policy/document level, then executed as a migration project.

## In Scope

- define migration execution policy before touching implementation paths
- supersede or replace the old ADR that blocked physical move
- migrate physical paths of `dag-core/` and `worker-py/` into `implementations/`
- update build/test/dev/CI/deployment references that currently depend on top-level paths
- align repository entrypoint documents with the new physical structure

## Out Of Scope

- redesign of runtime architecture unrelated to path move
- feature development inside Go/Python implementations
- observability or TSDB design changes not required by path migration
- product-level behavior changes

## Current State Snapshot

The repository is currently in a mixed state:

- top-level `dag-core/` and `worker-py/` exist and contain active implementation code
- `implementations/dag-core/` and `implementations/worker-py/` exist as guidance/asset locations
- `blueprint/architecture/system-boundaries.md` already describes implementations under `implementations/*`
- `blueprint/adr/ADR-0002-reference-implementation-paths.md` explicitly keeps top-level physical paths
- `deployments/compose/docker-compose.app.dev.yml` builds from `../../dag-core` and `../../worker-py`
- `Makefile` uses `GO_DIR := dag-core`
- `.github/workflows/contracts-check.yml` uses `dag-core/go.mod`
- multiple docs still explain "conceptual ownership only, no physical move"

## Execution Policy

Use a single migration stream with strict ordering:

1. policy alignment first (`blueprint/adr/` + entrypoint docs),
2. physical move second,
3. reference/path fixups third,
4. verification and cleanup last.

For this repository, a split long-lived compatibility period is not recommended.
Maintaining both old and new paths would increase ambiguity and violate the parent-repository intent.

## Task Breakdown

### T1. Decision Alignment (Docs/ADR)

- add a new ADR that supersedes `ADR-0002` and records the rationale for physical move
- update `README.md` to remove statements that paths stay top-level
- update `implementations/*/README.md` from "will become canonical" to "is canonical"
- update `deployments/README.md` ownership wording to reference `implementations/*`

### T2. Physical Move Preparation

- confirm target layout:
  - `dag-core/` -> `implementations/dag-core/` (merge with existing implementation docs/assets)
  - `worker-py/` -> `implementations/worker-py/` (merge with existing implementation docs)
- identify file collisions before move (`README.md` is expected and should be reconciled)
- decide whether OpenAPI artifact path remains `implementations/dag-core/openapi/` (recommended: keep)

### T3. Physical Move Execution

- move directories with history preservation
- reconcile conflicting files in destination directories
- remove obsolete top-level directories after move completion

### T4. Path Reference Fixups

- update `Makefile` (`GO_DIR`, generated artifact paths, contract checks)
- update compose build contexts and bind mounts in `deployments/compose/docker-compose.app.dev.yml`
- update CI references in `.github/workflows/contracts-check.yml`
- update scripts and docs that use old paths (`README.md`, `data/tsdb/README.md`, others found by search)

### T5. Verification

- run `make go-test`
- run `make contracts-check`
- run compose config validation for app dev stack
- verify no stale top-level path references remain in tracked files

## Acceptance Criteria

- no active implementation code remains in top-level `dag-core/` or `worker-py/`
- `implementations/dag-core/` and `implementations/worker-py/` are the only physical implementation roots
- old "no physical move" statements are removed or superseded in ADR/docs
- `make go-test` and `make contracts-check` pass
- app-development compose still resolves build contexts and volumes correctly
- CI workflow references point to moved paths
- repository-wide search confirms no unintended stale references to old top-level implementation paths

## Result

In progress.

- T1 completed.
- `ADR-0003` was added and `ADR-0002` status was changed to superseded.
- top-level implementation code was moved to `implementations/dag-core/` and `implementations/worker-py/`.
- path references in `Makefile`, compose, CI, and entrypoint docs were updated.
- `make go-test` passed after narrowing test target to `cmd/...` and `internal/...`.
- `make contracts-check` currently fails in this environment because `buf` is not installed.
- top-level `dag-core/tmp` cleanup is blocked by file ownership/permission (`nobody` owned files).
