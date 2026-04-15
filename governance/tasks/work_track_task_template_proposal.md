# Work Track Task Template Proposal

## Background

この文書は、`governance/tasks/marketdata_review_followups.md` で使った見出し構造を、今後の task document 向けテンプレート案として保持する。

これは正式な repository-wide policy ではなく、複数の作業トラックを 1 つの temporary task context で扱う場合の参考構造である。
正式な運用ルールへ昇格する場合は `governance/policies/` または `governance/tasks/README.md` への反映を別途判断する。

## In Scope

- 複数 work track を含む task document の見出し構造案
- 決定済み事項と未決事項を分けるための見出し粒度
- promotion target と acceptance criteria を残すための配置案

## Out Of Scope

- すべての task document への適用義務化
- `governance/tasks/README.md` の required fields 変更
- durable source of truth の代替

## Template

```md
# <Topic> フォローアップ

## Background

<一時的な共有作業文脈として、この文書を作る背景を書く。>

<恒久仕様ではないこと、durable な結論は適切な正本へ昇格することを書く。>

## In Scope

- <対象範囲>
- <対象範囲>

## Out Of Scope

- <対象外>
- <対象外>

## Work Tracks

### Track 1: <Track Name>

#### Goal

<この track の目的を書く。>

#### Decisions

- <決定済み事項>
- <決定済み事項>

#### Expected Impact

- <影響範囲>
- <影響範囲>

#### Notes

- <補足>
- <補足>

### Track 2: <Track Name>

#### Goal

<この track の目的を書く。>

#### Open Questions

- <未決論点>
- <未決論点>

#### Expected Impact

- <影響範囲>
- <影響範囲>

#### Notes

- <補足>
- <補足>

## Next Actions

- <次にやること>
- <次にやること>

## References

- `<参照先>`
- `<参照先>`

## Promotion Targets

- <durable な結論の昇格先>
- <durable な結論の昇格先>

## Acceptance Criteria

- <この task document が満たすべき条件>
- <この task document が満たすべき条件>

## Result

- <作成・更新した結果>
- <作成・更新した結果>
```

## Usage Notes

- `Decisions` は決定済みの track に使う
- `Open Questions` は検討中の track に使う
- 1 つの track 内に決定済み事項と未決事項が混在する場合は、両方の見出しを置いてよい
- `Expected Impact` は implementation、data、scenario、blueprint などの影響範囲を明示する
- `Promotion Targets` は temporary context から durable source of truth へ昇格する候補を書く

## Acceptance Criteria

- `marketdata_review_followups.md` の見出し構造がテンプレート案として再利用できる
- required fields を含んだまま、複数 work track の整理に使える
- 正式ルールではなく proposal であることが明確である

## Result

- `governance/tasks/marketdata_review_followups.md` の見出し構造を `governance/tasks/work_track_task_template_proposal.md` にテンプレート案として保持した
