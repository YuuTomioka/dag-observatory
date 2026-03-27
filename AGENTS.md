# AGENTS

## Purpose

This repository is a parent repository for derivative projects around DAG + time-series systems.

Code is not the source of truth. `blueprint/` is the primary source of structural truth.

## Read Order

1. `README.md`
2. `blueprint/identity/repository-purpose.md`
3. `blueprint/architecture/system-boundaries.md`
4. `blueprint/architecture/flow-and-time-model.md`
5. Relevant files under `blueprint/conventions/`
6. Relevant files under `blueprint/adr/` when a past structural decision may affect the task
7. `governance/policies/` for repository-wide change and documentation rules
8. `scenarios/` when validating or explaining representative usage
9. Then the target implementation, platform, data, deployment, or script directory

## AI Working Order

AI contributors should interpret the repository in this order:

1. principles in `blueprint/`
2. repository rules in `AGENTS.md` and `governance/policies/`
3. representative usage in `scenarios/`
4. implementation-local rules under the target directory
5. implementation code and runtime assets

Do not start from implementation details when the task could affect repository structure, boundaries, conventions, or durable operating assumptions.

## Change Rules

- Update `blueprint/` first when a change affects structure, boundaries, or conventions.
- Update `blueprint/adr/` when a durable repository-level decision or rationale changes.
- Update `governance/policies/` when the repository-wide change-handling rule changes.
- Keep implementation-specific details under `implementations/`.
- Keep runtime and observability environment details under `platform/`.
- Keep execution wrappers and local stack wiring under `deployments/`.
- Keep time-series design assets under `data/`.
- Keep AI-local workflow helpers under `.codex/` and avoid treating them as structural source of truth.
- Do not reintroduce `.docs/`; temporary work context belongs in `governance/tasks/`.

## Documentation Rules

- `README.md` is an entrypoint, not the full specification.
- Long-lived decisions belong in `blueprint/adr/`.
- Scenario-driven guidance belongs in `scenarios/`.
- Work-in-progress context belongs in `governance/tasks/`.
- `.codex/` may hold AI workflow helpers, but repository structure and design truth must remain in `blueprint/` and `governance/`.
