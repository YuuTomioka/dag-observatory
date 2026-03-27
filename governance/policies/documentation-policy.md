# Documentation Policy

## Placement Rules

- `README.md`: repository entrypoint
- `blueprint/`: long-lived design truth
- `blueprint/adr/`: durable decisions and rationale
- `implementations/`: implementation-local guidance
- `platform/`: environment and operating reference
- `data/`: time-series and analysis assets
- `scenarios/`: recommended entrypoint use cases
- `.codex/`: AI-local workflow helpers that do not become structural source of truth
- `governance/tasks/`: temporary task context
- derivative or migration-era proposal notes: `governance/tasks/` until promoted or retired

## Migration Rule

`.docs/` is no longer used in this repository.

Temporary or migration-era work context should go to `governance/tasks/`.

When a document becomes durable, move or rewrite it into the proper permanent location.
