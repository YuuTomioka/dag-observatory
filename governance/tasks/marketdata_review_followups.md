# Marketdata 見直しフォローアップ

## Purpose

marketdata まわりの現行仕様と、関連する検証・引き継ぎ観点を一時的な共有作業文脈として整理する。

この文書は恒久仕様ではない。durable な結論は `blueprint/`、`data/tsdb/`、`implementations/dag-core/` の適切な正本へ昇格させる。

## Current Scope

- `OHLCV` と `TimeframeBar`
- `timeframe_bar` の TSDB schema / query / repository
- tick から timeframe bar を backfill する usecase
- HTTP / workflow payload / OpenAPI / sqlc generated code に露出する `timeframe_bar` 入出力

`marketphase`、`marketcontext`、`phase_bar`、`session_bar` は現行実装の対象外。

## Current Decisions

- `OHLCV` は `Opentime`、`Closetime`、`Open`、`High`、`Hightime`、`Low`、`Lowtime`、`Close`、`Volume` を持つ。
- Go field name は既存の `Opentime` / `Closetime` に合わせて `Hightime` / `Lowtime` とする。
- JSON / OpenAPI / SQL の外部名は `high_time` / `low_time` とする。
- `Hightime` / `Lowtime` は last-touch semantics で扱う。
- high / low が同値で更新された場合も、対応時刻を最新 tick の時刻へ更新する。
- tick から `TimeframeBar` を集約する場合、open / close は bid-ask mid、high は ask、low は bid を採用する。
- `TimeframeBar.Source` は混合入力を示す `tick_bid_ask_mid` とする。
- `timeframe_bar` の identity は `(symbol_id, timeframe_code, open_time)` とする。

## Source Of Truth

- Domain OHLC:
  - `implementations/dag-core/internal/domain/marketdata/ohlc/`
- Domain timeframe bar:
  - `implementations/dag-core/internal/domain/marketdata/ohlc/timeframe/`
- TSDB schema:
  - `data/tsdb/schema/migrate/000050_create_timeframe_bar.sql`
- TSDB query:
  - `data/tsdb/query/000090_marketdata_timeframe_bar.sql`
- Application repository contract:
  - `implementations/dag-core/internal/application/marketdata/repository/timeframe_bar.go`
- Backfill usecase:
  - `implementations/dag-core/internal/application/marketdata/usecase/timeframe_bar.go`
- TSDB mapper / repository:
  - `implementations/dag-core/internal/infrastructure/persistence/tsdb/mapper/marketdata/timeframe_bar.go`
  - `implementations/dag-core/internal/infrastructure/persistence/tsdb/repository/marketdata/timeframe_bar.go`

## Operational Notes

- `timeframe_bar` schema changes must start from `data/tsdb/schema/migrate/` and `data/tsdb/query/`.
- SQL changes require `make sqlc-gen`.
- HTTP payload or response shape changes require OpenAPI regeneration.
- Existing HTTP payloads may omit `high_time` / `low_time` where backward-compatible parsing is intentionally supported.
- Existing DB をリセットしない環境では、修正済み migration の再適用ではなく、同等の forward migration または手動 `ALTER TABLE` 方針を別途決める。

## Validation

Use these checks for changes touching current marketdata bar behavior:

- `go test ./...` from `implementations/dag-core`
- `make tsdb-schema-lint`
- `make tsdb-query-lint`
- `make sqlc-gen` when SQL assets changed
- `make openapi` when HTTP request / response shape changed

Focused tests:

- `go test ./internal/domain/marketdata/ohlc/...` from `implementations/dag-core`
- `go test ./internal/application/marketdata/usecase` from `implementations/dag-core`
- `go test ./internal/infrastructure/persistence/tsdb/repository/marketdata` from `implementations/dag-core`

## Out Of Scope

- repository-wide marketdata taxonomy redesign
- market session / phase / context modeling
- `phase_bar` or `session_bar` persistence
- derivative-product-specific indicators, execution logic, or visualization requirements
