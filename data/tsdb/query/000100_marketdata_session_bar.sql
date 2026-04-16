-- name: BulkUpsertSessionBars :exec
INSERT INTO session_bar (
  symbol_id,
  session_code,
  session_date,
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
  unnest(@session_codes::text[]),
  unnest(@session_dates::date[]),
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
ON CONFLICT (symbol_id, session_code, session_date)
DO UPDATE SET
  open_time = EXCLUDED.open_time,
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

-- name: DeleteSessionBarsBySymbolSessionAndDateRange :exec
DELETE FROM session_bar
WHERE symbol_id = @symbol_id
  AND session_code = @session_code
  AND session_date >= @from_date::date
  AND session_date <  @to_date::date;

-- name: GetLatestSessionBarBySymbolAndSession :one
SELECT symbol_id, session_code, session_date, open_time, close_time, open, high, high_time, low, low_time, close, volume, source, created_at, updated_at
FROM session_bar
WHERE symbol_id = @symbol_id
  AND session_code = @session_code
ORDER BY session_date DESC
LIMIT 1;

-- name: ListSessionBarsBySymbolSessionAndDateRange :many
SELECT symbol_id, session_code, session_date, open_time, close_time, open, high, high_time, low, low_time, close, volume, source, created_at, updated_at
FROM session_bar
WHERE symbol_id = @symbol_id
  AND session_code = @session_code
  AND session_date >= @from_date::date
  AND session_date <  @to_date::date
ORDER BY session_date ASC;
