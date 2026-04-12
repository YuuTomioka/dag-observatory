# AGENTS

## Purpose

This repository is a parent repository for derivative projects around DAG + time-series systems.

Code is not the primary structural source of truth. Read and update the repository from `blueprint/` and `governance/` first when a task can affect structure, boundaries, conventions, or durable operating assumptions.

## Read Order

1. `README.md`
2. `blueprint/identity/repository-purpose.md`
3. `blueprint/architecture/system-boundaries.md`
4. `blueprint/architecture/flow-and-time-model.md`
5. relevant files under `blueprint/conventions/`
6. relevant files under `blueprint/adr/` when past structural decisions matter
7. `governance/policies/change-policy.md`
8. `governance/policies/documentation-policy.md`
9. relevant files under `scenarios/`
10. then the target implementation, platform, data, deployment, or script directory

## AI Working Rules

- Start from repository principles, not implementation details, when the task might affect structure or long-lived rules.
- Treat `README.md` as an entrypoint only, not as the full specification.
- Use `scenarios/` when validating or explaining recommended usage.
- Prefer the smallest boundary that correctly contains the change.
- Do not invent parallel sources of truth in implementation docs, scratch notes, or generated output.

## Change Classification

Classify the task in this order before editing:

1. repository structure, boundaries, or inheritance rules
2. repository-wide documentation or contributor operating rules
3. representative usage or validation flow
4. implementation-local, platform-local, data-local, or deployment-local change

If the task falls into a higher class, update that layer first before editing lower layers.

## Placement Rules

- `blueprint/`: long-lived design truth, boundaries, conventions, and ADRs
- `governance/policies/`: repository-wide change-handling and documentation rules
- `governance/contracts/`: reusable human-facing request and handoff templates
- `governance/guides/`: reusable human-facing operating guides and how-to entrypoints
- `governance/tasks/`: shared temporary task context
- `implementations/`: implementation-local code and guidance
- `platform/`: operating environment and observability reference assets
- `deployments/`: local stack wiring and execution wrappers
- `data/`: durable SQL, migration, and analysis assets
- `scenarios/`: Japanese-facing representative flows and validation guidance
- `.agents/skills/`: repository-scoped Codex skills
- `.codex/`: AI-local helpers only
- `var/`: untracked local runtime outputs and caches

## Documentation Rules

- Promote durable conclusions into the proper permanent location.
- Keep temporary planning notes in `governance/tasks/` until promoted or retired.
- Keep durable source-of-truth documents in English.
- Keep direct human-interface surfaces such as `scenarios/` and `governance/contracts/` in Japanese.
- Keep reusable human-facing operating guides under `governance/guides/` in Japanese.
- Do not use `.wrk/` as a tracked or referenced documentation surface.
- Do not reintroduce `.docs/`.
- Do not treat generated artifacts or runtime output as structural truth.

## Practical Editing Rules

- Update `blueprint/` first when a change affects structure, boundaries, conventions, or derivative inheritance.
- Update `blueprint/adr/` when repository-level rationale or durable decisions change.
- Update `governance/policies/` when repository-wide contributor rules change.
- Update `governance/contracts/` when reusable human-facing request or handoff templates change.
- Update `scenarios/` when the recommended entry flow or validation procedure changes.
- Keep implementation-specific details under `implementations/`.
- Keep platform topology and backend configuration under `platform/`.
- Keep local compose wiring and migrator wrappers under `deployments/`.
- Keep time-series schema, migrations, and shared queries under `data/`.
