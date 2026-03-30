-- name: BulkUpsertTimeframeBars :exec
INSERT INTO timeframe_bar (
  symbol_id,
  timeframe_code,
  open_time,
  close_time,
  open,
  high,
  low,
  close,
  volume,
  source
)
SELECT
  unnest(@symbol_ids::bigint[]),
  unnest(@timeframe_codes::text[]),
  unnest(@open_times::timestamptz[]),
  unnest(@close_times::timestamptz[]),
  unnest(@opens::bigint[]),
  unnest(@highs::bigint[]),
  unnest(@lows::bigint[]),
  unnest(@closes::bigint[]),
  unnest(@volumes::bigint[]),
  unnest(@sources::text[])
ON CONFLICT (symbol_id, timeframe_code, open_time)
DO UPDATE SET
  close_time = EXCLUDED.close_time,
  open = EXCLUDED.open,
  high = EXCLUDED.high,
  low = EXCLUDED.low,
  close = EXCLUDED.close,
  volume = EXCLUDED.volume,
  source = EXCLUDED.source,
  updated_at = now();

-- name: DeleteTimeframeBarsBySymbolTimeframeAndRange :exec
DELETE FROM timeframe_bar
WHERE symbol_id = @symbol_id
  AND timeframe_code = @timeframe_code
  AND open_time >= @from_time::timestamptz
  AND open_time <  @to_time::timestamptz;

-- name: GetLatestTimeframeBarBySymbolAndTimeframe :one
SELECT symbol_id, timeframe_code, open_time, close_time, open, high, low, close, volume, source, created_at, updated_at
FROM timeframe_bar
WHERE symbol_id = @symbol_id
  AND timeframe_code = @timeframe_code
ORDER BY open_time DESC
LIMIT 1;

-- name: ListTimeframeBarsBySymbolTimeframeAndRange :many
SELECT symbol_id, timeframe_code, open_time, close_time, open, high, low, close, volume, source, created_at, updated_at
FROM timeframe_bar
WHERE symbol_id = @symbol_id
  AND timeframe_code = @timeframe_code
  AND open_time >= @from_time::timestamptz
  AND open_time <  @to_time::timestamptz
ORDER BY open_time ASC;
