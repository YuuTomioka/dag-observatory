# 文書振り分けガイド

## 対象読者

- contributor
- operator
- repository maintainer

## 使う場面

- 新しい文書をどこに置くべきか迷ったとき
- 既存文書を昇格または移動すべきか判断したいとき

## 判断順序

1. その内容は構造や仕様の正本か
2. その内容は repository-wide のルールか
3. その内容は人間向けの受け渡し面か
4. その内容は人間向けの進め方や手順か
5. その内容は代表的な検証導線か
6. その内容は一時的な共有文脈か
7. その内容は実装、データ、platform、deployments に閉じるか

## 配置先

- 構造、境界、長期の設計原則: `blueprint/`
- durable decision と rationale: `blueprint/adr/`
- repository-wide のルール: `governance/policies/`
- 人間向けの依頼、ハンドオフ、報告テンプレート: `governance/contracts/`
- 人間向けの進め方、操作手順、文書の使い分け: `governance/guides/`
- 一時的な共有作業文脈: `governance/tasks/`
- 代表的な検証導線: `scenarios/`
- 実装固有の説明: `implementations/`
- 観測基盤の運用参照: `platform/`
- SQL とデータ設計資産: `data/`
- ローカル実行ラッパ: `deployments/`

## 判断基準

- 技術的な真実を定義するなら日本語面ではなく英語の正本へ置く
- 人間が依頼や確認のために直接使うなら日本語面へ置く
- 一時的にしか使わないなら `governance/tasks/`
- 何度も使うようになったら `governance/guides/` か `governance/contracts/` へ昇格する

## 参照先

- `governance/policies/change-policy.md`
- `governance/policies/documentation-policy.md`
- root `AGENTS.md`
