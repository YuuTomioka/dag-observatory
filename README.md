# dag-observatory

Clock + DAG 実行基盤、時系列データ設計、観測性設計を扱う親リポジトリです。

`README.md` は入口です。構造や仕様の正本は `blueprint/`、変更ルールは `governance/`、実装の正本は `implementations/` にあります。

文書運用の基本方針:

- 正本となる構造・仕様ドキュメントは英語
- 人間向けの要求伝達、手順、検証導線は日本語
- 詳細ルールは `governance/policies/documentation-policy.md` を参照

## 最初に読むもの

1. `blueprint/identity/repository-purpose.md`
2. `blueprint/architecture/system-boundaries.md`
3. `blueprint/architecture/flow-and-time-model.md`
4. `blueprint/conventions/implementation-principles.md`
5. `blueprint/conventions/observability-conventions.md`
6. `blueprint/conventions/data-conventions.md`
7. `governance/policies/change-policy.md`
8. `governance/policies/documentation-policy.md`
9. `scenarios/representative-scenario.md`
10. 対象の実装、platform、data、deployments 配下

AI ペアプロ向けの読解順と変更原則は `AGENTS.md` にあります。

## ルート直下の責務

- `blueprint/`: 親リポジトリとして継承させたい設計原則と ADR の正本
- `governance/`: 変更分類、文書配置、共有タスク文脈の管理
- `governance/contracts/`: 人間向けの要求受理、実装ハンドオフ、検証報告テンプレート
- `governance/guides/`: 人間向けの進め方や運用手順の集約先
- `implementations/`: 参照実装の正本
- `platform/`: 観測性などの恒久的な運用基盤リファレンス
- `deployments/`: ローカル実行や移行のための実行ラッパ
- `data/`: 時系列・分析向け SQL 資産と設計資産
- `scenarios/`: 日本語で管理する代表ユースケースと推奨導線
- `.agents/skills/`: このリポジトリで使う Codex の repository-scoped skills
- `scripts/`: 補助スクリプト
- `.codex/`: AI ローカル補助
- `var/`: 共有しないローカル runtime 出力・キャッシュ置き場

補足:

- 一時的な共有タスク文脈は `governance/tasks/`
- `.wrk/` は過去のローカル scratch 置き場であり、今後の参照先にはしない

## このリポジトリが保持するもの

- DAG runtime の構造原則
- event, artifact, state, partition, time に関する共通モデル
- 観測性を後付けではなく構造要件として扱う方針
- Go と Python の参照実装
- TimescaleDB 系の SQL 資産
- ローカル検証と代表シナリオ

## このリポジトリが保持しないもの

- 単一プロダクトの完全な業務仕様
- 環境固有の本番デプロイ判断
- 一時メモを恒久ドキュメントとして残す運用

## 主要ディレクトリの見方

### `blueprint/`

構造上の正本です。派生先へ継承したい境界、フロー、時間モデル、観測性、データ原則はここで定義します。

構造変更や責務変更を伴うときは、実装より先に `blueprint/` を更新します。

### `implementations/`

参照実装の物理正本です。

- `implementations/dag-core/`: Go 参照実装
- `implementations/worker-py/`: Python worker 参照実装

実装固有の API 詳細や handler、workflow、worker handler などはここで管理します。

### `governance/contracts/`

要求受理、実装依頼、検証報告のような人間向けの受け渡し面を置く場所です。

ここは日本語のテンプレートを置く面であり、構造や仕様の正本は持ちません。必要な技術的根拠は `blueprint/` や各実装の正本を参照します。

### `governance/guides/`

人間向けの進め方、運用手順、文書の使い分けを集約する場所です。

反復して参照する日本語の案内や how-to はここへ寄せ、案件固有の一時文脈は `governance/tasks/` に置きます。

### `platform/`

恒久的な運用基盤の参照置き場です。現在は `platform/observability/` が主要対象です。

Collector、Loki、Tempo、Prometheus、Promtail、Grafana などの観測基盤構成はここに置きます。

### `deployments/`

ローカル開発や移行で複数要素をまとめて起動するための実行ラッパです。設計や仕様の正本は置きません。

### `data/`

時系列ストア向けの SQL 資産の正本です。DDL、migration、query、TimescaleDB 向け操作は `data/tsdb/` に集約します。

### `scenarios/`

代表ユースケースと検証導線です。どの入口を推奨するか、何を確認すべきかを日本語で示します。

### `.agents/skills/`

このリポジトリ専用の Codex skills を置く場所です。

変更分類、blueprint から実装への落とし込み、scenario ベースの検証のような反復作業をここで定義します。

## 開発時の変更判断

- 構造、境界、継承ルールが変わる: `blueprint/` を先に更新
- 文書配置や変更ルールが変わる: `governance/policies/` を更新
- 代表導線や検証手順が変わる: `scenarios/` を更新
- 実装内部だけの変更: 対象の `implementations/` を更新
- 観測基盤構成の変更: `platform/observability/` を更新
- ローカル compose や migrator ラッパの変更: `deployments/` を更新
- SQL 資産の変更: `data/tsdb/` を更新

詳細ルールは `governance/policies/change-policy.md` を参照してください。

## ローカル開発の最短導線

前提:

- Docker / Docker Compose
- Go
- Python

基本導線:

1. `make dev-up`
2. `data/tsdb/README.md` に従って migration を適用
3. `implementations/dag-core/` と `implementations/worker-py/` を起点に必要な開発コマンドを実行
4. `scenarios/representative-scenario.md` に沿って疎通確認

代表コマンドは `Makefile` に集約しています。詳細な運用や実装固有コマンドは各 README を参照してください。

## 実装と仕様の正本

- 構造原則の正本: `blueprint/`
- API 実装の正本: `implementations/dag-core/`
- SQL 設計の正本: `data/tsdb/`
- 人間向けの要求・報告テンプレート: `governance/contracts/`
- 人間向けの進め方と手順の集約: `governance/guides/`
- 一時的な共有作業文脈: `governance/tasks/`
- repository-scoped Codex skills: `.agents/skills/`
- ローカル専用の生成物: `var/`

`README.md` 自体は要約と導線のための文書であり、詳細仕様の正本ではありません。

## 代表的な参照先

- 親リポジトリの位置づけ: `blueprint/identity/repository-purpose.md`
- 構造境界: `blueprint/architecture/system-boundaries.md`
- フローと時間モデル: `blueprint/architecture/flow-and-time-model.md`
- 実装原則: `blueprint/conventions/implementation-principles.md`
- 観測性原則: `blueprint/conventions/observability-conventions.md`
- データ原則: `blueprint/conventions/data-conventions.md`
- 代表シナリオ: `scenarios/representative-scenario.md`
- Go 実装: `implementations/dag-core/README.md`
- Python worker 実装: `implementations/worker-py/README.md`
- 観測基盤: `platform/observability/README.md`
- ローカル実行ラッパ: `deployments/README.md`
