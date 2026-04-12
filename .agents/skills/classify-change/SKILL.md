---
name: classify-change
description: Classify repository changes for dag-observatory before editing. Use when a request may affect structure, repository-wide rules, scenarios, or implementation-local work and you need to decide which layer to update first.
---

# Classify Change

Use this skill before editing when the correct change boundary is unclear.

## Read Order

1. Read root `AGENTS.md`.
2. Read `governance/policies/change-policy.md`.
3. Read `governance/policies/documentation-policy.md` if document placement or language is involved.
4. Read relevant `blueprint/` or `scenarios/` files only after classifying the change type.

## Classification

Classify in this order:

1. repository structure, boundaries, or inheritance rules
2. repository-wide documentation or contributor operating rules
3. representative usage or validation flow
4. implementation-local, platform-local, data-local, or deployment-local work

If the request matches a higher class, update that layer first before editing lower layers.

## Output

State:

- the chosen class
- the directories that should be edited first
- any higher-layer documents that must be updated before implementation
- whether the human-facing output should be Japanese or the durable source of truth should stay English
