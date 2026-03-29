-- name: InsertTick :exec
INSERT INTO tick (symbol_id, time, bid, ask)
VALUES ($1, $2, $3, $4);

-- name: UpsertTick :exec
INSERT INTO tick (symbol_id, time, bid, ask)
VALUES ($1, $2, $3, $4)
ON CONFLICT (symbol_id, time)
DO UPDATE SET bid = EXCLUDED.bid, ask = EXCLUDED.ask;

-- name: BulkUpsertTicks :exec
INSERT INTO tick (symbol_id, time, bid, ask)
SELECT
    unnest(@symbol_ids::bigint[]),
    unnest(@times::timestamptz[]),
    unnest(@bids::bigint[]),
    unnest(@asks::bigint[])
ON CONFLICT (symbol_id, time)
DO UPDATE SET
    bid = EXCLUDED.bid,
    ask = EXCLUDED.ask;

-- name: GetLatestTickBySymbol :one
SELECT symbol_id, time, bid, ask
FROM tick
WHERE symbol_id = $1
ORDER BY time DESC
LIMIT 1;

-- name: ListTicksBySymbolAndRange :many
SELECT symbol_id, time, bid, ask
FROM tick
WHERE symbol_id = @symbol_id
  AND time >= @from_time::timestamptz
  AND time <  @to_time::timestamptz
ORDER BY time ASC;

-- name: ListTicksBySymbolsAndRange :many
SELECT symbol_id, time, bid, ask
FROM tick
WHERE symbol_id = ANY(@symbol_ids::bigint[])
  AND time >= @from_time::timestamptz
  AND time <  @to_time::timestamptz
ORDER BY time ASC, symbol_id ASC;

-- name: ListTicksByRange :many
SELECT symbol_id, time, bid, ask
FROM tick
WHERE time >= @from_time::timestamptz
  AND time <  @to_time::timestamptz
ORDER BY time ASC, symbol_id ASC;
