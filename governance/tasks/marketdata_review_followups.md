# Marketdata 見直しフォローアップ

## Background

marketdata まわりの見直しを進めるため、一時的な共有作業文脈をここに置く。

この文書は恒久仕様ではなく、調査・設計判断・実装順序をまとめるための temporary context として扱う。
durable な結論は `blueprint/`、`data/tsdb/`、`implementations/dag-core/` の適切な正本へ昇格させる。

## In Scope

- `OHLCV` と `TimeframeBar` のフィールド見直し
- `Hightime` / `Lowtime` の意味論と永続化方針の整理
- `marketphase` 導入と phase-based context 解決
- `phase_bar` 導入検討と phase-based bar persistence
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

#### Status

Superseded by Track 4. `session` / `session_bar` は廃止し、`marketphase` に寄せる。

#### Goal

この track は一度 `Session` / `SessionBar` concept を導入したが、Track 4 の `marketphase` と責務が重複するため正本にはしない。

#### Working Definition

- `Session` は「特定タイムゾーン上の繰り返し時間窓」を表す marketdata domain concept とする
- `Session` は runtime の partition / state transaction の session とは別概念として扱う
- `SessionBar` は session window に含まれる tick を集約した OHLCV bar とする
- `SessionBar.OHLCV.Opentime` / `Closetime` は、session definition から解決された UTC instant とする
- session interval は `[open, close)` の half-open interval とする
- `SessionDate` は session open が属する local date とする

#### Decisions

- `SessionBar` は `TimeframeBar` と並ぶ別の bar concept として扱う
- `Session` は local recurring window の value object として扱う
- first scope は `TOKYO` / `Asia/Tokyo` / `09:00-15:00`
- first implementation は domain-only とし、`implementations/dag-core/internal/domain/marketdata/` に追加した
- aggregation の OHLC semantics は Track 1 の `TimeframeBar` と揃え、open/close は bid/ask mid、high は ask、low は bid、high/low time は last-touch とする
- `SessionBar.Source` は tick 由来の場合、`TimeframeBar` と同じく `tick_bid_ask_mid` を候補とする
- persistence / API exposure / DST / calendar rule は Track 3 に送る

#### First Scope

- `SessionCode`
- `Session`
  - code
  - timezone name: `Asia/Tokyo`
  - local open time
  - local close time
- `SessionBar`
  - symbol_id
  - session_code
  - session_date
  - embedded `OHLCV`
  - source
- session window resolver
  - input: `Session`, target date or tick time
  - output: `[open_time, close_time)` as UTC `UTCTime`
- tick aggregation
  - input: sorted ticks and resolved session window
  - output: zero or one `SessionBar`

#### Initial Session Set

- `TOKYO`
  - timezone: `Asia/Tokyo`
  - local window: `09:00` to `15:00`
  - interval semantics: `[09:00, 15:00)`
  - `SessionDate`: local open date
  - no DST handling needed

#### Explicitly Out Of First Scope

- persistence
- HTTP / workflow payload / OpenAPI exposure
- DST rule
- market calendar rule
- holiday calendars
- early close / special trading day calendars
- `LONDON` / `NEW_YORK`
- multi-session merge bars
- overlapping session conflict resolution
- exchange-specific lunch breaks or split sessions
- repository-wide session taxonomy

#### Open Questions

- `SessionDate` の Go 表現を専用型にするか、local date string / struct として扱うか
- `Session` の local time 表現を専用型にするか、`time.Duration` after midnight として扱うか
- ticks が session window 外を含む場合、aggregator が filter するか、caller が windowed ticks を渡す前提にするか

#### Result

- `Session` / `SessionBar` の試作は廃止した
- market session 相当の意味論は single phase を `marketphase` に、composite signal / strategy-facing context を `marketcontext` に分離して扱う
- `SessionDate` / `SessionLocalTime` / `TokyoSession()` / `AggregateSessionBar` は削除した

#### Expected Impact

- `implementations/dag-core/internal/domain/marketdata/` に `Session`、`SessionCode`、`SessionBar`、session window resolver を追加した
- first implementation は domain-only のため、`data/tsdb/`、HTTP、workflow payload、OpenAPI への影響は発生しない
- Track 3 で persistence / API exposure / DST / calendar rule を扱う場合、`blueprint/architecture/flow-and-time-model.md` への promotion を再検討する

#### Notes

