# AGENTS

## Purpose

This directory is the canonical home for shared TSDB schema, migration, seed, and query assets.

Repository-wide structure is defined above this directory. Read root `AGENTS.md` and relevant `blueprint/` documents first.

## Local Read Order

1. `README.md`
2. root `blueprint/architecture/system-boundaries.md`
3. root `blueprint/conventions/data-conventions.md`
4. `schema/README.md`
5. `query/README.md`
6. then the target SQL asset under `schema/` or `query/`

## Local Rules

- Treat SQL assets here as the source of truth for data design.
- Put schema evolution in migration SQL under `schema/migrate/`.
- Keep seed data under `schema/seed/` and do not mix it with durable schema change history.
- Keep shared query definitions under `query/` so implementations consume the same data contract.
- Do not move schema truth into generated code, ORM models, or implementation-local notes.

## Change Hints

- If a change alters repository-wide data conventions or storage ownership, update root `blueprint/` first.
- If a change only updates concrete schema, migration, seed, or query assets within existing conventions, keep it local to this directory.
- If a change modifies migrator wrappers or compose execution shells, update `deployments/` rather than treating them as data-owned truth.
