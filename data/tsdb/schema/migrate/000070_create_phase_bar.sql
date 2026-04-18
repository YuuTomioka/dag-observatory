CREATE TABLE IF NOT EXISTS phase_bar (
  symbol_id BIGINT NOT NULL REFERENCES symbol(id) ON DELETE RESTRICT,
  phase_id TEXT NOT NULL,
  market TEXT NOT NULL,
  timezone TEXT NOT NULL,
  open_time TIMESTAMPTZ NOT NULL,
  close_time TIMESTAMPTZ NOT NULL,
  open BIGINT NOT NULL,
  high BIGINT NOT NULL,
  high_time TIMESTAMPTZ NOT NULL,
  low BIGINT NOT NULL,
  low_time TIMESTAMPTZ NOT NULL,
  close BIGINT NOT NULL,
  volume BIGINT NOT NULL,
  source TEXT NOT NULL DEFAULT 'tick_bid_ask_mid',
  created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
  updated_at TIMESTAMPTZ NOT NULL DEFAULT now(),

  PRIMARY KEY (symbol_id, phase_id, open_time),

  CHECK (open_time < close_time),
  CHECK (high >= low),
  CHECK (high >= open),
  CHECK (high >= close),
  CHECK (low <= open),
  CHECK (low <= close),
  CHECK (volume >= 0)
);

SELECT create_hypertable('phase_bar', 'open_time', if_not_exists => TRUE);

CREATE INDEX IF NOT EXISTS idx_phase_bar_symbol_phase_open_desc
  ON phase_bar (symbol_id, phase_id, open_time DESC);

CREATE INDEX IF NOT EXISTS idx_phase_bar_phase_open_desc
  ON phase_bar (phase_id, open_time DESC);
