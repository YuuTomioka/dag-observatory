# System Boundaries

## Purpose

This document defines the main subsystem boundaries of the parent repository.

## Primary Subsystems

- `implementations/dag-core/`: reference API and DAG runtime implementation
- `implementations/worker-py/`: reference worker implementation for asynchronous execution
- `platform/observability/`: reference operating environment for telemetry and local visibility
- `data/tsdb/`: time-series and analysis design assets

## Layer Model

The reference Go implementation follows a layered model:

- `domain`: core specification such as workflow, node, runner, driver, policy, event, artifact, and state
- `application`: use-case composition and external ports
- `infrastructure`: concrete implementations such as store, Kafka, clock, and telemetry bridges
- `interface`: external entrypoints such as HTTP

The dependency direction is:

- `interface -> application -> domain`
- `infrastructure -> domain/application`

Reverse dependency from domain into infrastructure is not allowed.

## Bounded Contexts Inside The Runtime

Within the current runtime implementation, `dagruntime` and `observability` are treated as separate bounded contexts.

- `dagruntime` owns workflow execution, state progression, and event-driven runtime behavior
- `observability` owns telemetry semantics, emission, and correlation rules

They are connected through ports instead of direct infrastructure coupling.

## Boundary Rules

- API, worker, platform, and data are separate change units.
- Execution concerns and analysis concerns must be separated deliberately.
- Synchronous request handling and asynchronous task execution must not be mixed casually.
- Observability is part of the architecture, not an optional afterthought.

## Responsibility Boundaries

- API receives external input and translates it into runtime actions
- worker consumes execution requests and emits result events
- platform provides telemetry transport and visibility topology
- data owns durable schema, migration, and query assets for time-series analysis

These boundaries should remain readable even when implementations change.

## Multi-Transport Direction

When multiple transports coexist in one process:

- transport-specific request and response handling stays under `internal/interface/*`
- use cases remain the shared command and query entrypoint
- transport growth must not introduce transport concepts into domain or application layers
- WebSocket-style delivery must not become the sole source of state truth

## Current As-Is Runtime Topology

The current local development topology is centered on:

- `dag-core-api` as API and DAG orchestration entrypoint
- `worker-py` as asynchronous execution worker
- `worker-py-outbox` as deferred event publication worker
- `worker-py-gc` as cleanup worker
- `redpanda` as event broker
- `timescaledb` and `minio` as persistence services

In the current As-Is shape, `dag-core-api` is still a major integration point because it touches HTTP, event transport, persistence, and telemetry concerns at once.

## Migration Notes

This document is the durable home for repository-level structural boundary guidance.
