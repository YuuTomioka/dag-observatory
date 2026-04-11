# workflows 開発者ガイド

このドキュメントは `implementations/dag-core/workflows/` 配下の Workflow YAML を追加・変更する開発者向けの実装ガイドです。

## 1. 役割と責務

- `workflows/*.yaml` は DAG 実行順序とノード構成の宣言です。
- 実行ロジック本体は `internal/application/dagruntime/spec/factory/` 側の NodeFactory/Node 実装が担います。
- このディレクトリは「定義（宣言）」の責務に限定し、実行コードは持ちません。

## 1.1 現在の配置方針（2026-04 時点）

- 現在は `implementations/dag-core/workflows/` 直下に YAML を配置するフラット構成を採用しています。
- `workflows/smoke` / `workflows/marketdata` / `workflows/strategy` のようなサブディレクトリ分割は、workflow 数増加時の条件付き対応です。
- 分割トリガーの目安:
  - workflow ファイル数増加で探索コストが上がる
  - ドメイン別レビュー導線が不明瞭になる
  - 新規追加時に命名/配置の迷いが継続的に発生する
- 現段階では `DAGRUNTIME_WORKFLOW_SPEC_PATH=workflows/<name>.yaml` を前提に運用します。

## 1.2 breakout 系 workflow の現在位置づけ（2026-04-05 時点）

既存の breakout 系 YAML は、同じ strategy family の中で責務を段階的に切り分けた参照資産として扱います。

- `breakout_long_v0.yaml`
  - signal / filter / sizing / decision / observability の基線確認用です。
  - paper execution と position update は含みますが、fill progression, close, summary の正本 workflow としては扱いません。
  - 主目的は entry decision 周辺の構造確認と観測導線の最小検証です。

- `breakout_long_execution_position_minimal.yaml`
  - market order submit と position snapshot load を含む、execution 境界確認用です。
  - fill confirmation, close, summary は同一 workflow に抱え込まず、execution 直後の state / artifact 形状を確認するための最小構成です。
  - submit/fill/close を分割したときの entrypoint 候補として扱います。

- `breakout_long_v1_extended.yaml`
  - filter/risk/execution/position management をまとめた拡張検証用です。
  - trailing stop や timeout exit のような position-management ノードを含みますが、現時点では closed trade store や strategy summary update までは接続していません。
  - 将来の richer validation 用サンプルであり、現段階の結果正本 workflow ではありません。

- `breakout_long_result_reflection_minimal.yaml`
  - result reflection 専用の最小 workflow です。
  - open position state と現在 bar を使って close 判定を行い、`closed_trades` / `daily_pnl` / `strategy_summary` を更新します。
  - submit 系 workflow とは分離し、結果確定後の反映責務だけを担います。

これらの YAML は「どれが最終完成版か」を争うものではなく、責務境界の異なる比較対象として維持します。

## 1.3 breakout 系の分割方針

breakout 系では、entry decision / execution / position / close / summary を次の原則で扱います。

- entry decision は YAML で表現する。
  - strategy 差分が最も出やすく、宣言的に比較したい層だからです。

- execution と position management も YAML で表現する。
  - submit, snapshot load, trailing, timeout などの接続順は workflow 宣言として比較可能にしておくべきだからです。

- close と summary も node としては YAML で接続可能にする。
  - ただし常に entry workflow と同一ファイルへ載せる前提にはしません。

- state truth は `closed_trades` と `strategy_summary` に寄せる。
  - observability event や WebSocket 的通知は補助導線であり、結果の正本にはしません。

## 1.4 1 本に残すものと分けるものの基準

1 本の workflow に残すのは、同一 cycle で完結し、入力 artifact と state write の因果がその場で説明できる範囲に限ります。

- 1 本に残してよいもの
  - signal, filter, sizing, decision
  - submit のような command 生成
  - snapshot load や trailing/timeout のような position-management
  - paper execution のような cycle 内で完結する擬似処理

- 別 workflow または別 event entrypoint に分けるもの
  - fill confirmation のように外部 execution 結果を待つ処理
  - close result の確定
  - `closed_trade_store`, `daily_pnl_update`, `open_position_close`, `strategy_summary_update` のような結果確定後の反映

判断基準は次の 3 点です。

- 外部 event を待つか
- retry / idempotency の境界が submit 系と result 反映系で異なるか
- 失敗時に snapshot 再読込で回復できる read/write 境界を保ちたいか

この基準により、submit 系 workflow と result reflection 系 workflow は分割を基本とします。close と summary は result reflection 側に寄せ、entry decision 側へ常設しません。

## 1.5 現時点の採用方針

2026-04-05 時点では、次を採用します。

- `breakout_long_v0.yaml` は entry decision baseline
- `breakout_long_execution_position_minimal.yaml` は execution / position 最小検証
- `breakout_long_v1_extended.yaml` は position-management を含む拡張検証
- `breakout_long_result_reflection_minimal.yaml` は result reflection 最小検証
したがって、現時点では submit 系と result reflection 系を別 YAML として維持します。

## 2. Workflow 起点の依存関係

