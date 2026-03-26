# Transport Followups

## Background

Migration-era transport expansion notes covered HTTP refactoring, transport layering, contract generation, and multi-transport growth in a single process.

The durable structural parts have already been promoted into `blueprint/`.
This task document keeps only the remaining operational or product-planning follow-ups.

## In Scope

- unresolved follow-ups around CI enforcement and transport-specific rollout details
- implementation backlog that does not need to remain as permanent design truth

## Out Of Scope

- permanent transport layering rules already captured in `blueprint/`
- historical implementation notes that no longer affect current decisions

## Follow-Ups

- decide whether additional CI enforcement is needed for `contracts-check` and gRPC contract validation
- decide whether OpenAPI should remain v2 or move to a v3-oriented toolchain later
- define the minimum real API surface for GraphQL, gRPC, and WebSocket beyond placeholder skeletons
- clarify WebSocket message schema/versioning only when actual external use requires it

## Acceptance Criteria

- future contributors can find the remaining transport backlog without consulting `.docs`
- the old transport-expansion notes are no longer required for repository guidance

## Result

- migration-era transport planning was reduced to durable blueprint guidance plus this compact follow-up list
