# ADR-0004 Shared Task Context And Local Artifact Placement

## Status

Accepted

## Context

The repository already defines `blueprint/` as structural truth, `governance/tasks/` as temporary shared task context, and `implementations/` as canonical implementation roots.

However, the working tree still exposed two sources of contributor confusion:

- legacy references to `.wrk/` as if it were a repository reading surface
- implementation-local cache and runtime output paths that made canonical implementation roots look mixed with disposable artifacts

This weakens both developer onboarding and AI reading flow because shared task notes, local scratch notes, and generated runtime output are not clearly separated.

## Decision

Shared temporary task context must live under `governance/tasks/`.

Local untracked runtime artifacts and caches must live under repository-root `var/`.

`.wrk/` is treated as a legacy local scratch area only. It must not be referenced by tracked repository documentation as an active source of truth or shared workflow surface.

`implementations/` remains the only canonical implementation root. Platform-owned directories must not introduce parallel implementation roots such as `platform/implementations/`.

## Consequences

- contributors read shared temporary context from `governance/tasks/` rather than searching hidden local-note areas
- local caches, JSONL observation outputs, and similar disposable artifacts move out of implementation roots over time
- tracked docs and task notes should point examples to `var/` instead of implementation-local `tmp/` paths where possible
- empty or misleading parallel directory names under other root areas should be removed rather than preserved as placeholders
