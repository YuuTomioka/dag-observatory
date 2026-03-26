# ADR-0001 Parent Repository Positioning

## Status

Accepted

## Context

This repository has accumulated implementation code, operational assets, and design notes.

To support future derivative projects, it must be treated as a parent repository with reusable structural guidance rather than as a single evolving product.

## Decision

The repository is redefined as a parent repository for DAG + time-series systems.

Its permanent structure is organized around:

- `blueprint/` for source-of-truth design assets
- `implementations/` for reference implementations
- `platform/` for reference operating environments
- `data/` for time-series design assets
- `scenarios/` for entrypoint use cases
- `governance/` for change control

`.docs/` was treated as temporary migration-era storage during the restructuring and has been removed after migration.

## Consequences

- Long-lived design guidance moves out of the old migration-era note area and into permanent repository locations
- README becomes an entrypoint rather than the main specification
- implementation directories become examples, not the conceptual center
- future changes must consider whether `blueprint/` needs updating first

## Inherited Runtime Intent

The parent repository retains the runtime intent that motivated the earliest DAG runtime documents:

- deterministic behavior should be preferred for the same input stream
- state integrity should be preserved at cycle boundaries through commit and rollback
- clock, store, and observer concerns should remain replaceable
- runtime behavior should stay aligned with observability concerns

These points are now expressed through `blueprint/` documents rather than through a standalone implementation-era index document.
