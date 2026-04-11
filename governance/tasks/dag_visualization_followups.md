# DAG Visualization Followups

## Background

The note at `.wrk/202604/20260406-dag-visualization-next-themes.md` reorganized the next visualization theme for `implementations/dag-core`.

The core requirement is not visual polish.
The repository needs an observation path that can explain, for one input event, which nodes ran, in what order, with what state changes, and why the final decision was produced.

Current workflow and test coverage already reach entry decision, execution, and result reflection responsibilities.
However, the public path for tracing submit -> fill -> result reflection as one observable run is still weak.

This task document condenses that working note into a near-term implementation backlog.

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

- [ ] start with a two-pane UI composed of execution timeline and node detail plus diff
- [ ] avoid investing in a full DAG canvas before the run read model proves sufficient
- [ ] keep graph rendering, path highlighting, and broader visual polish as later work

### Phase 8: Fix Correlation IDs

- [ ] fix correlation semantics for `run_id`, `partition`, `intent_id`, `execution_id`, `trade_id`, and `sequence_no`
- [ ] verify these identifiers are enough to answer why a signal or order did or did not happen
- [ ] ensure the same correlation model can support backtest/live comparison later

### Phase 9: Choose The Initial Storage Format

- [ ] start with JSON Lines for execution observation records
- [ ] keep the first persistence path append-friendly and replay-friendly
- [ ] defer DB-backed storage until the observation model and read APIs stabilize

## Acceptance Criteria

- contributors can find a single task document that explains the intended order of visualization follow-up work
- the backlog makes clear that execution observability comes before UI work
- the first implementation slice is small and concrete enough to start from `NodeExecutionEvent` and runner instrumentation
- the task document is suitable as temporary planning context without being mistaken for durable design truth

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
