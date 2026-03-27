# worker-py

This directory is the canonical location for the Python worker reference implementation.

## Role

- reference asynchronous worker implementation
- concrete example of event-driven and trace-propagated execution

## What This Implementation Currently Covers

- Kafka-based task consumption
- Kafka-based result event publication
- trace context extraction and propagation across task handling
- task dispatching through a registry-based handler model
- TSDB-related query usage from shared SQL assets
- outbox publishing and artifact GC companion processes

## Worker Design Notes

The current worker model is based on:

- explicit event contract handling
- registry-based routing from event type to handler
- consumption and publication through Kafka-backed transport
- keeping task execution separate from transport glue

For derivative projects, this means the worker should stay a computation engine, not the owner of workflow truth.

It should consume delegated work, execute domain-specific handlers, and publish result events back into the shared runtime flow.

## Structural Mapping

The current Python implementation expresses the repository structure through:

- `src/worker/runtime/`: task dispatch execution
- `src/worker/kafka/`: transport integration
- `src/worker/contracts/`: event contract types
- `src/worker/db/`: shared SQL access
- `src/worker/tasks/`: task handlers
- `src/worker/outbox/`: deferred event publication
- `src/worker/minio/`: artifact lifecycle helpers

## What Derivative Projects Should Usually Keep

- separation between contract, runtime, transport, and task handler code
- event-driven execution model
- trace propagation across async boundaries
- reliance on shared data assets rather than local schema truth

## What Derivative Projects Should Usually Replace

- concrete task handlers
- payload conventions for business-specific tasks
- deployment-time topic names and environment settings
- product-specific artifact handling policies

## Read With

- `blueprint/architecture/flow-and-time-model.md`
- `blueprint/conventions/implementation-principles.md`
- `blueprint/conventions/observability-conventions.md`
- `blueprint/conventions/data-conventions.md`

## Migration Notes

Physical-path alignment policy is defined in:

- `blueprint/adr/ADR-0003-implementation-directories-under-implementations.md`
