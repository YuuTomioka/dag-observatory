# Documentation Policy

## Placement Rules

- `README.md`: repository entrypoint
- `blueprint/`: long-lived design truth
- `blueprint/adr/`: durable decisions and rationale
- `implementations/`: implementation-local guidance
- `platform/`: environment and operating reference
- `data/`: time-series and analysis assets
- `scenarios/`: recommended entrypoint use cases
- `governance/contracts/`: human-facing request and handoff templates
- `governance/guides/`: human-facing operating guides and navigation
- `.agents/skills/`: repository-scoped Codex skills
- `.codex/`: AI-local workflow helpers that do not become structural source of truth
- `governance/tasks/`: temporary task context
- `var/`: untracked local runtime outputs and caches, not documentation
- derivative or migration-era proposal notes: `governance/tasks/` until promoted or retired

## Language Policy

- durable source-of-truth documents stay in English
- this includes `blueprint/`, `blueprint/adr/`, repository structure rules, and implementation, platform, and data source-of-truth documents
- human-facing request intake, handoff, and operating procedure surfaces use Japanese
- `scenarios/` is treated as a Japanese-facing validation and operation entry surface
- `governance/contracts/` uses Japanese as the default language
- `governance/guides/` uses Japanese as the default language
- when a document's role is ambiguous between durable specification and direct human-interface guidance, prefer Japanese
- each document should have one primary language based on its role
- avoid maintaining parallel English and Japanese documents with duplicated truth
- Japanese human-facing documents should point to English source-of-truth documents instead of redefining repository structure or technical truth

## Migration Rule

`.docs/` is no longer used in this repository.

Temporary or migration-era work context should go to `governance/tasks/`.

`.wrk/` is a legacy local scratch area and should not be used as a tracked documentation surface or referenced as active repository guidance.

Disposable runtime output, caches, and local observation files should go to ignored `var/` instead of implementation-owned paths where practical.

When a document becomes durable, move or rewrite it into the proper permanent location.
