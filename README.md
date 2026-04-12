# dag-observatory

Clock＋DAG 実行基盤と時系列設計を持つ親リポジトリ。

このリポジトリは単一プロダクトではなく、DAG + 時系列システムの派生プロジェクト向け設計テンプレートとして再編中です。

## 入口

- 親リポジトリの位置づけ: `blueprint/identity/repository-purpose.md`
- 構造境界: `blueprint/architecture/system-boundaries.md`
- フローと時間モデル: `blueprint/architecture/flow-and-time-model.md`
- 実装原則: `blueprint/conventions/implementation-principles.md`
- 観測性原則: `blueprint/conventions/observability-conventions.md`
- データ原則: `blueprint/conventions/data-conventions.md`
- 代表シナリオ: `scenarios/representative-scenario.md`
- Codex 向け総則: `AGENTS.md`

---

## Root 直下の責務
- `AGENTS.md`: AI ペアプロ向けの読解順と変更原則
- `README.md`: 人間向けの総合入口と全体導線
- `Makefile`: 共通操作の実行入口
- `blueprint/`: 親リポジトリとしての設計原則の正本
- `governance/`: 変更ルールと作業文脈の管理
- `implementations/`: 参照実装の正本
- `platform/`: 環境と運用基盤の恒久的リファレンス
- `deployments/`: ローカル実行や移行のための実行ラッパ
- `data/`: 時系列・分析向けデータ設計資産
- `scenarios/`: 代表ユースケースの入口
- `scripts/`: 補助的な実行・保守スクリプト
- `.codex/`: AI ローカルな作業補助メモとチェックリスト

この構造は次の 3 層として読むと扱いやすいです。

- 上位原則層: `blueprint/`, `governance/`, `AGENTS.md`, `README.md`
- 実装・資産層: `implementations/`, `platform/`, `data/`
- 実行・利用層: `deployments/`, `scripts/`, `scenarios/`

---

## 1. プロジェクト概要
- Clock + DAG 実行基盤と観測可能性設計の同居を検証するためのリポジトリ
- SaaS バックエンドの非同期処理・ワークフロー基盤を想定
- OTel（意図ログ/トレース/メトリクス）と Promtail（環境ログ）の併用を前提に設計
- 親リポジトリとしては、実装より先に構造原則を保持する

---

## 2. 解決したい課題
- 非同期処理のブラックボックス化（因果が追えない）
- 実運用でのデバッグ導線が設計されていない
- 観測が後付けになり、ログ粒度やラベル設計が崩れる

---

## 3. コアアーキテクチャ概要
### 3.1 コンポーネント
- Go 参照実装: `implementations/dag-core`
- Python Worker 参照実装: `implementations/worker-py`
- 観測/基盤参照構成: `deployments`

注記:

- 参照実装の物理正本は `implementations/` 配下に統一する
- 判断根拠は `blueprint/adr/ADR-0003-implementation-directories-under-implementations.md`

### 3.2 Clock + DAG 実行モデル
- 外部入力（Clock/Event）を起点に DAG を起動
- Runner/Driver が Task を発行し、Worker が実行結果を返す
- Store/Txn により Cycle 単位の整合性を担保

### 3.3 参考ドキュメント
- `blueprint/architecture/system-boundaries.md`
- `blueprint/architecture/flow-and-time-model.md`
- `blueprint/conventions/implementation-principles.md`

---

## 4. 観測可能性設計（このリポジトリの主役）
### 4.1 意図ログ（OTel Logs）と環境ログ（Promtail）
- 意図ログ: ドメインイベントを OTel Logs として出力
- 環境ログ: コンテナ/OS の稼働ログを Promtail で収集
- Loki 上でラベルを統一して検索体験を揃える

### 4.2 Loki ラベル標準
- `service`, `env`, `app`, `instance`, `job`
- OTel → Loki は `resource/loki_labels` で正規化
- Promtail → Loki は `relabel_configs` で統一

### 4.3 トレース連携
- Intent Logs に `trace_id` / `span_id` を付与
- Tempo ↔ Loki を相互参照する導線を用意

### 4.4 参考ドキュメント
- `blueprint/conventions/observability-conventions.md`
- `platform/observability/README.md`
- `blueprint/architecture/flow-and-time-model.md`

---

## 5. Python Worker 連携
### 5.1 Kafka トピック
- tasks: `dagruntime-tasks`
- events: `dagruntime-events`

### 5.2 Trace context 伝播
- Kafka headers で W3C trace context を inject/extract
- Go/Python 両方で span を作成

### 5.3 参考ドキュメント
- `implementations/worker-py/README.md`
- `blueprint/conventions/observability-conventions.md`

---

## 5.4 RunView Phase（正規化仕様）
RunView は event.Type から phase を正規化します。

### 許容値
- `in_progress`
- `succeeded`
- `failed`
- `cancelled`
- `unknown`

### マッピング規則
- `*.requested` / `*.started` / `*.running` → `in_progress`
- `*.completed` / `*.succeeded` → `succeeded`
- `*.failed` / `*.error` → `failed`
- `*.cancelled` / `*.canceled` → `cancelled`
- それ以外 → `unknown`

