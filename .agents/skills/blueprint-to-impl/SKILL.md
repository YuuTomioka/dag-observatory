---
name: blueprint-to-impl
description: Translate blueprint or governance decisions into concrete implementation changes in dag-observatory. Use when a request starts from repository design rules and must be applied to Go, Python, data, platform, or deployment assets.
---

# Blueprint To Implementation

Use this skill when the task starts from repository design truth and must be carried into lower layers.

## Read Order

1. Read root `AGENTS.md`.
2. Read the relevant `blueprint/` and `governance/policies/` documents.
3. Read the target local `AGENTS.md` for `implementations/`, `data/tsdb/`, or `platform/observability/` when present.
4. Read only the target implementation files needed for the change.

## Workflow

1. Summarize the design rule or operating rule that constrains the change.
2. Identify the smallest boundary that can contain the implementation work.
3. Check whether upper-layer docs must be updated first.
4. Apply changes in the target implementation, data, platform, or deployment area.
5. Verify that the lower-layer change does not create a parallel source of truth.

## Guardrails

- Keep `blueprint/` as the structural source of truth.
- Keep `governance/` as the operating-rule source of truth.
- Do not move durable design rules into implementation READMEs or local notes.
- If user-facing procedures are needed, put them in Japanese-facing surfaces such as `scenarios/` or `governance/contracts/`.
