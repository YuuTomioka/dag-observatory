CREATE TABLE IF NOT EXISTS symbol (
  id            BIGSERIAL PRIMARY KEY,
  code          TEXT NOT NULL UNIQUE,
  base          TEXT NOT NULL,
  quote         TEXT NOT NULL,

  -- Price = Raw / 10^PriceScale
  price_scale   SMALLINT NOT NULL CHECK (price_scale >= 0 AND price_scale <= 18),

  -- 1Tick が PriceRaw 何カウントか
  tick_size_raw BIGINT NOT NULL CHECK (tick_size_raw > 0),

  -- 1Pip が PriceRaw 何カウントか
  pip_size_raw  BIGINT NOT NULL CHECK (pip_size_raw > 0)
);

CREATE INDEX IF NOT EXISTS idx_symbol_base_quote ON symbol (base, quote);
