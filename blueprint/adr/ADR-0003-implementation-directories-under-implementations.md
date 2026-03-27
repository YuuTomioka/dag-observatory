# ADR-0003 Implementation Directories Under `implementations/`

## Status

Accepted

## Supersedes

- `ADR-0002-reference-implementation-paths.md`

## Context

The parent-repository structure defines `implementations/` as the source-of-truth area for reference implementations.

However, the physical filesystem still has active implementation roots at top-level paths (`dag-core/`, `worker-py/`), while `implementations/` partially holds documentation and generated assets.
This mixed shape weakens the repository contract for contributors and AI pair-programming flows.

An external decision has been made to adopt the structure policy that requires physical alignment, not conceptual alignment only.

## Decision

`dag-core/` and `worker-py/` must be physically moved under `implementations/`.

Canonical implementation roots are:

- `implementations/dag-core/`
- `implementations/worker-py/`

`deployments/`, scripts, CI, and developer commands must reference those canonical paths.

## Consequences

- existing top-level path assumptions must be removed from docs and tooling
- migration must be executed as an explicit repository task with path-fix and verification steps
- temporary ambiguity between conceptual and physical ownership is eliminated
- future implementation additions should be introduced directly under `implementations/`
