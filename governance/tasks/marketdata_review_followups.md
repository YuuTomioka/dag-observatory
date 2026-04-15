# Marketdata 見直しフォローアップ

## Background

marketdata まわりの見直しを進めるため、一時的な共有作業文脈をここに置く。

この文書は恒久仕様ではなく、調査・設計判断・実装順序をまとめるための temporary context として扱う。
durable な結論は `blueprint/`、`data/tsdb/`、`implementations/dag-core/` の適切な正本へ昇格させる。

## In Scope

- `OHLCV` と `TimeframeBar` のフィールド見直し
- `Hightime` / `Lowtime` の意味論と永続化方針の整理
- `Session` / `SessionBar` 導入検討
- `data/tsdb/` 配下の schema / query SQL への影響整理
- HTTP / payload / OpenAPI / sqlc generated code まで含めた実装影響の整理
- durable rule に昇格すべき論点と implementation-local に閉じる論点の切り分け

## Out Of Scope

- marketdata 全体の構造を repository-wide に再定義すること
- 派生先プロダクト固有の指標、執行ロジック、可視化要件の追加
- 一時メモを `blueprint/` や `data/` の正本代わりに扱うこと
- 未合意のまま既存 migration 運用方針を破ること

## Work Tracks

### Track 1: OHLCV / TimeframeBar

#### Status

Implemented in the current working tree.

#### Goal

`OHLCV` に高値・安値の発生時刻を追加し、`timeframe_bar` の永続化対象にも含める。

#### Decisions

- `Hightime` / `Lowtime` は last-touch semantics で扱う
- high/low が同値で更新された場合も、対応時刻を最新 tick の時刻へ更新する
- tick から `TimeframeBar` を集約する場合、open/close は bid/ask mid、high は ask、low は bid を採用する
- `TimeframeBar.Source` は上記の混合入力を示す `tick_bid_ask_mid` とする
- 今回は開発環境リセット前提のため、`data/tsdb/schema/migrate/000050_create_timeframe_bar.sql` を直接修正対象とする

#### Expected Impact

- `implementations/dag-core/internal/domain/marketdata/OHLCV` に `Hightime` / `Lowtime` を追加した
- `timeframe_bar` に `high_time` / `low_time` を追加した
- `data/tsdb/query/000090_marketdata_timeframe_bar.sql` の insert / select / upsert を更新した
- sqlc generated code を再生成した
- TSDB mapper で bulk upsert と row mapping を更新した
- aggregation で high/low 更新時に対応時刻を更新した
- aggregation で high は ask、low は bid、open/close は mid を採用するようにした
- `TimeframeBar.Source` と schema default を `tick_bid_ask_mid` に更新した
- payload parser で `high_time` / `low_time` を受けられるようにした
- OpenAPI generated artifacts を再生成した
- fixture と integration test を更新して round-trip 確認点を追加した

#### Notes

- `OHLCV` は domain 型にとどまらず、HTTP request body、workflow payload、TSDB mapper、OpenAPI generated docs にそのまま露出している
- `timeframe_bar` の durable schema 変更は `data/tsdb/schema/migrate/` と `data/tsdb/query/` の両方を起点に進める必要がある
- sqlc generated code と OpenAPI generated artifacts は SQL / Go 型変更の後に再生成が必要になる
- Go field name は既存の `Opentime` / `Closetime` に合わせて `Hightime` / `Lowtime` のままとし、JSON / OpenAPI / SQL の外部名は `high_time` / `low_time` にそろえた
- `tick_bid_ask_mid` は high/low/open/close の採用価格が単一系列ではないことを明示する source 値として扱う
- 既存 HTTP payload は `high_time` / `low_time` なしでも後方互換で受けられる
- 既存 DB をリセットしない環境では、直接修正した migration ではなく同等の `ALTER TABLE` 適用が別途必要になる

### Track 2: Session / SessionBar

#### Goal

東京タイムのような market session を扱うため、`OHLCV` の派生として `SessionBar` を追加できるか検討する。

domain 側には session の時間帯概念として `Session` を追加する可能性が高い。

#### Open Questions

- `SessionBar` を `OHLCV` の単純埋め込みとして扱うか、`TimeframeBar` と並ぶ別の bar 概念として扱うか
- `Session` を固定 enum 的に持つか、タイムゾーンとセッション境界を持つ値オブジェクトとして扱うか
- 東京、ロンドン、ニューヨークのような代表 session を parent repository の reusable concept にするか
- session 集約が timeframe 集約と独立した責務か、それとも marketdata aggregation の別モードか
- session 境界の source of truth を domain に置くか、data/query 側の再利用資産に落とすか
- 夏時間や市場カレンダー例外を今回の最小スコープに含めるか

