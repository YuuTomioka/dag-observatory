# Real-Data Backtest And Trade Visualization Followups

## Background

The repository already has run-read APIs, summary/trade read paths, and a minimum run inspector UI.

To make "real-data backtest + trade-result visualization" practical, we still need a dedicated execution path for period backtests, stable comparison outputs, and scenario-level reproducibility.

This document captures that implementation backlog as temporary task context.

Primary prerequisite track:
`governance/tasks/dag_visualization_followups.md` Phase 1-12 and Phase 14.
This backlog starts after that foundation is available or in parallel only when interfaces are already stable.

## In Scope

- defining and implementing a real-data backtest execution path in `implementations/dag-core`
- adding result models and APIs needed for run-to-run comparison
- extending the current run inspector path to cover comparison-oriented visualization
- adding scenario-level validation steps for reproducible checks

## Out Of Scope

- changing repository-wide structure without promoting rules to `blueprint/`
- product-specific dashboard polish or broad BI requirements
- replacing existing runtime/state truth with UI-owned truth
- committing to a final persistent storage design beyond the minimum needed for this backlog

## Checklist

### Prerequisite Gate

- [x] confirm `dag_visualization_followups.md` Phase 1-12 are complete or interface-stable
- [x] confirm correlation hardening scope (Phase 14) is compatible with planned backtest identifiers

Gate note (2026-04-12):

- `dag_visualization_followups.md` Phase 1-12 are now complete in this repository state.
- Backtest identifier scheme (`run_id`, cycle suffix, `partition`, and step `sequence_no`) is compatible with current correlation model; full cross-surface hardening remains tracked under Phase 14.

### Phase 1: Structural Clarification

- [x] clarify in `blueprint/architecture/` that period backtest belongs to analysis flow and stays explicit from runtime execution flow
- [x] clarify subsystem ownership for backtest execution, comparison read model, and visualization read APIs

### Phase 2: Backtest Input Contract

- [x] define a dedicated backtest input contract (symbol, timeframe, from/to, source, timezone, gap handling)
- [x] expose the contract at HTTP boundary and OpenAPI artifacts
- [x] add validation tests for missing/invalid range and inconsistent input combinations

### Phase 3: Backtest Run Entrypoint

- [x] add a dedicated entrypoint such as `POST /algotrade/backtests:run`
- [x] return stable correlation keys (`run_id`, `partition`, mode/status) for later read/visual paths
- [x] support async-friendly handling consistent with current run-read APIs

### Phase 4: Real-Data Execution Path

- [x] add workflow/node path to load timeframe bars from TSDB by bounded windows
- [x] implement period-loop orchestration for multi-cycle execution over historical range
- [x] keep deterministic replay assumptions explicit (state initialization, ordering, retry behavior)

### Phase 5: Execution Realism Controls

- [x] add explicit inputs for spread/slippage/fee/min-lot constraints
- [x] ensure applied execution assumptions are persisted or emitted for run auditability
- [x] add tests covering pnl sensitivity to execution-assumption changes

### Phase 6: Backtest Result Model

- [x] define minimum run-level result projection (trade_count, win_rate, net pnl, max drawdown, equity series summary)
- [x] persist or reconstruct this projection through repository-approved SQL/query assets
- [x] keep SQL assets as source of truth when durable schema/query changes are needed

### Phase 7: Compare API

- [x] add compare endpoint such as `GET /runs/compare?base_run_id=...&target_run_id=...`
- [x] include summary-level deltas and key field-level divergences
- [x] add tests for pnl divergence, trade-count divergence, and skip/fail path differences

### Phase 8: Visualization API Extension

- [x] extend `/algotrade/trades` for run-scoped and range-scoped queries with paging/filter support
- [x] add API(s) for equity/drawdown time-series visualization data
- [x] align response shape with current minimum UI and avoid UI-owned data truth

### Phase 9: UI Compare View

- [x] extend `/runs/ui` with side-by-side run comparison using read APIs only
- [x] show summary delta and step-level differences without introducing a full DAG canvas requirement
- [x] add minimum integration checks for selecting two runs and inspecting diff highlights

### Phase 10: Reproducibility And Scenario Coverage

- [x] record reproducibility metadata (workflow version, parameter set, data range, source identifiers)
- [x] add scenario guidance under `scenarios/` for real-data backtest -> compare -> visualize flow
- [x] verify success/failure/skip/retry cases are reproducible with concrete API checks

### MVP Slice (Recommended First)

- [x] `POST /algotrade/backtests:run`
- [x] real-data period-loop execution path
- [x] `GET /runs/compare`
- [x] `/runs/ui` compare mode

## Acceptance Criteria

- contributors can follow one checklist to implement real-data backtest and trade-result visualization incrementally
- the backlog is ordered so execution/read-model fundamentals come before UI expansion
- MVP scope is explicit and small enough to start implementation immediately
- promotion targets are clear when decisions become durable (`blueprint/`, `scenarios/`, `data/`)

## Result

- added this task document as temporary planning context for real-data backtest and visualization follow-up work
- execution order is explicitly coordinated with `governance/tasks/dag_visualization_followups.md`
- integrated scenario guidance is promoted at:
  - `scenarios/run-inspection-backtest-integrated-scenario.md`
