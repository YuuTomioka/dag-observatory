# Intent Events

意図ログをドメインイベントとして仕様化するための一覧です。

| Event | 必須属性 | 用途 |
| --- | --- | --- |
| clock.tick.received | clock.utc, symbol | Clock入力の受信確認と外部入力の起点 |
| dag.run.started | dag.run_id, symbol | DAG実行開始の起点 |
| dag.run.finished | dag.run_id, status, duration_ms | DAG実行の完了と所要時間の記録 |
| dag.node.started | dag.run_id, dag.node_id | DAGノード実行開始 |
| dag.node.finished | dag.run_id, dag.node_id, status, duration_ms | DAGノード実行完了と結果 |
