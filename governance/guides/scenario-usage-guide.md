# シナリオ利用ガイド

## 対象読者

- contributor
- operator
- reviewer

## 使う場面

- どの scenario を使って確認するか迷うとき
- 実装や修正の結果を代表導線で確認したいとき
- シナリオの gap をどこへ返すか判断したいとき

## 使い方

1. 変更対象に最も近い scenario を選ぶ
2. 前提条件を確認する
3. 手順どおりに実行する
4. 期待結果と観測確認点を照合する
5. gap がある場合は `governance/tasks/` に記録する

## 現在の主な導線

- `scenarios/representative-scenario.md`: repository-wide の代表導線
- `scenarios/realdata-backtest-compare-scenario.md`: 実データ backtest compare
- `scenarios/run-inspection-backtest-integrated-scenario.md`: run inspection と backtest の統合確認

## gap の扱い

- 一時的な不足や未整備事項は `governance/tasks/` に置く
- 反復して必要な手順になったら `scenarios/` に昇格する
- シナリオの意味自体が変わる場合は必要に応じて `blueprint/` や `governance/policies/` を先に更新する

## 注意点

- `scenarios/` は開発プロセス全体の説明ではなく、検証導線の面
- 仕様の正本をここで再定義しない
