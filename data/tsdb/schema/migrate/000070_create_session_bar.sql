CREATE TABLE IF NOT EXISTS session_bar (
  symbol_id BIGINT NOT NULL REFERENCES symbol(id) ON DELETE RESTRICT,
  session_code TEXT NOT NULL,
  session_date DATE NOT NULL,
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

  PRIMARY KEY (symbol_id, session_code, session_date),

  CHECK (open_time < close_time),
  CHECK (high >= low),
  CHECK (high >= open),
  CHECK (high >= close),
  CHECK (low <= open),
  CHECK (low <= close),
  CHECK (volume >= 0)
);

CREATE INDEX IF NOT EXISTS idx_session_bar_symbol_session_date_desc
  ON session_bar (symbol_id, session_code, session_date DESC);

CREATE INDEX IF NOT EXISTS idx_session_bar_session_open_desc
  ON session_bar (session_code, open_time DESC);
