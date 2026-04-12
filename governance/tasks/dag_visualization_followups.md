# DAG Visualization Followups

## Background

The note at `.wrk/202604/20260406-dag-visualization-next-themes.md` reorganized the next visualization theme for `implementations/dag-core`.

The core requirement is not visual polish.
The repository needs an observation path that can explain, for one input event, which nodes ran, in what order, with what state changes, and why the final decision was produced.

Current workflow and test coverage already reach entry decision, execution, and result reflection responsibilities.
However, the public path for tracing submit -> fill -> result reflection as one observable run is still weak.

This task document condenses that working note into a near-term implementation backlog.

For real-data period backtest and trade-result comparison work, the primary tracker is:
`governance/tasks/backtest_realdata_visualization_followups.md`.
This document keeps the run-observability foundation and read-model stabilization track.

## In Scope

- fixing the observation unit for node execution in `implementations/dag-core`
- adding runner-side hooks that record execution start/end, reason, and state references
- defining the minimum state-diff shape needed for run inspection
- exposing a public or internal entrypoint for result reflection and replay-oriented validation
- adding run-oriented read APIs before any substantial UI work
- adding a compact result-summary API that supports run comparison
- planning the minimum UI only after the read model exists

## Out Of Scope

- durable repository-wide structural rules that belong in `blueprint/` or `blueprint/adr/`
- polishing a graph-heavy DAG UI before execution and state observability exist
- database-first persistence design for visualization records
- broad product-facing dashboard requirements outside the current validation foundation

## Coordination With Real-Data Backtest Tracker

- backtest execution path, compare API expansion, and compare-focused UI work are coordinated with `governance/tasks/backtest_realdata_visualization_followups.md`
- this document remains the primary tracker for run observation model, recorder/read-model durability, and cross-surface correlation hardening
- implementation should preserve the sequence: observability/read-model foundation first, then backtest-and-compare expansion

## Task List

### Phase 1: Fix The Observation Model

- [x] define `NodeExecutionEvent` as the primary execution observation record
- [x] fix the minimum fields for run, node, ordering, trigger/skip reason, duration, snapshot refs, and error
- [x] decide where the temporary event-shape memo lives until promoted into a durable location
- [x] add the minimum Go struct and supporting types in `implementations/dag-core`

Proposed minimum fields:

- `run_id`
- `partition`
- `event_time`
- `sequence_no`
- `node_id`
- `node_name`
- `status`
- `trigger_reason`
- `skip_reason`
- `upstream_node_ids`
- `input_ref`
- `output_ref`
- `snapshot_before_ref`
- `snapshot_after_ref`
- `duration_ns`
- `error`

Decision note (2026-04-11):

- temporary event-shape memo is maintained in this task document under `governance/tasks/`
- minimum Go shape is added at `implementations/dag-core/internal/domain/dagruntime/events/node_execution_event.go`

### Phase 2: Add Runner-Side Recording

- [x] add runner or observer hooks that record `before/after` execution without mixing visualization concerns into node logic
- [x] capture start time, end time, status, and duration for each node execution
- [x] emit trigger reason and skip reason in a normalized shape
- [x] capture state hash or snapshot references before and after execution
- [x] add tests that verify the recorder emits stable ordering and status data

Progress note (2026-04-11):

- runner now emits `NodeExecutionEvent` via recorder with `sequence_no`, `status`, `duration_ns`, and error text
- `snapshot_before_ref` / `snapshot_after_ref` are currently best-effort hash refs from `StateHasher`/`HashableStore`
- `trigger_reason` is normalized to `event_dispatch` or `retry`
- runner supports `SkipError` as a non-fatal node outcome and records normalized `skip_reason`

### Phase 3: Add Minimal Snapshot Diff

- [x] define a compact diff shape that prefers changed fields over full dumps
- [x] implement the first diff targets for `pending_orders`, `open_positions`, `closed_trades`, `daily_pnl`, and `strategy_summary`
- [x] keep the initial output optimized for immediate reading rather than deep generic diff coverage
- [x] add tests for representative diffs such as position changes, signal changes, trade count changes, and pnl changes

Progress note (2026-04-11):