#### Expected Impact

- `implementations/dag-core/internal/domain/marketdata/` に `Session` と `SessionBar` を追加する可能性が高い
- `timeframe_bar` とは別テーブルにするか、既存 schema に session 軸を加えるかを判断する必要がある
- HTTP / workflow / payload / OpenAPI に `SessionBar` を渡す入口を追加する場合、`OHLCV` 追加時と同種の波及が発生する
- reusable な session semantics に昇格する場合は `blueprint/architecture/flow-and-time-model.md` への promotion を検討する

#### Notes

- `SessionBar` は `TimeframeBar` の単純な拡張とは限らない
- session はタイムゾーン、営業日、日跨ぎ、夏時間、市場休日を含み得るため、最初の実装範囲を明示する必要がある
- parent repository の reusable concept にする場合は、実装より先に `blueprint/` 側の扱いを決める

## Next Actions

- Track 1 を実 DB で確認する場合は、開発 DB をリセットして updated migration / seed を適用する
- Track 1 をリセットなし DB に適用する場合は、`high_time` / `low_time` を追加する forward migration または手動 `ALTER TABLE` 方針を別途決める
- Track 1 の PR / handoff では `go test ./...`、`make tsdb-schema-lint`、`make tsdb-query-lint`、`make sqlc-gen`、`make openapi` の実行結果を添える
- Track 2 は implementation-local concept として始めるか、parent repository が継承させる reusable concept として扱うかを切り分ける
- Track 2 は既存 `timeframe_bar` と同じ persistence line に乗せるか、別 schema asset に切るかを比較する

## References

- `blueprint/architecture/flow-and-time-model.md`
- `blueprint/conventions/data-conventions.md`
- `data/tsdb/schema/migrate/000050_create_timeframe_bar.sql`
- `data/tsdb/query/000090_marketdata_timeframe_bar.sql`
- `implementations/dag-core/internal/domain/marketdata/ohlc.go`
- `implementations/dag-core/internal/domain/marketdata/timeframe_bar_aggregation.go`
- `implementations/dag-core/internal/domain/marketdata/timeframe_bar_aggregation_test.go`
- `implementations/dag-core/internal/application/dagruntime/usecase/payload_parse.go`
- `implementations/dag-core/internal/infrastructure/persistence/tsdb/mapper/marketdata/timeframe_bar.go`
- `implementations/dag-core/internal/infrastructure/persistence/tsdb/repository/marketdata/timeframe_bar_integration_test.go`
- `implementations/dag-core/internal/application/marketdata/usecase/timeframe_bar_integration_test.go`
- `implementations/dag-core/openapi/`

## Promotion Targets

- durable な時系列バー意味論に昇格する場合: `blueprint/architecture/flow-and-time-model.md`
- durable な session semantics に昇格する場合: `blueprint/architecture/flow-and-time-model.md` または relevant `blueprint/conventions/`
- durable な TSDB schema/query の決着: `data/tsdb/schema/` と `data/tsdb/query/`
- 実装固有の入出力更新: `implementations/dag-core/`

## Acceptance Criteria

- marketdata 見直しの論点と実装順序を 1 ファイルで再開できる
- `OHLCV` / `timeframe_bar` 変更の主要影響範囲が task 文書として共有されている
- `Session` / `SessionBar` の検討論点が `OHLCV` 変更と混ざらず確認できる
- durable source of truth に昇格すべき論点と temporary context の境界が明確である
- Track 1 の実装済み範囲と、実 DB 適用時の前提が確認できる

## Result

- marketdata 見直しテーマの temporary task context を `governance/tasks/marketdata_review_followups.md` に追加した
- 直近テーマを `OHLCV` の `Hightime` / `Lowtime` 追加と TSDB 永続化影響にフォーカスして整理した
- `Hightime` / `Lowtime` の意味論を last-touch として確定した
- 開発環境リセット前提で既存 migration を直接修正する方針を反映した
- `Session` / `SessionBar` 検討を独立した work track として整理した
- Track 1 の SQL / Go domain / aggregation / payload parser / TSDB mapper / sqlc generated code / OpenAPI generated artifacts を更新した
- Track 1 の tick aggregation は open/close に mid、high に ask、low に bid を採用し、source を `tick_bid_ask_mid` に更新した
- Track 1 の last-touch semantics、payload mapping、TSDB repository round-trip、backfill persistence の確認点をテストに追加した
- 検証として `go test ./...`、`make tsdb-schema-lint`、`make tsdb-query-lint` を実行し、通過を確認した
