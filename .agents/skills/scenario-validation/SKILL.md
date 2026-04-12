---
name: scenario-validation
description: Validate dag-observatory work through repository scenarios and observability checks. Use when a change should be checked through representative flows, acceptance-style validation, or operator-facing run procedures.
---

# Scenario Validation

Use this skill when validation should follow repository scenarios instead of only unit tests.

## Read Order

1. Read root `AGENTS.md`.
2. Read relevant files under `scenarios/`.
3. Read related `README.md` or local `AGENTS.md` files for the target implementation, data, or platform area.
4. Read `blueprint/conventions/observability-conventions.md` when telemetry or correlation checks matter.

## Validation Workflow

1. Choose the smallest scenario that exercises the requested behavior.
2. State the expected outcome in user or operator terms.
3. Identify which commands, APIs, or flows must be run.
4. Identify which logs, traces, metrics, or persisted results should confirm success.
5. Record gaps if the repository lacks a concrete acceptance scenario and avoid inventing fake procedural truth.

## Output

- validation scenario used
- commands or steps executed
- observability evidence checked
- remaining gaps that should become future `scenarios/` assets
