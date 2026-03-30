CREATE TABLE IF NOT EXISTS timeframe_bar (
  symbol_id BIGINT NOT NULL REFERENCES symbol(id) ON DELETE RESTRICT,
  timeframe_code TEXT NOT NULL,
  open_time TIMESTAMPTZ NOT NULL,
  close_time TIMESTAMPTZ NOT NULL,
  open BIGINT NOT NULL,
  high BIGINT NOT NULL,
  low BIGINT NOT NULL,
  close BIGINT NOT NULL,
  volume BIGINT NOT NULL,
  source TEXT NOT NULL DEFAULT 'tick_mid',
  created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
  updated_at TIMESTAMPTZ NOT NULL DEFAULT now(),

  PRIMARY KEY (symbol_id, timeframe_code, open_time),

  CHECK (open_time < close_time),
  CHECK (high >= low),
  CHECK (high >= open),
  CHECK (high >= close),
  CHECK (low <= open),
  CHECK (low <= close),
  CHECK (volume >= 0)
);

SELECT create_hypertable('timeframe_bar', 'open_time', if_not_exists => TRUE);

CREATE INDEX IF NOT EXISTS idx_timeframe_bar_symbol_tf_open_desc
  ON timeframe_bar (symbol_id, timeframe_code, open_time DESC);

CREATE INDEX IF NOT EXISTS idx_timeframe_bar_tf_open_desc
  ON timeframe_bar (timeframe_code, open_time DESC);
