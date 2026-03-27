# Implementation Change Checklist

Use this checklist when editing code under `implementations/`.

## Read First

1. Read root `AGENTS.md`.
2. Read the local `AGENTS.md` under the target implementation.
3. Read the relevant `blueprint/` files if the change may affect runtime semantics, boundaries, or shared conventions.
4. Read the implementation `README.md` for local commands and entrypoints.

## Decide Scope

1. Stay in `implementations/` only if the change fits existing boundaries.
2. Update `blueprint/` first if the change alters event semantics, flow/time model, observability assumptions, or subsystem ownership.
3. Update `data/` if the change alters durable schema, migration assets, or analysis queries.
4. Update `deployments/` or `platform/` only if runtime wiring or observability platform config must change with the implementation.

## Before Finishing

1. Check whether generated contracts or schema-derived artifacts must be refreshed.
2. Check whether implementation-local README or AGENTS guidance should change.
3. Check whether repository entry docs need a routing update.
