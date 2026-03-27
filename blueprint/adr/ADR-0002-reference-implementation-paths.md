# ADR-0002 Reference Implementation Paths

## Status

Superseded by ADR-0003

## Context

The repository has been reorganized around `blueprint/`, `implementations/`, `platform/`, `data/`, `scenarios/`, and `governance/`.

During that work, `dag-core/` and `worker-py/` were identified as conceptually belonging to the `implementations/` block.
One remaining question was whether they should also be physically moved under `implementations/`.

In the current repository shape, a physical move would affect:

- the Go module path rooted at `dag-core/go.mod`
- many internal Go import paths
- Docker build contexts and bind mounts
- CI workflow references
- local development commands and existing contributor muscle memory

That churn is operationally expensive and does not improve structural truth by itself, because the conceptual ownership is already documented elsewhere.

## Decision

`dag-core/` and `worker-py/` remain at their current physical top-level paths for now.

They are treated conceptually as the repository's reference implementation block, but their filesystem paths are preserved until a future move produces clear operational value beyond path alignment.

## Consequences

- documentation must keep distinguishing conceptual ownership from current filesystem paths
- implementation guidance continues to live under `implementations/*/README.md`
- local scripts, Docker assets, CI references, and Go module paths stay stable
- a future physical move should be treated as an explicit migration project, not as a routine cleanup step
