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
6. Then the target implementation or platform directory

## Change Rules

- Update `blueprint/` first when a change affects structure, boundaries, or conventions.
- Keep implementation-specific details under `implementations/`.
- Keep runtime and observability environment details under `platform/`.
- Keep time-series design assets under `data/`.
- Do not reintroduce `.docs/`; temporary work context belongs in `governance/tasks/`.

## Documentation Rules

- `README.md` is an entrypoint, not the full specification.
- Long-lived decisions belong in `blueprint/adr/`.
- Scenario-driven guidance belongs in `scenarios/`.
- Work-in-progress context belongs in `governance/tasks/`.
