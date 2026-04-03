# workflows 開発者ガイド

このドキュメントは `implementations/dag-core/workflows/` 配下の Workflow YAML を追加・変更する開発者向けの実装ガイドです。

## 1. 役割と責務

- `workflows/*.yaml` は DAG 実行順序とノード構成の宣言です。
- 実行ロジック本体は `internal/application/dagruntime/spec/factory/` 側の NodeFactory/Node 実装が担います。
- このディレクトリは「定義（宣言）」の責務に限定し、実行コードは持ちません。

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

`inputs` に書いた文字列は `internal/di/runtime_workflow.go` の入力マップで `artifact.Key` に変換されます。未登録名はコンパイル失敗します。

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
3. `inputs` は `runtime_workflow.go` の入力マップにあるキーのみ使う
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

パス解決は `internal/di/runtime_workflow.go` の `resolveWorkflowSpecPath` が行います。

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
