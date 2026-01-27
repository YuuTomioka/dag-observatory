# dag-observatory

dag-observatory は、Clock＋DAG 実行基盤を題材に「意図ログ・トレース・メトリクス（OTel）× 環境ログ（Promtail）」の混在設計を実運用目線で検証できるテンプレートです。

## Loki 検索キーの標準

Promtail と OTel 意図ログを同じ検索体験に寄せるため、Loki のラベルキーは以下で統一します。

- 標準ラベル: `service`, `env`, `app`, `instance`, `job`
- 例: `{service="dag-observatory-demo-go", env="dev", app="dag-observatory-demo-go"}`

### 変換ルール

- OTel → Loki
  - resource attributes を `service/env/app/instance/job` に正規化して Loki exporter で labels 化
- Promtail → Loki
  - `relabel_configs` で `job` を基準に `service`/`app` を作成し、`__path__` から `instance` を生成
