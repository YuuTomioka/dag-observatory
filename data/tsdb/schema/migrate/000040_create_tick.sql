CREATE TABLE IF NOT EXISTS tick (
  symbol_id BIGINT NOT NULL REFERENCES symbol(id) ON DELETE RESTRICT,
  time      TIMESTAMPTZ NOT NULL,
  bid       BIGINT NOT NULL,
  ask       BIGINT NOT NULL,
 
  PRIMARY KEY (symbol_id, time)
);

-- hypertable化（既にhypertableなら何もしない）
SELECT create_hypertable('tick', 'time', if_not_exists => TRUE);

-- 典型クエリ（symbol + 時間範囲、latest）用
CREATE INDEX IF NOT EXISTS idx_tick_symbol_time_desc ON tick (symbol_id, time DESC);

-- （任意）最近期間のスキャンを軽くするならtime単独も
CREATE INDEX IF NOT EXISTS idx_tick_time_desc ON tick (time DESC);
