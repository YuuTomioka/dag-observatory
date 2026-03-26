# Store Followups

## Background

Migration-era store planning covered TimescaleDB, MinIO, transactional outbox, artifact upload sessions, GC, and end-to-end integration.

The durable structural parts have already been promoted into `data/` and `blueprint/`.
This task document keeps only the remaining follow-up questions that are still operational rather than structural.

## In Scope

- remaining operational follow-ups around migrator, GC, backups, and end-to-end validation
- implementation backlog that does not need to live as permanent design truth

## Out Of Scope

- permanent SQL and storage design rules
- already absorbed repository-wide data conventions
- closed migration-era planning history

## Follow-Ups

- decide whether migrator, GC, and backup procedures need more concise operational documentation
- revisit whether read-model or caching strategies are needed only when real load justifies them
- keep artifact upload session GC behavior under review if object lifecycle grows more complex
- review whether outbox processing should remain separate or evolve with future scaling needs

## Acceptance Criteria

- future maintainers can see the remaining store-related backlog without reopening `.docs`
- the old store planning documents are no longer needed for day-to-day repository guidance

## Result

- migration-era store planning was reduced to durable design guidance plus this short follow-up task list
