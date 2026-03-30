# Representative Scenario

## Purpose

This scenario will describe the most representative flow of this parent repository.

## Candidate Scenario

- trigger a DAG execution
- emit and propagate execution events
- execute asynchronous work
- record time-series data
- observe the run through telemetry

## Scenario Outline

1. an external request or clock input starts a run
2. the API converts the input into runtime events
3. task requests are published for asynchronous execution
4. the worker processes the task and publishes result events
5. the runtime consumes the result events and advances state
6. telemetry and time-series data make the run observable

## HTTP Entrypoint Guidance

When both feature-local and generic workflow entrypoints exist, the recommended flow is:

1. use the feature-local direct endpoint when the caller needs immediate domain results
2. use the feature-local workflow endpoint when the caller needs workflow execution under the feature namespace
3. use the generic workflow endpoint only when no feature-local workflow entrypoint exists or when the caller is intentionally operating at the runtime level

For the current marketdata timeframe-bar backfill flow this means:

- `POST /marketdata/timeframe-bars:backfill` is the primary direct execution entrypoint
- `POST /marketdata/timeframe-bars:backfill-workflow` is the primary workflow execution entrypoint
- `POST /dag/run` is a generic runtime entrypoint and is not the recommended first choice for marketdata backfill

## Smoke Test Shape

The representative scenario should be verifiable across multiple execution modes.

Recommended request modes:

- success path
- failure path
- timeout path
- skip path

Recommended checks:

- logs can be searched by event type
- logs can be searched by `dag.run_id`
- traces can be found by `trace_id`
- error cases surface error status in traces
- HTTP and DAG metrics increase as expected

The purpose of this scenario is not only to prove execution, but also to prove cross-signal observability consistency.

Additional end-to-end checks may include:

- artifact metadata is written after worker execution
- outbox publication completes after database confirmation
- presigned access flow is available for stored artifacts when relevant

## Migration Notes

This document is the durable home for the repository's representative end-to-end validation scenario.
