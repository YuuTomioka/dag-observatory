# Workflow Summary API Followups

## Background

A derivative-planning note described a summary API that would synthesize workflow-run state, task dispatch status, attempts, and artifacts for high-frequency reads.

This is useful implementation backlog, but it is not permanent parent-repository design truth yet.

## In Scope

- derivative-scoped follow-ups for a workflow summary read API
- read-model and query-shape questions that may matter when a product needs polling-friendly run summaries

## Out Of Scope

- treating this endpoint as a mandatory repository-wide interface today
- locking product-specific response fields too early
- permanent runtime structure rules already covered by `blueprint/`

## Follow-Ups

- decide whether a derivative needs a `workflow-runs/{id}/summary` style endpoint or an equivalent read model
- if adopted, keep effective task status as a synthesis of dispatch and attempt data rather than duplicating state ownership
- prefer a small response surface with counts, task state, and artifact metadata, while keeping presigned download URLs out of the summary response
- start with simple query composition and revisit dedicated read models or caching only if real load requires it
- define retry and latest-attempt selection rules before implementing any summary aggregation

## Acceptance Criteria

- future derivative work can recover the intent of the original summary-API proposal without consulting the old `.docs` note
- the proposal stays clearly in backlog form until a concrete implementation needs it

## Result

- the original summary-API proposal was reduced to a compact implementation follow-up instead of a standing migration document

