-- name: CreateSymbol :one
INSERT INTO symbol (code, base, quote, price_scale, tick_size_raw, pip_size_raw)
VALUES ($1, $2, $3, $4, $5, $6)
RETURNING id, code, base, quote, price_scale, tick_size_raw, pip_size_raw;

-- name: GetSymbolByID :one
SELECT id, code, base, quote, price_scale, tick_size_raw, pip_size_raw
FROM symbol
WHERE id = $1;

-- name: GetSymbolByCode :one
SELECT id, code, base, quote, price_scale, tick_size_raw, pip_size_raw
FROM symbol
WHERE code = $1;

-- name: ListSymbols :many
SELECT id, code, base, quote, price_scale, tick_size_raw, pip_size_raw
FROM symbol
ORDER BY id ASC;
