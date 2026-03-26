# dag-observatory

Clock＋DAG実行基盤を持つ Golang + Python SaaS テンプレート。
「観測可能性設計を前提としたシステムアーキテクチャ」の実験場。

---

## 1. プロジェクト概要
- Clock + DAG 実行基盤と観測可能性設計の同居を検証するためのリポジトリ
- SaaS バックエンドの非同期処理・ワークフロー基盤を想定
- OTel（意図ログ/トレース/メトリクス）と Promtail（環境ログ）の併用を前提に設計

---

## 2. 解決したい課題
- 非同期処理のブラックボックス化（因果が追えない）
- 実運用でのデバッグ導線が設計されていない
- 観測が後付けになり、ログ粒度やラベル設計が崩れる

---

## 3. コアアーキテクチャ概要
### 3.1 コンポーネント
- Go API / DAG Runtime: `dag-core`
- Python Worker: `worker-py`
- 認可実験: `auth-n-z`
- 観測/基盤コンテナ: `deployments`

### 3.2 Clock + DAG 実行モデル
- 外部入力（Clock/Event）を起点に DAG を起動
- Runner/Driver が Task を発行し、Worker が実行結果を返す
- Store/Txn により Cycle 単位の整合性を担保

### 3.3 参考ドキュメント
- `.docs/report/p2_抽象Clock+DAG/00_概要.md`
- `.docs/report/p2_抽象Clock+DAG/10_アーキテクチャ.md`
- `.docs/report/p2_抽象Clock+DAG/50_Runner_Driver_EventStream.md`

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
- `.docs/report/p1_観測可能性/意図ログと環境ログについて.md`
- `.docs/report/p1_観測可能性/OTelの構成について.md`
- `.docs/report/p1_観測可能性/メトリクス一覧.md`
- `.docs/spec/30_intent_events.md`

---

## 5. Python Worker 連携
### 5.1 Kafka トピック
- tasks: `dagruntime-tasks`
- events: `dagruntime-events`

### 5.2 Trace context 伝播
- Kafka headers で W3C trace context を inject/extract
- Go/Python 両方で span を作成

### 5.3 参考ドキュメント
- `.docs/report/p3_Worker(Python)/5_実装状況まとめ.md`
- `.docs/report/p3_Worker(Python)/3_TraceContext伝播案.md`

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
- Go（`dag-core` 用）
- Python（`worker-py` 用）

### 6.2 起動・停止
- `make dev-up` でローカルスタック起動
- `make dev-down` で停止

#### 追加サービス（Compose）
- TimescaleDB / MinIO / worker-py-outbox / worker-py-gc が `deployments/compose/docker-compose.app.dev.yml` に追加済み
- マイグレーションは `docs/tsdb/README.md` の手順で実行

### 6.3 Go 側の開発コマンド
- `make go-mod-download` / `make go-mod-tidy`
- `make go-fmt` / `make go-vet` / `make go-test`
- `make go-build`
 - `make sqlc-gen`

---

## 7. 仕様管理ポリシー
### 7.1 仕様の正本（Source of Truth）
- API 仕様の正本: `dag-core/internal/interface/http/router.go` と handler/dto 実装
- DB 仕様の正本: `docs/tsdb/schema/migrate/*.sql`
- `README.md` は概要と導線を提供する要約ドキュメント（正本ではない）

### 7.2 変更フロー（開発）
1. API 変更時は `dag-core` 実装を先に更新する
2. DB 変更時は `docs/tsdb/schema/migrate` に forward-only な migration を追加する
3. OpenAPI / sqlc などの生成物を更新する
4. `go test ./...` で回帰確認する
5. `README.md` と関連ドキュメントを更新し、導線を揃える

### 7.3 起動・運用フロー
1. 基盤コンテナ（DB/MinIO/メッセージ基盤）を起動する
2. `TSDB_URL` を設定し、migrator を実行する
3. `dag-core` と worker 群を起動する
4. `/healthz` と主要 API を疎通確認する

### 7.4 受け入れ条件（DoD）
- 実装と OpenAPI の API 差分がない
- migration / query / sqlc 設定の参照パスが一致している
- `go test ./...` が成功する
- `docs/tsdb/README.md` の手順でセットアップ再現できる

---

## 8. リポジトリ構成
- `dag-core`: Go API / DAG Runtime
- `worker-py`: Python Worker
- `deployments`: Grafana / Loki / Tempo / OTel Collector / Promtail など
- `configs`: ローカル設定
- `scripts`: 開発用スクリプト
- `.docs`: 設計・運用ドキュメント

---

## 9. ドキュメントガイド
### 9.1 観測可能性
- `.docs/report/p1_観測可能性/OTelの構成について.md`
- `.docs/report/p1_観測可能性/OTelの運用手順.md`
- `.docs/report/p1_観測可能性/スモークテスト手順.md`
- `.docs/report/p1_観測可能性/環境変数一覧.md`

### 9.2 DAG Runtime（抽象 Clock + DAG）
- `.docs/report/p2_抽象Clock+DAG/00_概要.md`
- `.docs/report/p2_抽象Clock+DAG/10_アーキテクチャ.md`
- `.docs/report/p2_抽象Clock+DAG/20_Workflow_Compile.md`
- `.docs/report/p2_抽象Clock+DAG/40_Node_ExecutionSpec_Txn.md`

### 9.3 Python Worker
- `.docs/report/p3_Worker(Python)/1_ワーカー追加案.md`
- `.docs/report/p3_Worker(Python)/5_実装状況まとめ.md`

---

## 10. 今後の拡張予定
- Python 側の OTLP exporter 追加
- task_name 拡張時の registry 分割
- error payload の code 規約の厳密化
