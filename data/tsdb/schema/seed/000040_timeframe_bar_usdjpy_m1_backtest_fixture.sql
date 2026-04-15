-- USDJPY M1 fixture for real-data backtest manual validation.
-- Range: 2026-04-01T00:00:00Z .. 2026-04-01T05:59:00Z (360 bars)
-- Replay-safe via ON CONFLICT update.

WITH target_symbol AS (
  SELECT id AS symbol_id
  FROM symbol
  WHERE code = 'USDJPY'
  LIMIT 1
),
bars AS (
  SELECT
    s.symbol_id,
    'm1'::text AS timeframe_code,
    gs.open_time,
    gs.open_time + INTERVAL '1 minute' AS close_time,
    -- deterministic but non-flat synthetic path
    (145000 + (gs.idx * 2) + CASE WHEN (gs.idx % 20) < 10 THEN (gs.idx % 10) ELSE (10 - (gs.idx % 10)) END)::bigint AS open_raw,
    (145000 + (gs.idx * 2) + CASE WHEN (gs.idx % 20) < 10 THEN (gs.idx % 10) ELSE (10 - (gs.idx % 10)) END
      + CASE WHEN (gs.idx % 2) = 0 THEN 1 ELSE -1 END)::bigint AS close_raw,
    (100 + (gs.idx % 50))::bigint AS volume_raw
  FROM target_symbol s
  CROSS JOIN LATERAL (
    SELECT
      ts AS open_time,
      (ROW_NUMBER() OVER (ORDER BY ts) - 1)::bigint AS idx
    FROM generate_series(
      '2026-04-01T00:00:00Z'::timestamptz,
      '2026-04-01T05:59:00Z'::timestamptz,
      '1 minute'::interval
    ) AS ts
  ) gs
)
INSERT INTO timeframe_bar (
  symbol_id,
  timeframe_code,
  open_time,
  close_time,
  open,
  high,
  high_time,
  low,
  low_time,
  close,
  volume,
  source
)
SELECT
  symbol_id,
  timeframe_code,
  open_time,
  close_time,
  open_raw,
  GREATEST(open_raw, close_raw) + 3,
  open_time + INTERVAL '30 seconds',
  LEAST(open_raw, close_raw) - 3,
  open_time,
  close_raw,
  volume_raw,
  'manual.backtest.fixture.v1'
FROM bars
ON CONFLICT (symbol_id, timeframe_code, open_time)
DO UPDATE SET
  close_time = EXCLUDED.close_time,
  open = EXCLUDED.open,
  high = EXCLUDED.high,
  high_time = EXCLUDED.high_time,
  low = EXCLUDED.low,
  low_time = EXCLUDED.low_time,
  close = EXCLUDED.close,
  volume = EXCLUDED.volume,
  source = EXCLUDED.source,
  updated_at = now();
