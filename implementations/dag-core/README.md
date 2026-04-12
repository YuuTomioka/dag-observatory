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
- Developer guide for workflow design and extension: `implementations/dag-core/workflows/README.md` (Japanese).
- Representative workflows:
  - `default.yaml` (smoke-test default)
  - `sma_cross_signal.yaml` (basic signal sample)
  - `breakout_long_v0.yaml` (breakout v0 baseline)
  - `breakout_long_execution_position_minimal.yaml` (submit + position snapshot minimal set)
  - `breakout_long_v1_extended.yaml` (extended breakout pipeline)
- The compile path is `YAML -> WorkflowSpec -> validator -> factory registry -> adapter -> workflow.Compile`.
- Node `kind` must be registered in the builtin registry.
- Supported spec version is currently `v1`.

## Runtime Workflow Selection

- `DAGRUNTIME_WORKFLOW_SPEC_PATH` selects which workflow YAML to load at startup.
- Default value is `workflows/default.yaml` (smoke-test workflow).
- Strategy run example:
  - `DAGRUNTIME_WORKFLOW_SPEC_PATH=workflows/breakout_long_execution_position_minimal.yaml`
- Extended strategy run example:
  - `DAGRUNTIME_WORKFLOW_SPEC_PATH=workflows/breakout_long_v1_extended.yaml`

## HTTP Input Notes

- `POST /dag/run` accepts generic runtime inputs such as `symbol`, `mode`, and optional `bars`.
- `POST /marketdata/timeframe-bars:backfill` is the direct marketdata backfill entrypoint and returns aggregation counts.
- `POST /marketdata/timeframe-bars:backfill-workflow` is the feature-local workflow entrypoint for marketdata backfill.
- Prefer feature-local endpoints over `POST /dag/run` when both exist.
- `bars` is mapped to runtime input key `market.bars`.

## Algotrade Validation Read Model

The current algotrade validation path distinguishes between execution-time artifacts and result-side read models.

- `TradeIntent`: normalized strategy intent produced by entry decision flows
- `ExecutionResult`: normalized execution progression derived from that intent
- `ClosedTrade`: completed-trade fact used for validation and comparison
- `StrategySummary`: projection over closed trades for run/strategy comparison

Current durable state keys in `internal/domain/algotrade/` include:

- `pending_orders`
- `open_positions`
- `closed_trades`
- `daily_pnl`
- `strategy_summary`

Current read and execution-facing HTTP paths include:

- `POST /algotrade/backtests:run`
- `POST /algotrade/result-reflection:run`
- `GET /algotrade/trades`
- `GET /algotrade/equity`
- `GET /algotrade/summary`
- `GET /runs/ui`

These paths are intended as validation and inspection surfaces. They do not replace runtime/state truth with UI-owned truth.

## Tick Ingestion Notes

The implementation also exposes a cTrader-oriented tick ingestion entrypoint:

- `POST /v1/ctrader/ticks`

Current request characteristics:

- JSON body, with optional `gzip` content encoding
- required fields: `symbol`, `price_scale`, `ticks[]`
- optional correlation/batching fields: `request_id`, `day`, `batch_seq`
- each tick carries `time`, `bid`, `ask`

Validation rules currently enforced:

- `symbol` must resolve to a known symbol
- `price_scale` must match the symbol master
- `ticks` must not be empty
- `batch_seq` must be non-negative
- `time` must be RFC3339Nano or the accepted legacy timestamp layout
- when `day` is provided, every tick must fall within that UTC day

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

## OpenAPI And Swagger UI Operation Policy

### Source Of Truth

- OpenAPI source of truth is implementation code under `internal/interface/http/` and `cmd/api/main.go` annotation metadata.
- Generated artifacts under `openapi/` are not hand-edited.
- Contract changes are reviewed as both source diff and generated artifact diff.

### Generation And Verification

- Generate OpenAPI: `make openapi`
- Verify contract drift: `make contracts-check`
- API changes must regenerate `implementations/dag-core/openapi/` in the same change set.

### Swagger UI Serving Policy (`echo-swagger`)

- Canonical route: `/swagger/*`
- Responsibility split:
  - `swag`: OpenAPI artifact generation
  - `echo-swagger`: runtime docs/UI serving
- Runtime config:
  - `SWAGGER_UI_ENABLED`: optional bool override (`true`/`false`)
  - `SWAGGER_UI_ROUTE`: route pattern (default `/swagger/*`)
- Default environment policy:
  - `dev`: enabled by default
  - `staging`: enabled only with access control
  - `prod`: disabled by default; temporary enablement requires explicit approval and access control

### Review And Local Flow

- Recommended local flow after API changes:
  1. update handler/DTO annotations and related implementation code
  2. run `make openapi`
  3. run `make go-test` and `make contracts-check`
  4. if docs serving is enabled in the target environment, verify `/swagger/*`
- PR review checklist:
  - source-level API intent is clear
  - generated OpenAPI artifacts are updated and reviewed
  - compatibility risk (breaking/non-breaking) is called out in PR description

### Deferred Item

- OpenAPI v2 to v3 migration remains a separate future decision and is not coupled to `echo-swagger` introduction.

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
