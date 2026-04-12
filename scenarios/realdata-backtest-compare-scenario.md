# 実データバックテスト比較シナリオ

## 目的

この文書は、次の流れを再現可能な形で検証するためのシナリオです。

- 実データの timeframe-bar を使った period backtest 実行
- 2 つの backtest run の比較
- minimum UI を使った run 差分の確認

normal、skip、fail、retry、JSONL replay まで含む統合導線は、`scenarios/run-inspection-backtest-integrated-scenario.md` を使います。

## 前提条件

- ローカルスタックが起動していること: `make dev-up`
- TSDB migration が適用済みであること: `data/tsdb/README.md`
- `dag-core-api` が marketdata repository を利用できる設定で起動していること
- 対象 symbol、timeframe、range に対する historical timeframe bar が存在すること

### 最小 fixture 準備

seed SQL を適用します。`USDJPY` と `2026-04-01` の M1 bar を含みます。

```bash
cd /home/user/shiq/dag-observatory
make tsdb-seed
```

`strict` / `skip` の gap behavior を確認したい場合は、必要に応じて次を実行します。

```sql
DELETE FROM timeframe_bar
WHERE symbol_id = (SELECT id FROM symbol WHERE code = 'USDJPY')
  AND timeframe_code = 'm1'
  AND open_time >= '2026-04-01T02:00:00Z'::timestamptz
  AND open_time <  '2026-04-01T02:10:00Z'::timestamptz;
```

## 入力メタデータ方針

各 backtest request には、range や symbol に加えて再現性のための metadata を含めます。

- `workflow_name`
- `workflow_version`
- `parameter_set_id`
- `source`
- `timezone`
- `gap_handling`

これらは backtest API response にも返るため、検証メモに保持します。

## 検証手順

### 1. base backtest を実行する

`POST /algotrade/backtests:run` を呼びます。

必須項目:

- `partition`
- `symbol_code`
- `timeframe_code`
- `from`
- `to`

推奨 metadata:

- `run_id`
- `workflow_name`
- `workflow_version`
- `parameter_set_id`
- `source`
- `timezone`
- `gap_handling`
- `spread_bps`
- `fee_bps`
- `slippage_bps`
- `min_lot`

期待結果:

- HTTP `200` または `202`
- response に base `run_id` が含まれる
- response に `cycle_count` と `window_size_bars` が含まれる

### 2. 条件を変えた target backtest を実行する

同じ endpoint を使い、次のいずれかを変更して実行します。

- 実行前提: `fee_bps`, `slippage_bps`, `spread_bps`, `min_lot`
- パラメータ metadata: `parameter_set_id`, `workflow_version`

期待結果:

- 2 本目の run が別の `run_id` を持つ
- 2 つの run を run read API から参照できる

### 3. run compare API を確認する

`GET /runs/compare?base_run_id=<base>&target_run_id=<target>` を呼びます。

期待結果:

- `summary_delta` が存在する
- `state_field_deltas` が存在する
- 少なくとも 1 つ以上の changed field が見える

### 4. minimum UI で確認する

`GET /runs/ui?run_id=<base>&target_run_id=<target>` を開きます。

期待結果:

- run timeline を確認できる
- compare target selector が表示される
- detail pane に `Run Compare` section が出る

### 5. summary と step 詳細を直接確認する

次を呼びます。

- `GET /runs/<run_id>/steps`
- `GET /runs/<run_id>/steps/<sequence_no>`
- `GET /algotrade/summary?partition=<partition>`

期待結果:

- step の順序と status を確認できる
- summary が compare output の方向と整合している

### 6. success / failure / skip / retry の signal を確認する

少なくとも次を含む比較対象を用意します。

- mostly-success な run pair
- target 側に failure または skip を含む pair
- retry count が異なる pair

具体的な確認項目:

- `GET /runs/compare?...` に次が含まれる
  - `summary_delta.failed_step_delta`
  - `summary_delta.skipped_step_delta`
  - `summary_delta.succeeded_step_delta`
  - `summary_delta.retry_count_delta`
- assumption や parameter set が異なる場合、`state_field_deltas` に次のような changed field が含まれる
  - `strategy_summary.trade_count`
  - `strategy_summary.total_net_pnl`

## 参考コマンド

```bash
# base
curl -s -X POST http://localhost:8080/algotrade/backtests:run \
  -H 'Content-Type: application/json' \
  -d '{
    "partition":"bt:USDJPY:m1:v1",
    "run_id":"bt-base-001",
    "symbol_code":"USDJPY",
    "timeframe_code":"m1",
    "from":"2026-04-01T00:00:00Z",
    "to":"2026-04-01T06:00:00Z",
    "workflow_name":"breakout_long",
    "workflow_version":"v1",
    "parameter_set_id":"pset-baseline",
    "source":"tsdb.timeframe_bars.v1",
    "timezone":"UTC",
    "gap_handling":"strict",
    "spread_bps":2.0,
    "fee_bps":0.5,
    "slippage_bps":0.8,
    "min_lot":0.01
  }' | jq

# target
curl -s -X POST http://localhost:8080/algotrade/backtests:run \
  -H 'Content-Type: application/json' \
  -d '{
    "partition":"bt:USDJPY:m1:v1",
    "run_id":"bt-target-001",
    "symbol_code":"USDJPY",
    "timeframe_code":"m1",
    "from":"2026-04-01T00:00:00Z",
    "to":"2026-04-01T06:00:00Z",
    "workflow_name":"breakout_long",
    "workflow_version":"v1",
    "parameter_set_id":"pset-higher-cost",
    "source":"tsdb.timeframe_bars.v1",
    "timezone":"UTC",
    "gap_handling":"strict",
    "spread_bps":2.0,
    "fee_bps":1.2,
    "slippage_bps":1.6,
    "min_lot":0.01
  }' | jq

# compare
curl -s "http://localhost:8080/runs/compare?base_run_id=bt-base-001&target_run_id=bt-target-001" | jq
```

## 観測確認点

- compare API の差分と UI 上の表示が整合していること
- run ごとの step と summary が再現可能に読めること
- success / failure / skip / retry の差分が API 上で確認できること

## 更新条件

次のいずれかが変わったときに更新します。

- backtest compare の推奨導線
- 最低限確認すべき API
- 再現に必要な fixture や metadata 方針
