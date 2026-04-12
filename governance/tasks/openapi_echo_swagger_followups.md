# OpenAPI / Echo Swagger Followups

## Background

このリポジトリでは、契約の正本を実装ソースに置き、生成された OpenAPI 成果物を `implementations/dag-core/openapi/` に配置する運用を採用している。

今回のフォローアップは、正本モデルを維持したまま OpenAPI の仕様・運用を再整理し、`echo-swagger` によるドキュメント提供方針を確定することを目的とする。

このタスク文書は実装を含まない検討用ドキュメントであり、意思決定と運用準備に限定する。

## In Scope

- OpenAPI の正本、生成、検証責務の再明確化
- `echo-swagger` で Swagger UI を提供する際の運用方針の定義
- 実装前に必要なタスクリストの確定
- ドキュメント、CI、運用整合の受け入れ基準定義

## Out Of Scope

- ランタイム API エンドポイントの追加・変更
- `echo-swagger` ミドルウェアやルートの実装
- 実装正本モデルからの逸脱や transport 構造変更
- このフェーズでの OpenAPI 生成ツールチェーン置換

## Task List

1. OpenAPI 運用ルールを 1 つの実装向け導線に統合する。
2. 役割分担を明文化する（`swag` は生成、`echo-swagger` は表示）。
3. Swagger UI 公開方針を環境別に決める（`dev`、`staging`、`prod`）。
4. 既定ルートと互換方針を決める（例: `/swagger/*`）。
5. handler 注釈と `openapi/` 生成物の同期ルールを定義する。
6. `make contracts-check` における OpenAPI ドリフト時の CI 判定を明確化する。
7. API 変更後のローカル検証フローを定義する（再生成、テスト、ドキュメントルート確認）。
8. PR レビュー観点を定義する（ソース差分、生成物差分、互換性リスク）。
9. 後続課題（例: OpenAPI v2 から v3 移行）を本件と分離して記録する。
10. このタスクで確定した恒久ルールを適切な恒久文書へ昇格する。

## Implementation Checklist

- [x] `echo-swagger` 依存を `implementations/dag-core/go.mod` に追加する。
- [x] `Config` に Swagger UI 制御設定（有効/無効、ルート）を追加する。
- [x] `ENV=dev` 既定有効、`staging/prod` 既定無効の判定ロジックを実装する。
- [x] `internal/di/echo_container.go` で `SwaggerUIEnabled` 時のみ `/swagger/*` を登録する。
- [x] OpenAPI 生成パッケージを Swagger UI 提供経路から参照できるようにする。
- [x] 設定解決ロジックのユニットテストを追加する。
- [x] `implementations/dag-core` の `go test ./...` で回帰確認する。
- [x] `make contracts-check` を実行し、`swag` 探索範囲不足と `buf` 未導入の問題を解消する。
- [x] `make contracts-check` の最終 diff チェックで出る OpenAPI 差分（`implementations/dag-core/openapi/*`）を、実装変更に伴う契約更新としてレビュー対象に確定する。
- [ ] 実運用のアクセス制御（認証/ネットワーク）実装を追加する必要がある場合は別タスクで扱う。

## Acceptance Criteria

- OpenAPI の正本、生成、表示責務をリポジトリ文書から追える。
- 実装開始前に `echo-swagger` の環境別公開方針が確定している。
- API 契約レビュー観点が既存 `contracts-check` 運用と整合している。
- 将来課題が defer 項目として明示され、本件スコープと混在しない。

## Result

- OpenAPI / Swagger UI の採用方針を恒久文書へ昇格した。
- 昇格先:
  - `README.md`（repository-level 要約方針）
  - `implementations/dag-core/README.md`（実装運用の詳細方針）
- `echo-swagger` 実装、設定トグル、単体テスト、`go test` 実行、`contracts-check` 実行まで完了した。
- `contracts-check` の最終失敗は生成差分検出による想定動作であり、差分を取り込み対象に確定した。
- 本タスク文書は実装前検討メモとしては完了扱いとし、将来の v2->v3 などの別課題は独立タスクで追跡する。
