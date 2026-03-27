# AGENTS

## Purpose

This directory is the canonical Python reference worker implementation for asynchronous task execution and follow-up workers.

Repository-wide structure is defined above this directory. Read root `AGENTS.md` and relevant `blueprint/` documents first.

## Local Read Order

1. `README.md`
2. root `blueprint/architecture/system-boundaries.md`
3. root `blueprint/architecture/flow-and-time-model.md`
4. root `blueprint/conventions/observability-conventions.md`
5. then the target module under `src/worker/`

## Local Rules

- Keep Kafka integration, TSDB access, and MinIO access behind focused modules under `src/worker/`.
- Keep task registration and task implementations explicit under `src/worker/tasks/`.
- Preserve trace-context propagation and event-shape compatibility with `dag-core`.
- Treat outbox publication and GC workers as implementation-local runtime workers, not repository-wide platform assets.

## Change Hints

- If a change alters event contracts, runtime semantics, or shared observability rules, update root `blueprint/` first.
- If a change stays within worker execution behavior or adapter details, keep it local to this directory.