- `SessionBar` は `TimeframeBar` の単純な拡張とは限らない
- first scope では session はタイムゾーンと local recurring window だけを扱う
- 現行 `TimeframeCode.Window` には `sessionOffset` という名前の UTC+22:00 起点があるが、これは timeframe bucket 起点であり、Track 2 の market session definition とは分けて扱う

### Track 3: Session Persistence / Exposure / Calendar Rules

#### Status

Cancelled. `session_bar` persistence は採用しない。

#### Goal

`session_bar` persistence の検討経緯を残す。今後の正本は `marketphase` であり、`session_bar` table は維持しない。

#### Decisions

- `SessionBar` persistence は `timeframe_bar` とは別の `session_bar` table として追加する
- `timeframe_bar` に session 軸を追加しない
- `session_bar` の primary key は `(symbol_id, session_code, session_date)` を候補とする
- `session_date` は Track 2 の決定に従い、session open が属する local date とする
- `session_date` の SQL 型は `DATE` とする
- `open_time` / `close_time` は resolved UTC instant とする
- OHLC semantics は Track 1 / Track 2 と揃え、open/close は bid/ask mid、high は ask、low は bid、high/low time は last-touch とする
- `source` は tick 由来の場合 `tick_bid_ask_mid` を使う
- query first scope は `BulkUpsert`, `GetLatestBySymbolAndSession`, `ListBySymbolSessionAndDateRange`, `DeleteBySymbolSessionAndDateRange` とする

#### Deferred Scope

- HTTP API exposure
- workflow payload / OpenAPI exposure
- DST rule
- market calendar rule
- holiday calendars
- early close / special trading day calendars
- `LONDON` / `NEW_YORK`
- repository-wide reusable session taxonomy

#### Persistence Options

Chosen: separate `session_bar` table.

- Keeps `timeframe_bar` clean and avoids overloading `timeframe_code`
- Supports session-specific identity: `(symbol_id, session_code, session_date)`
- Makes session storage reviewable as a TSDB asset under `data/tsdb/`

Rejected for this track: keep domain-only.

- Pros: keeps session semantics cheap to iterate
- Pros: avoids premature schema truth before boundary rules settle
- Cons: no reusable historical `SessionBar` storage

Rejected for this track: add session columns to `timeframe_bar`.

- Pros: fewer tables
- Cons: conflates fixed-duration timeframe bars with calendar/session bars
- Cons: primary key and query semantics become ambiguous

#### Open Questions

- `session_bar` table に `timezone` / local open-close definition snapshot を持たせるか、`session_code` と resolved UTC times のみにするか
- `SessionBar` を workflow payload / OpenAPI に露出する入口を作るか
- `Session` definition を Go domain constant から SQL seed / queryable table へ移すか
- London / New York を扱う前に DST rule を reusable concept として `blueprint/architecture/flow-and-time-model.md` に昇格するか
- market calendar rule を parent repository の reusable concept にするか、derivative project specific にするか

#### Expected Impact If Adopted

- `data/tsdb/schema/migrate/`: create `session_bar`
- `data/tsdb/query/`: add `session_bar` query asset
- sqlc generated code
- TSDB mapper / repository
- domain `SessionBar` persistence mapping
- fixture / integration test
- HTTP request / response only if API exposure is adopted later
- workflow payload parser / encoder only if workflow exposure is adopted later
- OpenAPI generated artifacts only if API / payload exposure is adopted later
- possible `blueprint/architecture/flow-and-time-model.md` promotion

#### Result

- `session_bar` table, query asset, sqlc generated code, mapper, repository, test は削除した
- TSDB persistence は `session` concept に依存させず、必要なら将来 `marketphase` ベースで再設計する

### Track 4: Market Phase Resolver with DST Support

#### Status

Domain-local first implementation completed. Strategy / workflow integration remains pending.

#### Goal

FX / macro trading systems 向けに、市場ローカル時刻基準で単一市場フェーズを解決し、DST を timezone conversion に委ねられるようにする。複合シグナルと strategy-facing context は adjacent domain で扱う。

この track は order execution ではなく、market structure interpretation 用の context 解決を対象とする。

#### Scope

- major markets: `Tokyo`, `London`, `New York`, `Sydney`
- single phase resolution
- deterministic test coverage
- composite transition / overlap signal を別領域へ切り出すための境界整理

#### Core Decisions

