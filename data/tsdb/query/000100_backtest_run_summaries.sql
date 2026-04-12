-- name: UpsertBacktestRunSummary :exec
INSERT INTO backtest_run_summaries (
  run_id,
  partition_key,
  trade_count,
  win_rate,
  total_net_pnl,
  max_drawdown,
  equity_point_count,
  has_strategy_fields
) VALUES (
  @run_id,
  @partition_key,
  @trade_count,
  @win_rate,
  @total_net_pnl,
  @max_drawdown,
  @equity_point_count,
  @has_strategy_fields
)
ON CONFLICT (run_id)
DO UPDATE SET
  partition_key = EXCLUDED.partition_key,
  trade_count = EXCLUDED.trade_count,
  win_rate = EXCLUDED.win_rate,
  total_net_pnl = EXCLUDED.total_net_pnl,
  max_drawdown = EXCLUDED.max_drawdown,
  equity_point_count = EXCLUDED.equity_point_count,
  has_strategy_fields = EXCLUDED.has_strategy_fields,
  updated_at = now();

-- name: GetBacktestRunSummaryByRunID :one
SELECT
  run_id,
  partition_key,
  trade_count,
  win_rate,
  total_net_pnl,
  max_drawdown,
  equity_point_count,
  has_strategy_fields,
  created_at,
  updated_at
FROM backtest_run_summaries
WHERE run_id = @run_id;