- `NodeExecutionEvent` now includes `state_diff[]` with `{field, before, after}`
- runner emits only changed fields and prioritizes compact summaries such as:
  - `pending_orders.count`
  - `open_positions.count`
  - `closed_trades.count`
  - `daily_pnl.realized_pnl`, `daily_pnl.unrealized_pnl`, `daily_pnl.loss_limit_hit`
  - `strategy_summary.trade_count`, `strategy_summary.total_net_pnl`, `strategy_summary.win_rate`

### Phase 4: Expose Result Reflection Path

- [x] add one external or internal entrypoint that can drive result reflection from outside the workflow tests
- [x] support explicit `partition` selection for manual verification and replay-oriented execution
- [x] decide whether the first entrypoint should be HTTP, debug API, or replay endpoint based on the smallest implementation surface
- [x] verify submit -> fill -> result reflection can be exercised as one traceable flow

Progress note (2026-04-11):

- added HTTP entrypoint `POST /algotrade/result-reflection:run`
- request supports explicit `partition` and `run_id` to drive manual/replay-oriented execution
- `RunWorkflowRequest` now supports explicit partition override so reflection runs can target existing state scope
- added integration scenario test that runs submit workflow -> fill confirm workflow -> result reflection workflow on one partition and verifies resulting closed trade / summary state with recorder-traceable run IDs

### Phase 5: Add Run Read APIs

- [x] add a run listing API such as `/runs`
- [x] add a run detail API such as `/runs/:run_id`
- [x] add a step listing API such as `/runs/:run_id/steps`
- [x] add a node execution detail API such as `/runs/:run_id/nodes/:execution_id`
- [x] prioritize execution order, status, trigger/skip reason, and before/after diff over graph rendering concerns

Progress note (2026-04-12):

- added run read APIs:
  - `GET /runs`
  - `GET /runs/:run_id`
  - `GET /runs/:run_id/steps`
  - `GET /runs/:run_id/nodes/:execution_id`
- introduced in-memory run read model backed by recorder events/cycle results
- run step response exposes execution order (`sequence_no`), `status`, `trigger_reason`, `skip_reason`, and `state_diff`

### Phase 6: Add Result Summary API

- [x] add `GET /algotrade/summary?partition=...` or equivalent
- [x] expose enough summary data to support hypothesis comparison beyond `TradeView`
- [x] verify the summary response can act as the minimum judgment view for run comparison

Progress note (2026-04-12):

- added `GET /algotrade/summary?partition=...`
- response exposes summary projection fields including trade counts, win/loss metrics, and pnl aggregates
- added handler tests for partition validation and summary projection retrieval

### Phase 7: Add The Minimum UI

- [x] start with a two-pane UI composed of execution timeline and node detail plus diff
- [x] avoid investing in a full DAG canvas before the run read model proves sufficient
- [x] keep graph rendering, path highlighting, and broader visual polish as later work

Progress note (2026-04-12):

- added `GET /runs/ui` as a minimal inspector UI served directly from `implementations/dag-core`
- UI stays intentionally thin and reads from existing APIs:
  - `GET /runs`
  - `GET /runs/:run_id/steps`
  - `GET /runs/:run_id/nodes/:execution_id`
  - `GET /algotrade/summary?partition=...`
- left pane focuses on run selection plus execution timeline
- right pane focuses on node detail, refs, and compact state diff
- graph canvas, path highlighting, and broader frontend framework choices remain deferred
- UI now tolerates short read-model lag after enqueue:
  - when `step_count > 0` but `/runs/:run_id/steps` is temporarily empty, the UI keeps polling briefly and auto-hydrates steps/node detail once available
  - step area shows a transitional message (`Waiting for step records to become available...`) instead of a hard terminal empty state

### Phase 8: Fix Correlation IDs

- [x] fix correlation semantics for `run_id`, `partition`, `intent_id`, `execution_id`, `trade_id`, and `sequence_no`
- [x] verify these identifiers are enough to answer why a signal or order did or did not happen
- [x] ensure the same correlation model can support backtest/live comparison later

Progress note (2026-04-12):

