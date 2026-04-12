# レビュー運用ガイド

## 対象読者

- reviewer
- contributor

## 使う場面

- 変更内容をどうレビューすべきか整理したいとき
- `validation-report` と `/review` の使い分けを確認したいとき

## 基本方針

- 現時点では repository-scoped の `governance/reviews/` は持たない
- 既存の review flow と `governance/contracts/validation-report.md` を組み合わせて使う

## 使い分け

### `/review`

- コードレビュー
- 変更差分の問題点洗い出し
- 実装上のリスク確認

### `validation-report`

- 実施した検証の記録
- 結果と差分の共有
- 残リスクや follow-up の明文化

## 典型的な流れ

1. 実装を行う
2. scenario や test で検証する
3. `validation-report` に結果をまとめる
4. `/review` で差分レビューを行う
5. 必要なら follow-up を `governance/tasks/` に戻す

## 注意点

- `validation-report` はレビューコメントの代替ではない
- `/review` は検証記録の代替ではない
- 反復して同じレビュー面が必要になったときだけ `governance/reviews/` 新設を再検討する
