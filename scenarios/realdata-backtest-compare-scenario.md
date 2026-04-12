# Real-Data Backtest Compare Scenario

## Purpose

This document defines a reproducible validation flow for:

- running a period backtest with real timeframe-bar data
- comparing two backtest runs
- inspecting run and node-level differences from the minimum UI

For the promoted integrated flow including normal/skip/fail/retry and JSONL replay checks, use:

- `scenarios/run-inspection-backtest-integrated-scenario.md`

## Preconditions

- local stack is up (`make dev-up`)
- TSDB migrations are applied (`data/tsdb/README.md`)
- `dag-core-api` is running with marketdata repositories configured
- historical timeframe bars exist for the target symbol/timeframe/range

### Minimum Fixture Setup

Apply seed SQL (includes `USDJPY` symbol and `2026-04-01` M1 bars):

```bash
cd /home/user/shiq/dag-observatory
make tsdb-seed
```

Optional gap fixture for `strict` / `skip` behavior checks:

```sql
DELETE FROM timeframe_bar
WHERE symbol_id = (SELECT id FROM symbol WHERE code = 'USDJPY')
  AND timeframe_code = 'm1'
  AND open_time >= '2026-04-01T02:00:00Z'::timestamptz
  AND open_time <  '2026-04-01T02:10:00Z'::timestamptz;
```

## Input Metadata Rule

Each backtest request should include reproducibility metadata in addition to range and symbol:

- `workflow_name`
- `workflow_version`
- `parameter_set_id`
- `source`
- `timezone`
- `gap_handling`

These values are returned by the backtest API response and should be preserved in test notes.

## Validation Flow

### Step 1. Run base backtest

Call:

- `POST /algotrade/backtests:run`

Required fields:

- `partition`
- `symbol_code`
- `timeframe_code`
- `from`
- `to`

Recommended metadata:

- `run_id`
- `workflow_name`
- `workflow_version`
- `parameter_set_id`
- `source`
- `timezone`
- `gap_handling`
- `spread_bps`
- `fee_bps`
- `slippage_bps`
- `min_lot`

Expected outcome:

- HTTP `200` or `202`
- response includes base `run_id`
- response includes `cycle_count` and `window_size_bars`

### Step 2. Run target backtest with changed assumptions

Call the same endpoint with either:

- different execution assumptions (`fee_bps`, `slippage_bps`, `spread_bps`, `min_lot`)
- different parameter metadata (`parameter_set_id` or workflow version)

Expected outcome:

- second run has a different `run_id`
- both runs are queryable through run read APIs

### Step 3. Compare runs

Call:

- `GET /runs/compare?base_run_id=<base>&target_run_id=<target>`

Expected checks:

- `summary_delta` is present
- `state_field_deltas` is present
- changed fields are visible for at least one comparison target

### Step 4. Inspect in minimum UI

Open:

- `GET /runs/ui?run_id=<base>&target_run_id=<target>`

Expected checks:

- run timeline remains available
- compare target selector is populated
- detail pane shows `Run Compare` section with delta fields

### Step 5. Verify summary and steps directly

Call:

- `GET /runs/<run_id>/steps`
- `GET /runs/<run_id>/steps/<sequence_no>`
- `GET /algotrade/summary?partition=<partition>`

Expected checks:

- step order and statuses are visible
- summary remains coherent with compare output direction

### Step 6. Verify success/failure/skip/retry signals

Run at least two compare pairs that include:

- one mostly-success run pair
- one pair where target includes failure/skip behavior
- one pair where retry count differs

Concrete API checks:

- `GET /runs/compare?...` contains:
  - `summary_delta.failed_step_delta`
  - `summary_delta.skipped_step_delta`
  - `summary_delta.succeeded_step_delta`
  - `summary_delta.retry_count_delta`
- when assumptions/parameter set differ, `state_field_deltas` includes changed summary fields such as:
  - `strategy_summary.trade_count`
  - `strategy_summary.total_net_pnl`

## Suggested API Commands

```bash
# base
curl -s -X POST http://localhost:8080/algotrade/backtests:run \
  -H 'Content-Type: application/json' \
  -d '{
    "partition":"bt:USDJPY:m1:v1",
    "run_id":"bt-base-001",
    "symbol_code":"USDJPY",
    "timeframe_code":"m1",
    "from":"2026-04-01T00:00:00Z",
    "to":"2026-04-01T06:00:00Z",
    "workflow_name":"breakout_long",
    "workflow_version":"v1",
    "parameter_set_id":"pset-baseline",
    "source":"tsdb.timeframe_bars.v1",
    "timezone":"UTC",
    "gap_handling":"strict",
    "spread_bps":2.0,
    "fee_bps":0.5,
    "slippage_bps":0.8,
    "min_lot":0.01
  }' | jq

# target
curl -s -X POST http://localhost:8080/algotrade/backtests:run \
  -H 'Content-Type: application/json' \
  -d '{
    "partition":"bt:USDJPY:m1:v1",
    "run_id":"bt-target-001",
    "symbol_code":"USDJPY",
    "timeframe_code":"m1",
    "from":"2026-04-01T00:00:00Z",
    "to":"2026-04-01T06:00:00Z",
    "workflow_name":"breakout_long",
    "workflow_version":"v1",
    "parameter_set_id":"pset-higher-cost",
    "source":"tsdb.timeframe_bars.v1",
    "timezone":"UTC",
    "gap_handling":"strict",
    "spread_bps":2.0,
    "fee_bps":1.2,
    "slippage_bps":1.6,
    "min_lot":0.01
  }' | jq

# compare
curl -s "http://localhost:8080/runs/compare?base_run_id=bt-base-001&target_run_id=bt-target-001" | jq
```
