# Codex 利用ガイド

## 対象読者

- contributor
- repository maintainer

## 使う場面

- この repository で Codex を使って作業するとき
- どの文書を先に読むべきか迷うとき
- skills や contracts をどう使うか確認したいとき

## 基本方針

- 先に `README.md` と root `AGENTS.md` を読む
- 構造やルールに触れる可能性があれば `blueprint/` と `governance/policies/` を先に読む
- 実装だけの変更に見えても change classification を先に行う

## 人間向け面の使い分け

- 依頼や引き継ぎは `governance/contracts/`
- 一時的な共有文脈は `governance/tasks/`
- 進め方の案内は `governance/guides/`
- 検証導線は `scenarios/`

## repository skills

- `classify-change`
- `blueprint-to-impl`
- `scenario-validation`

必要なときだけ使い、正本ドキュメントの代替にはしない。

## 注意点

- `.codex/` は AI ローカル補助であり、正本ではない
- 技術的な真実は英語の正本へ戻る
- 日本語面では説明を再定義せず、参照先を案内する
