# AI Pairing Checklist

Use this checklist for repository-level AI pair-programming work.

## Read First

1. Read root `AGENTS.md`.
2. Read the relevant `blueprint/` files before implementation details.
3. Read `governance/policies/` when the task may affect repository-wide rules.
4. Read `scenarios/` when the task changes validation flow or representative usage.
5. Read implementation-local `AGENTS.md` before editing code under `implementations/`.

## Decide Where To Edit

1. Edit `blueprint/` if the change affects boundaries, flow/time model, or inherited rules.
2. Edit `governance/` if the change affects classification, placement, or repository-wide working rules.
3. Edit `platform/` if the change affects observability platform definition.
4. Edit `deployments/` if the change affects local stack wiring or execution wrappers.
5. Edit `data/` if the change affects durable schema, migrations, or analysis queries.
6. Edit `implementations/` only when the change stays within existing boundaries.

## Before Finishing

1. Check whether a higher-level doc should be updated before or with the implementation change.
2. Check whether `README.md` needs a routing or entrypoint update.
3. Keep `.codex/` notes operational and short; move durable rules into `blueprint/` or `governance/`.
