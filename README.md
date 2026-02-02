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

### 6.3 Go 側の開発コマンド
- `make go-mod-download` / `make go-mod-tidy`
- `make go-fmt` / `make go-vet` / `make go-test`
- `make go-build`

---

## 7. リポジトリ構成
- `dag-core`: Go API / DAG Runtime
- `worker-py`: Python Worker
- `deployments`: Grafana / Loki / Tempo / OTel Collector / Promtail など
- `configs`: ローカル設定
- `scripts`: 開発用スクリプト
- `.docs`: 設計・運用ドキュメント

---

## 8. ドキュメントガイド
### 8.1 観測可能性
- `.docs/report/p1_観測可能性/OTelの構成について.md`
- `.docs/report/p1_観測可能性/OTelの運用手順.md`
- `.docs/report/p1_観測可能性/スモークテスト手順.md`
- `.docs/report/p1_観測可能性/環境変数一覧.md`

### 8.2 DAG Runtime（抽象 Clock + DAG）
- `.docs/report/p2_抽象Clock+DAG/00_概要.md`
- `.docs/report/p2_抽象Clock+DAG/10_アーキテクチャ.md`
- `.docs/report/p2_抽象Clock+DAG/20_Workflow_Compile.md`
- `.docs/report/p2_抽象Clock+DAG/40_Node_ExecutionSpec_Txn.md`

### 8.3 Python Worker
- `.docs/report/p3_Worker(Python)/1_ワーカー追加案.md`
- `.docs/report/p3_Worker(Python)/5_実装状況まとめ.md`

---

## 9. 今後の拡張予定
- Python 側の OTLP exporter 追加
- task_name 拡張時の registry 分割
- error payload の code 規約の厳密化