- London / New York を fixed JST range では定義しない
- single phase は各市場の local timezone wall clock で定義する
- `marketphase` domain は single phase のみを扱う
- `PhaseCategory` という型と category-based taxonomy は採用しない
- composite transition / overlap signal は phase bar の対象にしないため、`marketphase` とは別領域で扱う
- DST offset は手書きせず、IANA timezone name と標準 timezone conversion に委ねる
- core resolver は explicit timestamp を受ける pure function 寄りの設計にする
- strategy code は timezone math を直接持たず、必要なら別領域の stable な market context を参照する

#### Single Phase Definitions

- `pre_tokyo`
  - market: `Tokyo`
  - timezone: `Asia/Tokyo`
  - local window: `08:00-09:00`
- `tokyo_core`
  - market: `Tokyo`
  - timezone: `Asia/Tokyo`
  - local window: `09:00-15:00`
- `late_tokyo`
  - market: `Tokyo`
  - timezone: `Asia/Tokyo`
  - local window: `15:00-16:00`
- `london_early`
  - market: `London`
  - timezone: `Europe/London`
  - local window: `08:00-11:00`
- `london_core`
  - market: `London`
  - timezone: `Europe/London`
  - local window: `11:00-13:00`
- `london_late`
  - market: `London`
  - timezone: `Europe/London`
  - local window: `13:00-16:00`
- `pre_newyork`
  - market: `New York`
  - timezone: `America/New_York`
  - local window: `07:00-08:00`
- `newyork_core`
  - market: `New York`
  - timezone: `America/New_York`
  - local window: `08:00-11:00`
- `late_newyork`
  - market: `New York`
  - timezone: `America/New_York`
  - local window: `11:00-16:00`
- `sydney_reset`
  - market: `Sydney`
  - timezone: `Australia/Sydney`
  - local window: `07:00-09:00`

#### Derived Phase Definitions

- `tokyo_london_transition`
  - composite signal として別領域で扱う
  - active when `late_tokyo` is active and London local time is in `[07:00, 13:00)`
  - interpretation: Tokyo-led flow から Europe-led flow への handoff
- `london_newyork_overlap`
  - active when (`london_core` or `london_late`) and (`pre_newyork` or `newyork_core`) are active
  - interpretation: highest liquidity / breakout continuation potential / macro sensitivity
- `newyork_close_transition`
  - optional
  - first scope では未実装でもよい
  - 採用する場合は deterministic な activation rule を別途固定する
  - 例: `late_newyork` の local `15:00-16:00` に限定し、かつ他の主要 single phase が active でない場合
  - interpretation: thinning liquidity / unwind / mean-reversion risk

#### Proposed Domain Shapes

- single phase spec
  - `ID`
  - `Market`
  - `Timezone`
  - `StartHour`
  - `StartMinute`
  - `EndHour`
  - `EndMinute`
  - `Tags`
  - `Notes`
- resolved single phase
  - `ID`
  - `Market`
  - `Timezone`
  - `Active`
  - `LocalNow`
  - `LocalStart`
  - `LocalEnd`
  - `UTCStart`
  - `UTCEnd`
  - `JSTStart`
  - `JSTEnd`
- separate context layer
  - strategy-facing context は `marketphase` package の外で扱う
  - composite signal id, tag merge, dominant market assembly も phase domain の外へ置く

#### Resolution Flow

1. For each single `PhaseSpec`, load IANA location and convert input UTC into local time.
   - first implementation では `time.Location` を毎回 hot path で解決しない
   - spec 初期化時または resolver 初期化時に timezone validation / cache を済ませる
2. Compare local wall-clock against half-open interval `[start, end)`.
3. Collect all active single phases as `ResolvedPhase`.
4. Let adjacent context logic consume active single phases when composite signals are needed.

#### Suggested Implementation Boundary

- package candidate: `implementations/dag-core/internal/domain/marketphase/`
- expected files
  - `phase.go`
  - `spec.go`
  - `resolver.go`
  - `phase_bar.go`
  - `phase_bar_aggregation.go`
  - `resolver_test.go`
- expected API
  - `DefaultPhaseSpecs() []PhaseSpec`
  - `ResolveSinglePhases(at time.Time, specs []PhaseSpec) ([]ResolvedPhase, error)`
  - `AggregatePhaseBar(phase ResolvedPhase, symbolID marketdata.SymbolID, ticks []marketdata.Tick) (PhaseBar, bool, error)`

#### Required Test Coverage