---

## 6. ローカル開発セットアップ
### 6.1 必要ツール
- Docker / Docker Compose
- Go（`implementations/dag-core` 用）
- Python（`implementations/worker-py` 用）

### 6.2 起動・停止
- `make dev-up` でローカルスタック起動
- `make dev-down` で停止

#### 追加サービス（Compose）
- TimescaleDB / MinIO / worker-py-outbox / worker-py-gc が `deployments/compose/docker-compose.app.dev.yml` に追加済み
- 観測基盤の compose は `platform/observability/compose/docker-compose.observability.yml` に配置
- マイグレーションは `data/tsdb/README.md` の手順で実行

### 6.3 Go 側の開発コマンド
- `make go-mod-download` / `make go-mod-tidy`
- `make go-fmt` / `make go-vet` / `make go-test`
- `make go-build`
- `make sqlc-gen`
- `make contracts-check`（OpenAPI / gRPC / GraphQL の生成差分チェック）

---

## 7. 仕様管理ポリシー
### 7.1 仕様の正本（Source of Truth）
- API 仕様の正本: `implementations/dag-core/internal/interface/http/router.go` と handler/dto 実装
- DB 仕様の正本: `data/tsdb/` 配下の SQL 資産
- `README.md` は概要と導線を提供する要約ドキュメント（正本ではない）

生成された OpenAPI は `implementations/dag-core/openapi/` に置く。

### 7.2 変更フロー（開発）
1. API 変更時は `implementations/dag-core` 実装を先に更新する
2. DB 変更時は `data/tsdb/` 配下の migration 資産を更新する
3. OpenAPI / sqlc などの生成物を更新する
4. `make go-test` で回帰確認する
5. `make contracts-check` で契約生成物の差分がないことを確認する
6. `README.md` と関連ドキュメントを更新し、導線を揃える

### 7.3 起動・運用フロー
1. 基盤コンテナ（DB/MinIO/メッセージ基盤）を起動する
2. `TSDB_URL` を設定し、migrator を実行する
3. `implementations/dag-core` と `implementations/worker-py` を起点に core / worker 群を起動する
4. `/healthz` と主要 API を疎通確認する

### 7.4 受け入れ条件（DoD）
- 実装と OpenAPI の API 差分がない
- migration / query / sqlc 設定の参照パスが一致している
- `make go-test` が成功する
- `make contracts-check` が成功する
- `data/tsdb/README.md` の手順でセットアップ再現できる

### 7.5 CI ゲート
- PR では `contracts-check` workflow を必須とし、以下を検証する
- `make go-test`
- `make contracts-check`

### 7.6 DB アクセス方針（現時点）
- `data/tsdb/` 配下の schema/migrate を DDL の正本、query を SQL 仕様の正本とする
- `db_backups` の query 定義は `data/tsdb/query/` 配下で管理する
- `implementations/dag-core` 側の実装が inline SQL でも、仕様変更時は先に `data/tsdb/query/` 相当の SQL 資産を更新する

### 7.7 OpenAPI / Swagger UI 運用方針（採用）
- OpenAPI の正本は `implementations/dag-core` の HTTP 実装と注釈。
- 生成物は `implementations/dag-core/openapi/` に置き、手編集しない。
- API 変更時は `make openapi` と `make contracts-check` を必須運用とする。
- Swagger UI の標準ルートは `/swagger/*` とする。
- 環境別公開方針:
  - `dev`: デフォルト有効
  - `staging`: アクセス制御下で有効
  - `prod`: デフォルト無効（明示承認時のみ一時有効化）
- 詳細運用は `implementations/dag-core/README.md` の OpenAPI/Swagger UI ポリシーに従う。

---

## 8. リポジトリ構成
- `blueprint`: 恒久的な設計原則
- `implementations`: 参照実装の受け先
- `platform`: 実行環境・運用基盤の受け先
- `data`: 時系列・分析設計資産の受け先
- `scenarios`: 代表導線
- `governance`: 変更制御
- `implementations/dag-core`: Go 参照実装
- `implementations/worker-py`: Python 参照実装
- `deployments`: ローカル実行用の compose / container wrapper

実装ディレクトリ移行方針は `blueprint/adr/ADR-0003-implementation-directories-under-implementations.md` に固定している。

`deployments` は恒久的な設計や platform/source-of-truth の置き場ではなく、現時点では app dev と migrator を起動するための実行ラッパとして扱う。

---

## 9. ドキュメントガイド
### 9.1 観測可能性
- `blueprint/conventions/observability-conventions.md`
- `platform/observability/README.md`
- `scenarios/representative-scenario.md`

### 9.2 DAG Runtime（抽象 Clock + DAG）
- `blueprint/architecture/system-boundaries.md`
- `blueprint/architecture/flow-and-time-model.md`
- `blueprint/conventions/implementation-principles.md`

### 9.3 Python Worker
- `implementations/worker-py/README.md`
- `blueprint/conventions/observability-conventions.md`

---

## 10. 今後の拡張予定
- Python 側の OTLP exporter 追加
- task_name 拡張時の registry 分割
- error payload の code 規約の厳密化
