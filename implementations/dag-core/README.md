# dag-core

This directory is the canonical location for the Go reference implementation.

## Role

- reference API and DAG runtime implementation
- concrete example of the structural rules defined in `blueprint/`

## What This Implementation Currently Covers

- HTTP entrypoint for DAG execution
- runtime orchestration through runner and driver
- event-driven execution with Kafka-backed transport
- YAML-based workflow spec loading and compile pipeline
- state handling and persistence integration
- observability integration through the application container
- placeholder transports for gRPC, GraphQL, and WebSocket

## Structural Mapping

The current Go implementation expresses the repository structure through:

- `internal/domain/`: runtime and observability core models
- `internal/application/`: use cases and ports
- `internal/infrastructure/`: storage, event transport, telemetry, and external adapters
- `internal/interface/`: transport-facing handlers and protocol boundaries
- `internal/di/`: composition root

## What Derivative Projects Should Usually Keep

- layer boundaries
- event-driven runtime model
- explicit separation between domain, application, infrastructure, and interface
- observability as a built-in concern

## What Derivative Projects Should Usually Replace

- HTTP handlers and DTO details
- workflow definitions and payload conventions
- product-specific transport exposure
- product-specific persistence and deployment settings

## Workflow Specs (YAML)

- Workflow specs are stored under `implementations/dag-core/workflows/`.
- Current examples are `default.yaml` and `sma_cross_signal.yaml`.
- The compile path is `YAML -> WorkflowSpec -> validator -> factory registry -> adapter -> workflow.Compile`.
- Node `kind` must be registered in the builtin registry.
- Supported spec version is currently `v1`.

## Runtime Workflow Selection

- `DAGRUNTIME_WORKFLOW_SPEC_PATH` selects which workflow YAML to load at startup.
- Default value is `workflows/default.yaml`.
- Example: `DAGRUNTIME_WORKFLOW_SPEC_PATH=workflows/sma_cross_signal.yaml`.

## HTTP Input Notes

- `POST /dag/run` accepts generic runtime inputs such as `symbol`, `mode`, and optional `bars`.
- `POST /marketdata/timeframe-bars:backfill` is the direct marketdata backfill entrypoint and returns aggregation counts.
- `POST /marketdata/timeframe-bars:backfill-workflow` is the feature-local workflow entrypoint for marketdata backfill.
- `bars` is mapped to runtime input key `market.bars`.

## Contract Assets

This implementation owns implementation-facing contract assets such as:

- OpenAPI generation inputs and outputs
- gRPC proto definitions and generated code
- GraphQL schema and generated code
- typed transport contracts for WebSocket

Generated artifacts are implementation assets, not blueprint assets.

## OpenAPI Outputs

OpenAPI outputs belong under:

- `implementations/dag-core/openapi/`

They should be generated from the Go HTTP implementation and reviewed as implementation artifacts.

## Read With

- `blueprint/architecture/system-boundaries.md`
- `blueprint/architecture/flow-and-time-model.md`
- `blueprint/conventions/implementation-principles.md`
- `blueprint/conventions/observability-conventions.md`
- `blueprint/conventions/data-conventions.md`

## Migration Notes

OpenAPI outputs are implementation-owned artifacts under `implementations/dag-core/openapi/`.

Physical-path alignment policy is defined in:

- `blueprint/adr/ADR-0003-implementation-directories-under-implementations.md`
