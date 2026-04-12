# 変更要求とハンドオフのガイド

## 対象読者

- contributor
- reviewer
- repository maintainer

## 使う場面

- 変更要求を受けたとき
- 実装へ渡す前に要求を整理したいとき
- 検証結果を引き継ぎたいとき

## 基本の流れ

1. 要求を受ける
2. 変更分類を行う
3. 必要なら上位レイヤの文書を先に更新する
4. `governance/contracts/change-request.md` に要求を整理する
5. 実装作業へ進む場合は `governance/contracts/implementation-task.md` に落とす
6. 検証後に `governance/contracts/validation-report.md` に結果を残す

## 使い分け

### `change-request`

- 何を変えたいか
- なぜ必要か
- どの領域が関係するか

### `implementation-task`

- 実装方針
- 対象範囲
- 完了条件
- 検証手順

### `validation-report`

- 実施した確認
- 結果
- 差分
- 残リスク

## 注意点

- 人間向け文書で仕様の正本を書き換えない
- 変更が `blueprint/` や `governance/policies/` に波及するなら先に上位を更新する
- 一案件固有の追加メモは `governance/tasks/` に置く

## 参照先

- `governance/contracts/README.md`
- `governance/policies/change-policy.md`