Workflow は以下の順で実行可能な `pipeline.Compiled` に変換されます。

1. `workflows/*.yaml`
2. `spec/loader.LoadFile` で YAML 読み込み
3. `spec/validator.ValidateWorkflowSpec` で構文・kind・config 妥当性チェック
4. `spec/adapter.ToWorkflow` で `spec.NodeSpec` を `node.Node` へ変換
5. `workflow.Compile` で依存解決とトポロジカルソート
6. `engine.Runner` が 1 cycle 実行

主な配線ポイント:

- エントリ: `internal/di/runtime_workflow.go`
- compile 集約: `internal/application/dagruntime/spec/compiler/`
- spec 型: `internal/application/dagruntime/spec/spec.go`
- kind 登録: `internal/application/dagruntime/spec/factory/builtin.go`

## 3. YAML スキーマ（v1）

```yaml
name: dagruntime.example
version: v1
inputs:
  - market.symbol
nodes:
  - id: example_node
    kind: some_kind
    config:
      foo: bar
```

- `name`: workflow 名（空不可）
- `version`: 現在は `v1` のみ
- `inputs`: 実行時入力キー名
- `nodes[].id`: ノード識別子（workflow 内で一意）
- `nodes[].kind`: 登録済み kind 名
- `nodes[].config`: kind ごとの設定

## 4. inputs の解決ルール

`inputs` に書いた文字列は `internal/application/dagruntime/spec/inputmap/default.go` の入力マップで `artifact.Key` に変換されます。未登録名はコンパイル失敗します。

現状の代表キー:

- `symbol`, `mode`
- `market.symbol`, `market.bars`, `market.ohlcv_bars`, `market.ohlcv_bars.h1`, `market.tick`, `market.spread_bps`
- `account.balance`
- `marketdata.symbol_id`, `marketdata.symbol_code`, `marketdata.timeframe_code`, `marketdata.from`, `marketdata.to`

## 5. ノード間依存（artifact/state）

依存は YAML の並び順ではなく Node 実装の `Requires()/Provides()/Writes()` で解決されます。

- `Requires`: 上流ノードまたは `inputs` から供給される必要がある
- `Provides`: 他ノードへ渡す artifact
- `Writes`: 永続 state への書き込み

`workflow.Compile` で以下を検出します。

- input と provides の衝突
- artifact provider 重複
- state writer 重複
- 未解決 require
- 循環依存（cycle）

## 6. Workflow 追加手順

### 6.1 既存 kind だけで新規 Workflow を追加する場合

1. `workflows/<name>.yaml` を追加する
2. `name` は `dagruntime.<domain>_<purpose>` 形式で命名する
3. `inputs` は `spec/inputmap/default.go` の入力マップにあるキーのみ使う
4. `nodes` の `kind/config` を既存 factory 仕様に合わせる
5. `internal/di/runtime_workflow_test.go` に compile テストを追加する
6. 必要なら E2E 相当の runner テストを追加する
7. `go test ./internal/di/...` と `go test ./internal/application/dagruntime/spec/...` を通す

### 6.2 新しい kind を追加して Workflow で使う場合

1. `internal/application/dagruntime/spec/factory/` に NodeFactory と Node 実装を追加
2. `Kind()` を定義し、`Build` で config バリデーションを実装
3. `Requires/Provides/Reads/Writes/Spec/Run` を実装
4. `NewBuiltinRegistryWithDependencies` に factory を登録
5. factory 単体テストを追加（不正 config ケースを含む）
6. Workflow YAML で新 kind を参照
7. compile テスト・必要な実行テストを追加

## 7. 実行時の Workflow 切り替え

`DAGRUNTIME_WORKFLOW_SPEC_PATH` で起動時に読み込む YAML を切り替えます。

例:

- `DAGRUNTIME_WORKFLOW_SPEC_PATH=workflows/default.yaml`
- `DAGRUNTIME_WORKFLOW_SPEC_PATH=workflows/marketdata_timeframe_bar_backfill.yaml`
- `DAGRUNTIME_WORKFLOW_SPEC_PATH=workflows/breakout_long_v1_extended.yaml`
- `DAGRUNTIME_WORKFLOW_SPEC_PATH=workflows/breakout_long_result_reflection_minimal.yaml`

パス解決は `internal/application/dagruntime/spec/compiler/resolver.go` が行います。

## 8. 典型的な失敗と確認ポイント

- `kind "... is not registered"`: builtin registry 未登録
- `input "... is not mapped"`: `inputs` 名が入力マップ未登録
- `duplicate node id`: `nodes[].id` 重複
- `config validation failed`: kind 側の config 不一致
- `cycle detected in workflow`: `Requires/Provides` 依存が循環

## 9. レビュー観点

- Workflow は宣言責務に収まっているか（実装ロジックを YAML に埋め込んでいないか）
- `inputs` が最小で過不足ないか
- kind 設定値が factory 側バリデーションと一致しているか
- 依存関係が明確で、観測・データ副作用の意図が説明できるか
- 追加した Workflow を再現可能にするテストがあるか
