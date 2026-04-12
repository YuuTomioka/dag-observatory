# 代表シナリオ

## 目的

この文書は、このリポジトリにおける代表的な end-to-end 検証導線を定義します。

smoke test として実行できる具体性を持ちながら、派生先でも再利用しやすい形を保つことを目的とします。

## 対象フロー

現在の代表フローは、**marketdata timeframe-bar backfill を workflow 実行する流れ**です。

流れの全体像:

1. 呼び出し元が backfill リクエストを送る
2. API がその要求を runtime event に変換する
3. 非同期実行のための task request が publish される
4. worker 側の処理が result event を発行する
5. runtime が result event を消費して state を進める
6. telemetry と時系列データにより進行状況が観測可能になる

このシナリオは、実行の正しさと observability の相関確認を同時に検証します。

関連シナリオ:

- `scenarios/realdata-backtest-compare-scenario.md`
- `scenarios/run-inspection-backtest-integrated-scenario.md`

## 前提条件

- ローカルスタックが起動していること: `make dev-up`
- migration と schema の適用が完了していること: `data/tsdb/README.md`
- API と worker が正常に起動していること
- observability サービス群が到達可能であること
  - collector
  - Loki
  - Tempo
  - Grafana

## 推奨エントリポイント

feature-local な endpoint と generic runtime endpoint が共存する場合は、次の優先順で使います。

1. 即時にドメイン結果を得るための feature-local direct endpoint
2. namespaced workflow 実行のための feature-local workflow endpoint
3. feature-local endpoint が無い場合、または runtime-level operation が意図されている場合のみ generic runtime endpoint

marketdata timeframe-bar backfill に対する対応:

- primary direct entrypoint: `POST /marketdata/timeframe-bars:backfill`
- primary workflow entrypoint: `POST /marketdata/timeframe-bars:backfill-workflow`
- fallback runtime entrypoint: `POST /dag/run`

## 検証手順

### 1. workflow backfill request を送る

`POST /marketdata/timeframe-bars:backfill-workflow` を使い、時間範囲を bounded にし、mode を明示して送ります。

期待結果:

- response に run identifier が含まれる
- 重い処理を同期実行せずに request が受理される

### 2. 非同期 task/event の進行を確認する

task request と result event が runtime model に沿って進行することを確認します。

代表的な event progression:

- `task.requested`
- `task.completed` または `task.failed`

期待結果:

- consumed result event に応じて run state が一貫して進む

### 3. データ書き込み結果を確認する

指定した symbol、timeframe、range に対する timeframe-bar 出力が書き込まれたことを確認します。

期待結果:

- 時系列ストアに durable record が存在する
- write shape が指定 mode の意味論と一致する

### 4. observability の相関を確認する

run と trace context によって cross-signal correlation が取れることを確認します。

確認点:

- intent log を event type と `dag.run_id` で検索できる
- log の相関フィールド `trace_id`, `span_id` から trace へ辿れる
- error や timeout が trace と log 上で明示的な failure semantic として現れる
- 関連する HTTP metric と runtime metric が増加する

### 5. artifact/outbox の整合を必要時に確認する

artifact を出力する workflow の場合は、次も確認します。

- worker 完了後に artifact metadata が書かれる
- outbox publication が database confirmation の後に進む
- artifact storage を有効にしている場合は presigned retrieval flow が利用できる

## 観測確認点

- 非同期 task/event flow が end-to-end で観測できること
- data-side effect が run と対応づけて確認できること
- log、trace、metric が run context で相互に辿れること
- failure semantic が明示的で検索可能であること

## 受け入れ条件

次がすべて満たされた場合、この代表シナリオは検証済みとみなします。

- 推奨エントリポイントの優先順に従っている
- 非同期 task/event flow を end-to-end で観測できる
- 指定 range に対する data-side effect を確認できる
- log、trace、metric が run context で相互参照できる
- failure semantic が明示的で検索できる

## 補足

時間経過とともに次の execution outcome をカバーできるのが望ましいです。

- success
- failure
- timeout
- skip

安定して再現できないものがある場合は、`governance/tasks/` に gap を記録し、この文書は安定した目標形として保ちます。

## 更新条件

次のいずれかが変わったときに更新します。

- 推奨エントリフロー
- このリポジトリの代表ユースケース
- 運用上の検証手順

構造変更は `blueprint/`、ポリシー変更は `governance/policies/` を先に更新します。
