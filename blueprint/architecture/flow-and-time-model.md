# Flow And Time Model

## Purpose

This document defines how flow, events, state, and time are modeled across the repository.

## Core Model

- processing is decomposed as flows
- state is observed along a time axis
- events and state transitions are first-class design elements
- execution flow and analysis flow are related but distinct

## Execution Model

The current reference runtime separates single-cycle execution from multi-cycle orchestration.

### Runner

Runner owns one execution cycle.

Its cycle is:

1. clear artifact state for the cycle
2. inject inputs into artifact storage
3. begin a state transaction for the partition
4. execute the pipeline in order
5. commit on success
6. rollback on failure

### Driver

Driver owns repeated execution across an event stream.

- it consumes events
- it invokes runner per cycle
- it decides whether to continue on error according to runtime policy

## Event Model

Events are the common unit across HTTP input, timer input, and asynchronous processing.

The minimal event shape is:

- `EventID`: idempotency and record identity
- `EventTime`: logical or physical time
- `Partition`: state scope
- `Type`: event kind
- `Payload`: typed payload

Clock-derived input should also be converted into events rather than bypassing the event model.

## Task And Event Topic Split

Asynchronous flow should separate command-like task requests from result events.

- `dagruntime-tasks`: consumed by workers for execution requests such as `task.requested`
- `dagruntime-events`: consumed by the runtime for result events such as `task.completed` and `task.failed`

This split keeps worker and driver responsibilities explicit.

## Artifact And State Model

Artifacts and state have different lifetimes.

### Artifact

- artifact is cycle-local temporary data
- it is cleared per cycle
- it should be accessed through typed keys

### State

- state is durable across cycles
- it is updated through transactions
- transactions are scoped by partition

Partition is mandatory because state is scoped to units such as symbol, workflow instance, or session.

## Intent Events

Intent events are part of the system model and should be treated as durable semantic events, not only as log text.

Representative event families include:

- `clock.tick.received`
- `dag.run.started`
- `dag.run.state_changed`
- `dag.run.finished`
- `dag.run.failed`
- `dag.node.started`
- `dag.node.state_changed`
- `dag.node.finished`
- `dag.node.failed`
- `dag.node.timeout`
- `dag.node.skipped`

## Boundary Rules

- event flow must remain explicit across synchronous and asynchronous boundaries
- time-series data design must support operational observation as well as later analysis
- intent events are part of the structural model, not only logging output

## Migration Notes

This document is the durable home for repository-level flow, event, artifact, and time-model guidance.