- `NodeExecutionEvent` now carries `intent_id`, domain `execution_id`, and `trade_id` separately from `sequence_no`
- canonical node-step read path is now `GET /runs/:run_id/steps/:sequence_no`
- legacy `GET /runs/:run_id/nodes/:execution_id` remains as a deprecated compatibility alias, but it still resolves by sequence number
- runner now records artifact refs for `input_ref` / `output_ref` and extracts correlation IDs from known algotrade artifacts such as `TradeIntent`, `ExecutionResult`, and `ClosedTrade`
- tests verify sequence-based step lookup and correlation propagation through runner-recorded node events
- path-encoded run IDs are now decoded at HTTP handler boundaries (`%3A` etc.) before lookup so UI/network clients can safely call:
  - `GET /runs/:run_id/steps`
  - `GET /runs/:run_id/steps/:sequence_no`
  - `GET /runs/:run_id/nodes/:execution_id`

### Phase 9: Choose The Initial Storage Format

- [x] start with JSON Lines for execution observation records
- [x] keep the first persistence path append-friendly and replay-friendly
- [x] defer DB-backed storage until the observation model and read APIs stabilize

Progress note (2026-04-12):

- added optional JSON Lines observation recorder under `implementations/dag-core/internal/infrastructure/dagruntime/recorder`
- recorder appends `run_event`, `node_execution`, and `cycle_result` records while preserving the existing in-memory run read model
- file output is enabled only when `DAGRUNTIME_OBSERVATION_JSONL_PATH` is set
- JSON Lines path is intentionally append-only and local-file based; DB-backed storage remains deferred until the observation shape and read APIs settle

### Phase 10: Add JSONL Replay Read Path

- [ ] add startup-time replay loading from JSON Lines into the run read model
- [ ] preserve append-first write path while allowing process-restart recovery for run inspection
- [ ] ensure replay logic accepts mixed record kinds (`run_event`, `node_execution`, `cycle_result`) in one stream
- [ ] add tests that verify run APIs can serve previously recorded runs after restart-like reconstruction

### Phase 11: Add JSONL Operational Guards

- [ ] add simple file-rotation policy for long-running local observation output
- [ ] make JSONL reader tolerant to partial or malformed lines without stopping whole API operation
- [ ] expose minimum health counters for skipped lines and replay errors
- [ ] add tests for malformed-line tolerance and continued replay of valid trailing records

### Phase 12: Stabilize Run Read API Query Contract

- [ ] add filtering and paging query surface for `GET /runs` (for example `partition`, `status`, `since`, `until`, `limit`, `cursor`)
- [ ] keep response shape compact and backward compatible with current minimum UI usage
- [ ] define ordering and cursor semantics explicitly for deterministic client pagination
- [ ] add handler tests for filter combinations and cursor continuity

### Phase 13: Add Run Comparison API

- [ ] execute this phase under `governance/tasks/backtest_realdata_visualization_followups.md` and keep this document aligned only for run-read compatibility constraints
- [ ] add a compact compare endpoint such as `GET /runs/compare?base_run_id=...&target_run_id=...`
- [ ] include summary-level deltas and per-field state-diff deltas as minimum comparison output
- [ ] ensure comparison output can explain meaningful result differences without requiring UI-first interpretation
- [ ] add tests covering common comparison cases such as pnl divergence, trade-count divergence, and skip/fail path differences

### Phase 14: Strengthen Correlation Across APIs, Logs, And Traces

- [ ] align `run_id`, `partition`, `intent_id`, `execution_id`, `trade_id`, and `sequence_no` across read APIs and intent logs
- [ ] ensure trace attributes include identifiers needed to jump between trace view and run-step view
- [ ] verify high-cardinality identifiers remain fields/attributes rather than promoted labels
- [ ] add focused observability tests or checks for correlation-field presence in representative flows

### Phase 15: Extend Minimum UI With Compare View

- [ ] execute this phase under `governance/tasks/backtest_realdata_visualization_followups.md` and keep this document aligned only for base run-inspector compatibility
- [ ] keep the current two-pane shape and add side-by-side run comparison without introducing full DAG canvas work
- [ ] drive compare view only from existing and new read APIs rather than adding UI-owned state truth
- [ ] keep performance acceptable for near-term run volumes through compact rendering and bounded fetch size
- [ ] add minimum UI integration checks for selecting two runs and inspecting diff highlights

### Phase 16: Promote Run-Inspection Scenario Coverage

