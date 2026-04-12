# Run Inspection と Backtest の統合シナリオ

## 目的

この文書は、次の確認を 1 つの再現可能な流れに統合したシナリオです。

- run inspection
- normal / skip / fail / retry の確認
- 実データバックテスト比較
- minimum UI の確認
- JSONL replay の確認

`implementations/dag-core` の run observability を確認するための昇格済みシナリオとして扱います。

## 前提条件

- ローカルスタックが起動していること: `make dev-up`
- `dag-core-api` が起動していること
- 対象 range の timeframe bar が存在すること
- JSONL replay を確認する場合は `DAGRUNTIME_OBSERVATION_JSONL_PATH=<path>` を有効にしていること

### 最小データ準備

```bash
cd /home/user/shiq/dag-observatory
make tsdb-seed
```

seed には次が含まれます。

- `USDJPY` symbol master
- `2026-04-01T00:00:00Z` から `2026-04-01T05:59:00Z` までの `m1` timeframe bar

## 検証手順

### 1. 比較可能な backtest run を 2 本作る

次を実行します。

- `POST /algotrade/backtests:run` を base 条件で実行
- `POST /algotrade/backtests:run` を changed cost または parameter 条件で実行

保持する値:

- `run_id`
- `partition`
- `workflow_name`
- `workflow_version`
- `parameter_set_id`
- `source`
- `timezone`
- `gap_handling`

期待結果:

- 比較可能な 2 つの run が揃う
- run ごとの metadata を後続 API で追える

### 2. run list / detail / step 導線を確認する

次を確認します。

- `GET /runs?partition=<partition>&limit=200`
- `GET /runs/<run_id>`
- `GET /runs/<run_id>/steps`
- `GET /runs/<run_id>/steps/<sequence_no>`

期待結果:

- step order が `sequence_no` で読める
- node detail に correlation ID が含まれる
  - `partition`
  - `intent_id`
  - `execution_id`
  - `trade_id`

### 3. compare と visualization API を確認する

次を確認します。

- `GET /runs/compare?base_run_id=<base>&target_run_id=<target>`
- `GET /algotrade/trades?run_id=<base>&limit=100&offset=0`
- `GET /algotrade/equity?run_id=<base>`
- `GET /runs/<run_id>/backtest-summary`

期待結果:

- compare に summary delta と changed field が出る
- trades と equity が run scope で再現可能に読める

### 4. UI 導線を bounded fetch 前提で確認する

`/runs/ui?partition=<partition>&run_id=<base>&target_run_id=<target>` を開きます。

期待結果:

- compare selector と diff summary が見える
- UI が bounded run fetch と bounded step rendering を前提に動いている

### 5. normal / skip / fail / retry を確認する

少なくとも次を含む run pair または run set を確認します。

- normal: target が概ね成功している
- skip: skip reason が step や compare delta に現れる
- fail: failed step が反映される
- retry: retry count の差分が出る

具体的な確認項目:

- `summary_delta.failed_step_delta`
- `summary_delta.skipped_step_delta`
- `summary_delta.retry_count_delta`

### 6. JSONL replay を確認する

JSONL replay を有効にしている場合は次を行います。

1. run 記録後に API を停止する
2. 同じ `DAGRUNTIME_OBSERVATION_JSONL_PATH` で API を再起動する
3. 次を呼ぶ
   - `GET /runs?partition=<partition>`
   - `GET /healthz`

期待結果:

- 再起動後も recorded run と step を読み出せる
- `/healthz.observation_replay` に次の counter が出る
  - `skipped_lines`
  - `errors`

## 参考コマンド

```bash
curl -s "http://localhost:8080/runs?partition=bt:USDJPY:m1:v1&limit=200" | jq
curl -s "http://localhost:8080/runs/compare?base_run_id=bt-base-001&target_run_id=bt-target-001" | jq
curl -s "http://localhost:8080/algotrade/trades?run_id=bt-base-001&limit=100&offset=0" | jq
curl -s "http://localhost:8080/algotrade/equity?run_id=bt-base-001" | jq
curl -s "http://localhost:8080/healthz" | jq
```

## 観測確認点

- run list、detail、steps が partition と run_id で安定して読めること
- compare delta と visualization API が整合していること
- skip / fail / retry の差分が summary と steps で追えること
- JSONL replay 有効時に再起動後も run observability を再構成できること

## 更新条件

次のいずれかが変わったときに更新します。

- 統合 run inspection 導線
- compare / visualization API の最小確認面
- JSONL replay の確認手順