- Tokyo core on a normal day
- London early during UK winter
- London early during UK summer
- New York core during US winter
- New York core during US summer
- period where US and UK DST are misaligned
- boundary condition correctness with half-open interval semantics
- evidence that London / New York JST-equivalent windows shift seasonally while local definitions stay stable

#### Explicit Anti-Requirements

- London / New York phases を fixed JST constants で定義しない
- DST offset value を hardcode しない
- `marketphase` package に composite phase / transition signal を混在させない
- strategy code に timezone conversion を直接持ち込まない

#### Future Extension Hooks

- Tokyo fixing window
- London fix
- NY option cut
- macro event overlays
- holiday calendars
- half-day / special trading day effects
- pair-specific context transformation

#### Open Questions

- `newyork_close_transition` を first scope に含めるか、optional composite signal として postpone するか
- 含める場合、`newyork_close_transition` の deterministic activation window を local clock ベースでどこに固定するか
- adjacent context layer の tag merge policy を set-union のみで始めるか、priority / dedupe ordering を入れるか
- event overlay や holiday calendar を将来追加する前提で `PhaseSpec` と resolver option をどこまで分けるか
- London / New York / Sydney の calendar effect を扱う前に `blueprint/architecture/flow-and-time-model.md` に reusable timezone / phase rule を昇格するか

#### Promotion Targets

- reusable な market-phase / timezone rule に昇格する場合: `blueprint/architecture/flow-and-time-model.md`
- implementation-local な phase resolver code: `implementations/dag-core/internal/domain/`
- scenario-based validation を追加する場合: `scenarios/`

#### Implementation Result

- `implementations/dag-core/internal/domain/marketphase/` を追加した
- `DefaultPhaseSpecs`, `ResolveSinglePhases`, `AggregatePhaseBar` を single phase 専用 API として整理した
- single phase は各市場の IANA timezone に変換して local wall-clock `[start, end)` で判定する
- timezone resolution は package 内 cache を使い、hot path で spec ごとに `time.LoadLocation` を繰り返さない
- `PhaseCategory` と category-based branching は削除した
- table-driven tests で Tokyo / London / New York の seasonal window、US-UK DST misalignment、boundary を確認した
- composite signal と strategy-facing context は `implementations/dag-core/internal/domain/marketcontext/` に分離した
- `marketcontext` 側で `tokyo_london_transition`, `london_newyork_overlap`, `newyork_close_transition` と `Context` を扱う

#### Deferred Integration

- 既存 `filter_session` の fixed UTC hour 判定を single phase / market context に置き換えること
- strategy node / DAG node が `marketcontext.Context` を直接参照する integration
- holiday / fix / option cut / macro overlay の追加
- `session` terminology を workflow / config / doc から整理すること

### Track 5: Phase Bar Persistence

#### Status

Implemented for single phases. Derived phase persistence remains out of scope.

#### Goal

`marketphase` を正本として、single phase window に含まれる tick から `phase_bar` を集約・永続化できるようにする。

#### Decisions

- `phase_bar` は `single phase only` とする
- composite signal は保存せず、adjacent context layer から都度導出する
- identity は `(symbol_id, phase_id, open_time)` とする
- `session_date` のような local-date identity は使わない
- `phase_bar` は resolved UTC `[open_time, close_time)` を正本として保持する
- OHLC semantics は `timeframe_bar` と揃え、open/close は bid/ask mid、high は ask、low は bid、high/low time は last-touch とする
- `market` と `timezone` は phase metadata snapshot として row に保持する

#### Expected Shape

- schema
  - `symbol_id`
  - `phase_id`
  - `market`
  - `timezone`
  - `open_time`
  - `close_time`
  - `open`
  - `high`
  - `high_time`
  - `low`
  - `low_time`
  - `close`
  - `volume`
  - `source`
- repository API
  - `BulkUpsert`
  - `GetLatestBySymbolAndPhase`
  - `ListBySymbolPhaseAndRange`
  - `DeleteBySymbolPhaseAndRange`

#### Explicit Anti-Requirements

- composite signal を `phase_bar` として保存しない
- `phase_id + local_date` を primary identity にしない
- `timeframe_bar` に phase 列を足して兼用しない
- London / New York を fixed JST / fixed UTC で phase bar 化しない

#### Implementation Result

