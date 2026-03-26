# Runtime Followups

## Background

Migration-era review notes identified several follow-up items after the initial DAG runtime, store integration, and Python worker work.

Most structural guidance has already been promoted into `blueprint/`.
This task document keeps only the remaining implementation follow-ups that may still matter for future work.

## In Scope

- runtime and store operation follow-ups that were not promoted into permanent design rules
- worker-specific implementation follow-ups that remain useful as near-term backlog

## Out Of Scope

- permanent structural rules already captured in `blueprint/`
- superseded migration-era implementation history
- completed work that no longer changes current planning

## Follow-Ups

### Runtime And Store

- clarify whether any remaining payload normalization logic should be centralized further in the use-case layer
- confirm whether additional consumer/driver integration tests are still needed beyond current coverage
- decide whether any Kafka poison-message handling should be formalized beyond current documented policy
- keep watching whether `dag-core-api` remains too central an integration point as the runtime grows
- decide whether worker-side operational roles such as outbox and GC should stay co-located or be separated further
- add concise operational guidance for migrator, GC, and backup handling if these flows become active maintenance concerns

### Worker

- decide whether Python OTLP exporter improvements are still needed
- split task registry if task surface grows materially
- tighten error payload code conventions if cross-language consumers need stronger guarantees

## Acceptance Criteria

- remaining work is small enough that migration-era review notes can be removed
- future contributors can find open implementation follow-ups without consulting `.docs`

## Result

- migration-era review notes were reduced to this compact task document
- durable rules were moved into `blueprint/`, `implementations/`, `platform/`, and `data/`
- completed specification-gap lists are no longer kept as active design documents once their contents are absorbed or closed
