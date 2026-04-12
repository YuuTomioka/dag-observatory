CREATE TABLE IF NOT EXISTS backtest_run_summaries (
  run_id TEXT PRIMARY KEY,
  partition_key TEXT NOT NULL,
  trade_count BIGINT NOT NULL DEFAULT 0,
  win_rate DOUBLE PRECISION NOT NULL DEFAULT 0,
  total_net_pnl DOUBLE PRECISION NOT NULL DEFAULT 0,
  max_drawdown DOUBLE PRECISION NOT NULL DEFAULT 0,
  equity_point_count BIGINT NOT NULL DEFAULT 0,
  has_strategy_fields BOOLEAN NOT NULL DEFAULT FALSE,
  created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
  updated_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE INDEX IF NOT EXISTS idx_backtest_run_summaries_partition_updated
  ON backtest_run_summaries (partition_key, updated_at DESC);