- [ ] coordinate this phase with `governance/tasks/backtest_realdata_visualization_followups.md` so run-inspection and backtest-compare scenario guidance converge in one scenario flow
- [ ] add scenario-level verification guidance under `scenarios/` for normal, skip, fail, and retry-oriented run inspection
- [ ] align scenario steps with current API/UI/JSONL verification flow so contributors can reproduce checks consistently
- [ ] keep scenario guidance implementation-agnostic where possible while preserving concrete command examples
- [ ] add references from this task document to the promoted scenario once added

## Acceptance Criteria

- contributors can find a consistent execution order across this task document and `governance/tasks/backtest_realdata_visualization_followups.md`
- the backlog makes clear that execution observability comes before UI work
- the first implementation slice is small and concrete enough to start from `NodeExecutionEvent` and runner instrumentation
- the task document is suitable as temporary planning context without being mistaken for durable design truth

## Post-Implementation Verification Flow

Use this verification order after changes to run read APIs, minimum UI, correlation IDs, or JSON Lines observation output.

1. Run targeted Go tests from `implementations/dag-core`.
2. Start the API with optional JSON Lines output enabled.
3. Drive one traceable run flow.
4. Verify run read APIs directly.
5. Verify the minimum UI against the same run.
6. Verify JSON Lines append output when enabled.

Recommended commands:

```bash
cd /home/user/shiq/dag-observatory/implementations/dag-core
mkdir -p .tmp/go-cache
GOCACHE=$(pwd)/.tmp/go-cache go test ./internal/infrastructure/dagruntime/recorder ./internal/domain/dagruntime/engine ./internal/interface/http/dag/handler ./internal/di
```

```bash
cd /home/user/shiq/dag-observatory/implementations/dag-core
export DAGRUNTIME_OBSERVATION_JSONL_PATH=$(pwd)/tmp/observations/runs.jsonl
go run ./cmd/api
```

Recommended API checks:

```bash
curl -s http://localhost:8080/runs | jq
curl -s http://localhost:8080/runs/<run_id> | jq
curl -s http://localhost:8080/runs/<run_id>/steps | jq
curl -s http://localhost:8080/runs/<run_id>/steps/1 | jq
curl -s http://localhost:8080/runs/<run_id>/nodes/1 | jq
curl -s "http://localhost:8080/algotrade/summary?partition=<partition>" | jq

# optional: validate URL-encoded run_id compatibility (e.g. run IDs containing `:`)
ENCODED_RUN_ID=$(printf '%s' "<run_id>" | jq -sRr @uri)
curl -s "http://localhost:8080/runs/${ENCODED_RUN_ID}/steps" | jq
curl -s "http://localhost:8080/runs/${ENCODED_RUN_ID}/steps/1" | jq
```

Recommended UI check:

- open `http://localhost:8080/runs/ui`
- optionally narrow by partition with `http://localhost:8080/runs/ui?partition=<partition>`
- confirm left pane shows runs and step order
- confirm right pane shows node detail, refs, correlation IDs, and compact diff

Recommended JSON Lines check:

```bash
tail -n 20 /home/user/shiq/dag-observatory/implementations/dag-core/tmp/observations/runs.jsonl
```

Verification points:

- `GET /runs/:run_id/steps/:sequence_no` resolves step detail by `sequence_no`
- compatibility alias `GET /runs/:run_id/nodes/:execution_id` still resolves the same step when passed the sequence number
- path-encoded run IDs (such as `%3A`) resolve identically to decoded run IDs
- node detail exposes `intent_id`, `execution_id`, `trade_id`, `input_ref`, and `output_ref` when available
- state changes appear under `state_diff`
- JSON Lines output appends `run_event`, `node_execution`, and `cycle_result` records without replacing the in-memory read model

## Result

- the working note at `.wrk/202604/20260406-dag-visualization-next-themes.md` was condensed into a repository-aligned task document under `governance/tasks/`
- the recommended implementation order is fixed as:
  1. observation model
  2. runner recording
  3. snapshot diff
  4. result reflection entrypoint
  5. run read APIs
  6. result summary API
  7. minimum UI
  8. correlation IDs
  9. initial storage format
  10. JSONL replay read path
  11. JSONL operational guards
  12. run read API query contract stabilization
  13. handoff to `backtest_realdata_visualization_followups.md` for compare/backtest/expanded scenario track
  14. cross-surface correlation strengthening
