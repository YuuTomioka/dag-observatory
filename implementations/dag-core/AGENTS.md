# AGENTS

## Purpose

This directory is the canonical Go reference implementation for API and DAG runtime behavior.

Repository-wide structure is defined above this directory. Read root `AGENTS.md` and relevant `blueprint/` documents first.

## Local Read Order

1. `README.md`
2. root `blueprint/architecture/system-boundaries.md`
3. root `blueprint/architecture/flow-and-time-model.md`
4. root `blueprint/conventions/implementation-principles.md`
5. then the target package under `internal/` or `cmd/`

## Local Rules

- Keep transport-specific code under `internal/interface/`.
- Keep use-case composition and ports under `internal/application/`.
- Keep runtime concepts and policies under `internal/domain/`.
- Keep concrete adapters under `internal/infrastructure/`.
- Do not introduce reverse dependencies from domain into infrastructure.
- Keep generated contract artifacts aligned with implementation changes.

## Change Hints

- If a change affects runtime boundaries or repository conventions, update root `blueprint/` first.
- If a change is transport-local or adapter-local, stay within this implementation unless a wider rule changes.
