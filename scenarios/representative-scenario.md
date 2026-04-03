# Representative Scenario

## Purpose

This document defines the repository's representative end-to-end validation flow.

The scenario is intentionally concrete enough to run as a smoke test, while remaining reusable for derivative projects.

## Representative As-Is Scenario

The current representative flow is **marketdata timeframe-bar backfill with workflow execution**:

1. a caller submits a backfill request
2. the API translates the request into runtime events
3. task requests are published for asynchronous execution
4. worker-side processing emits result events
5. runtime consumes result events and advances state
6. telemetry and time-series records make run progress observable

This scenario validates both execution correctness and cross-signal observability.

## Entrypoint Priority

When both feature-local and generic runtime entrypoints exist, use this order:

1. feature-local direct endpoint for immediate domain results
2. feature-local workflow endpoint for namespaced workflow execution
3. generic runtime endpoint only when feature-local workflow entrypoints are unavailable or runtime-level operation is intentional

For marketdata timeframe-bar backfill:

- primary direct entrypoint: `POST /marketdata/timeframe-bars:backfill`
- primary workflow entrypoint: `POST /marketdata/timeframe-bars:backfill-workflow`
- generic runtime entrypoint (fallback): `POST /dag/run`

## Precondition Checklist

Before running the scenario:

- local stack is up (`make dev-up`)
- migration and schema setup are complete (`data/tsdb/README.md`)
- API and worker services are healthy
- observability services (collector, Loki, Tempo, Grafana) are reachable

## Validation Flow

### Step 1. Submit workflow backfill request

Use `POST /marketdata/timeframe-bars:backfill-workflow` with a bounded time range and explicit mode.

Expected immediate outcome:

- response includes a run identifier
- request is accepted without synchronous heavy processing

### Step 2. Confirm asynchronous task/event progression

Confirm that task request and result events progress through the runtime model.

Representative event progression:

- `task.requested`
- `task.completed` or `task.failed`

Expected outcome:

- run state advances consistently with consumed result events

### Step 3. Confirm data-side effects

Verify that expected timeframe-bar outputs are written for the requested symbol/timeframe/range.

Expected outcome:

- durable records exist in the time-series store
- write shape matches requested mode semantics

### Step 4. Confirm observability correlation

Verify cross-signal correlation by run and trace context.

Expected checks:

- intent logs are searchable by event type and `dag.run_id`
- traces can be reached from log correlation fields (`trace_id`, `span_id`)
- errors/timeouts surface as explicit failure semantics in traces and logs
- relevant HTTP and runtime metrics increase

### Step 5. Confirm artifact/outbox consistency when applicable

For workflows that emit artifacts:

- artifact metadata is written after worker-side completion
- outbox publication follows database confirmation
- presigned retrieval flow is available when artifact storage is enabled

## Smoke Test Modes

The representative scenario should cover multiple execution outcomes over time:

- success path
- failure path
- timeout path
- skip path

If a mode is not currently reproducible in a stable automated way, record that gap in task-level notes under `governance/tasks/` and keep this scenario document as the stable target shape.

## Acceptance Checklist

A representative run is considered validated when all are true:

- recommended entrypoint priority is followed
- asynchronous task/event flow is observable end-to-end
- data-side effects are confirmed for the requested range
- logs/traces/metrics are mutually correlatable by run context
- failure semantics are explicit and searchable

## Maintenance Rule

Update this document when one of the following changes:

- recommended entry flow
- representative repository use case
- operational validation steps

Repository-level structural changes still belong first in `blueprint/` and policy changes in `governance/policies/`.
