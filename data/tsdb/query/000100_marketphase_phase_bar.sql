-- name: BulkUpsertPhaseBars :exec
INSERT INTO phase_bar (
  symbol_id,
  phase_code,
  phase_date,
  market,
  timezone,
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
  unnest(@symbol_ids::bigint[]),
  unnest(@phase_codes::text[]),
  unnest(@phase_dates::date[]),
  unnest(@markets::text[]),
  unnest(@timezones::text[]),
  unnest(@open_times::timestamptz[]),
  unnest(@close_times::timestamptz[]),
  unnest(@opens::bigint[]),
  unnest(@highs::bigint[]),
  unnest(@high_times::timestamptz[]),
  unnest(@lows::bigint[]),
  unnest(@low_times::timestamptz[]),
  unnest(@closes::bigint[]),
  unnest(@volumes::bigint[]),
  unnest(@sources::text[])
ON CONFLICT (symbol_id, phase_code, phase_date)
DO UPDATE SET
  market = EXCLUDED.market,
  timezone = EXCLUDED.timezone,
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

-- name: DeletePhaseBarsBySymbolPhaseAndRange :exec
DELETE FROM phase_bar
WHERE symbol_id = @symbol_id
  AND phase_code = @phase_code
  AND open_time >= @from_time::timestamptz
  AND open_time <  @to_time::timestamptz;

-- name: GetLatestPhaseBarBySymbolAndPhase :one
SELECT symbol_id, phase_code, phase_date, market, timezone, open_time, close_time, open, high, high_time, low, low_time, close, volume, source, created_at, updated_at
FROM phase_bar
WHERE symbol_id = @symbol_id
  AND phase_code = @phase_code
ORDER BY open_time DESC
LIMIT 1;

-- name: ListPhaseBarsBySymbolPhaseAndRange :many
SELECT symbol_id, phase_code, phase_date, market, timezone, open_time, close_time, open, high, high_time, low, low_time, close, volume, source, created_at, updated_at
FROM phase_bar
WHERE symbol_id = @symbol_id
  AND phase_code = @phase_code
  AND open_time >= @from_time::timestamptz
  AND open_time <  @to_time::timestamptz
ORDER BY open_time ASC;
