# AuthN AuthZ Followups

## Background

A derivative-planning note proposed splitting authentication and entitlement concerns into separate modules for future product work.

This is not parent-repository structural truth yet.
The note is kept here as a compact backlog item until a derivative project actually needs the design.

## In Scope

- derivative-scoped follow-ups around module boundaries between authentication and authorization
- minimum interface and policy questions that may matter if a future implementation adopts the split

## Out Of Scope

- treating this proposal as a repository-wide mandatory structure today
- detailed product-specific identity provider setup
- permanent authorization conventions beyond what `blueprint/` already defines

## Follow-Ups

- decide whether a future derivative should split `auth` and `entitlement` into separate modules instead of a single access-control area
- if adopted, keep identity verification limited to token validation and subject-context extraction
- if adopted, keep contract, feature, role, and quota decisions in one entitlement source of truth rather than in token claims
- define the minimum ports between business modules and any future authorization service before implementation starts
- add deny-path observability requirements if a derivative promotes this design into real APIs

## Acceptance Criteria

- future derivative work can recover the intent of the original proposal without reopening the old `.docs` memo
- this note remains clearly optional until a real derivative product adopts it

## Result

- the original AuthN/AuthZ proposal was reduced to a derivative backlog note instead of being kept as a migration-era report

