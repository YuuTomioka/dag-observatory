# Run Inspection And Backtest Integrated Scenario

## Purpose

This document unifies:

- run inspection (normal/skip/fail/retry)
- real-data backtest compare
- minimum UI verification
- JSONL replay verification

It is the promoted reproducible flow for `implementations/dag-core` run observability.

## Preconditions

- local stack is up (`make dev-up`)
- `dag-core-api` is running
- timeframe bars exist for the test range
- optional JSONL path is enabled when replay checks are needed:
  - `DAGRUNTIME_OBSERVATION_JSONL_PATH=<path>`

### Minimum Data Setup

```bash
cd /home/user/shiq/dag-observatory
make tsdb-seed
```

The seed set includes:

- `USDJPY` symbol master row
- `m1` timeframe bars for `2026-04-01T00:00:00Z` to `2026-04-01T05:59:00Z`

## Flow

### 1. Create two comparable backtest runs

Run:

- `POST /algotrade/backtests:run` (base)
- `POST /algotrade/backtests:run` (target, changed cost/parameter assumptions)

Capture:

- `run_id`
- `partition`
- `workflow_name`, `workflow_version`, `parameter_set_id`
- `source`, `timezone`, `gap_handling`

### 2. Verify run list/detail/step path

Check:

- `GET /runs?partition=<partition>&limit=200`
- `GET /runs/<run_id>`
- `GET /runs/<run_id>/steps`
- `GET /runs/<run_id>/steps/<sequence_no>`

Validation points:

- step order uses `sequence_no`
- node detail includes correlation IDs:
  - `partition`, `intent_id`, `execution_id`, `trade_id`

### 3. Verify compare and visualization APIs

Check:

- `GET /runs/compare?base_run_id=<base>&target_run_id=<target>`
- `GET /algotrade/trades?run_id=<base>&limit=100&offset=0`
- `GET /algotrade/equity?run_id=<base>`
- `GET /runs/<run_id>/backtest-summary`

Validation points:

- compare has summary deltas and changed fields
- trades/equity are run-scoped and reproducible from the same partition state

### 4. Verify UI path with bounded fetch

Open:

- `/runs/ui?partition=<partition>&run_id=<base>&target_run_id=<target>`

Validation points:

- compare selector and diff summary are visible
- UI uses bounded run fetch (`limit`) and bounded step rendering

### 5. Verify normal/skip/fail/retry coverage

Confirm at least one pair/set where:

- normal: target mostly succeeds
- skip: skip reason appears in steps/compare deltas
- fail: failed steps are reflected
- retry: retry count differs

Concrete fields:

- `summary_delta.failed_step_delta`
- `summary_delta.skipped_step_delta`
- `summary_delta.retry_count_delta`

### 6. Verify JSONL replay (when enabled)

1. stop API after runs are recorded
2. restart API with same `DAGRUNTIME_OBSERVATION_JSONL_PATH`
3. call:
  - `GET /runs?partition=<partition>`
  - `GET /healthz`

Validation points:

- previously recorded runs/steps are readable after restart
- `/healthz.observation_replay` exposes counters:
  - `skipped_lines`
  - `errors`

## Suggested Commands

```bash
curl -s "http://localhost:8080/runs?partition=bt:USDJPY:m1:v1&limit=200" | jq
curl -s "http://localhost:8080/runs/compare?base_run_id=bt-base-001&target_run_id=bt-target-001" | jq
curl -s "http://localhost:8080/algotrade/trades?run_id=bt-base-001&limit=100&offset=0" | jq
curl -s "http://localhost:8080/algotrade/equity?run_id=bt-base-001" | jq
curl -s "http://localhost:8080/healthz" | jq
```