- `data/tsdb/schema/migrate/000070_create_phase_bar.sql` を追加した
- `data/tsdb/query/000100_marketphase_phase_bar.sql` を追加した
- `implementations/dag-core/internal/domain/marketphase/PhaseBar` と `AggregatePhaseBar` を追加した
- `AggregatePhaseBar` は resolved single phase の UTC window を使って tick を half-open interval で集約する
- TSDB mapper / repository に `PhaseBarRepository` を追加した
- `UnitOfWork` / DI container / fake repositories を更新して `PhaseBars()` を通した
- sqlc generated code を更新した
- domain test で phase window 集約を確認し、TSDB integration test を追加した

#### Deferred Scope

- composite signal persistence
- strategy / workflow からの `phase_bar` 利用
- `phase_spec_version` / `phase_spec_hash` の snapshot 管理
- holiday / fix / option cut overlay 反映後の bar semantics 再設計

## Next Actions

- Track 1 を実 DB で確認する場合は、開発 DB をリセットして updated migration / seed を適用する
- Track 1 をリセットなし DB に適用する場合は、`high_time` / `low_time` を追加する forward migration または手動 `ALTER TABLE` 方針を別途決める
- Track 1 の PR / handoff では `go test ./...`、`make tsdb-schema-lint`、`make tsdb-query-lint`、`make sqlc-gen`、`make openapi` の実行結果を添える
- `marketcontext` の composite signal rules を strategy / DAG integration に通す
- `tokyo_london_transition` の final rule を task 文書と code で一致させたまま維持する
- Track 4/5 の durable rule が固まったら、timezone-local single phase definition と DST handling 方針のうち reusable な部分だけを `blueprint/architecture/flow-and-time-model.md` へ昇格する
- holiday / fixing / option cut を入れる場合も `marketphase` には single phase だけを残す

## References

- `blueprint/architecture/flow-and-time-model.md`
- `blueprint/conventions/data-conventions.md`
- `data/tsdb/schema/migrate/000050_create_timeframe_bar.sql`
- `data/tsdb/query/000090_marketdata_timeframe_bar.sql`
- `implementations/dag-core/internal/domain/marketdata/ohlc.go`
- `implementations/dag-core/internal/domain/marketdata/timeframe_bar_aggregation.go`
- `implementations/dag-core/internal/domain/marketdata/timeframe_bar_aggregation_test.go`
- `implementations/dag-core/internal/domain/marketphase/`
- `implementations/dag-core/internal/domain/marketcontext/`
- `implementations/dag-core/internal/application/dagruntime/usecase/payload_parse.go`
- `implementations/dag-core/internal/infrastructure/persistence/tsdb/mapper/marketdata/timeframe_bar.go`
- `implementations/dag-core/internal/infrastructure/persistence/tsdb/repository/marketdata/timeframe_bar_integration_test.go`
- `implementations/dag-core/internal/application/marketdata/usecase/timeframe_bar_integration_test.go`
- `implementations/dag-core/openapi/`
- supplied market phase resolver spec from review follow-up

## Promotion Targets

- durable な時系列バー意味論に昇格する場合: `blueprint/architecture/flow-and-time-model.md`
- durable な session semantics に昇格する場合: `blueprint/architecture/flow-and-time-model.md` または relevant `blueprint/conventions/`
- durable な market phase / DST handling semantics に昇格する場合: `blueprint/architecture/flow-and-time-model.md`
- durable な composite market context semantics に昇格する場合: `blueprint/architecture/flow-and-time-model.md`
- durable な TSDB schema/query の決着: `data/tsdb/schema/` と `data/tsdb/query/`
- 実装固有の入出力更新: `implementations/dag-core/`

## Acceptance Criteria

- marketdata 見直しの論点と実装順序を 1 ファイルで再開できる
- `OHLCV` / `timeframe_bar` 変更の主要影響範囲が task 文書として共有されている
- `Session` / `SessionBar` の検討論点が `OHLCV` 変更と混ざらず確認できる
- market phase resolver の local-time / DST / single-phase-only 方針が temporary context として再開できる
- composite signal / market context が phase domain とは別領域で再開できる
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
- Track 2 の domain-only `Session` / `SessionBar` / aggregation を実装した
- Track 2 の検証として `go test ./internal/domain/marketdata` と `go test ./...` を実行し、通過を確認した
- Track 4 として market phase resolver の仕様を `governance/tasks/` の temporary context に整理し、local timezone 基準、DST handling、single-phase-only 境界、test coverage、promotion target を明記した
- Track 4/5 の実装整理として `marketphase` domain を single phase 専用に戻し、composite signal / market context を別領域へ分離した
