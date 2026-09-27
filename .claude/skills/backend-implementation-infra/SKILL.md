---
name: backend-implementation-infra
description: Go バックエンドの API 実装のうち Infra 層(IRepository / IQueryService の Postgres 実装、ITimeGetter・ILogger などの汎用インフラ実装)を実装する手順。Usecase 層の実装が終わった後、またはリポジトリ実装・DB アクセスの追加や変更を行うときに、ユーザーが手順を指定しなくても必ず使う。バックエンド実装の5スキル(domain → usecase → infra → handler → integration)の3番目。スケルトン実装 → テストケース洗い出し → テストコード実装 → 本実装 を進め、各段階でサブエージェントによるルール適合判定とユーザー確認を挟む。
---

# バックエンド実装: Infra 層

## 最初に読むもの

作業を始める前に、必ず `.claude/skills/backend-implementation-shared/workflow.md` を読み、そこに書かれた共通手順(実装対象の確認、常に守ること、判定ループ、ユーザー確認の出し方)に従う。このファイルには Infra 層特有の手順だけを書いている。

## 前提

- 実装対象の API に必要なインターフェース(`domain/` の `IRepository`、`application/` の `IQueryService`・汎用インターフェース)が確定しており、`go test ./domain/... ./application/...` が通ること。
- 満たしていない場合は、この層で補わずにユーザーへ伝え、`backend-implementation-domain` / `backend-implementation-usecase` に戻るかどうかを確認する。

## 対象

`IRepository` / `IQueryService` / 汎用インフラの実装(`infra/`)と、それに必要なマイグレーション(`infra/postgres/migrations/`)。

ハンドラ、`main.go` の DI 配線はこの層では作らない(それぞれ `backend-implementation-handler` / `backend-implementation-integration` で扱う)。インターフェースのシグネチャを変えたくなった場合は、この層で直さずユーザーに相談する。

## フェーズ1: スケルトン実装

作るもの:

- `infra/` の実装型とコンストラクタ、インターフェースを満たすメソッドのシグネチャ。型名は `PostgresThreadRepository` のように技術要素を含む名前にする
- テーブル・カラムの追加が必要な場合は、マイグレーション SQL(既存の `infra/postgres/migrations/` の命名に倣う)
- すでに同じインターフェースの実装がある場合は、新しい型を作らずメソッドを追加する

判定は **architecture-reviewer** に依頼する。

ユーザー確認では、テーブル設計(マイグレーション)と、実装がインターフェースのメソッドと1対1に対応しているかを重点的に見てもらう。

## フェーズ2: テストケースの洗い出し

`.claude/rules/backend-testing.md` の「インフラのテスト」に従い、DB を伴う `infra/` のテストケースを洗い出す。

- DB とのやりとり(フィルタ条件、並び順、件数、カラムの読み書き)が正しいことの確認に絞る。
- ドメインルール(必須項目、文字種など)をインフラ層で再検証するケースは作らない。
- DB 以外の実装(時刻取得、ロガーなど)は、ロジックを持たない限りテストケースを作らない。テストしたくなるロジックがある場合は、Domain 層に寄せられないかユーザーに相談する。

判定は **testcase-reviewer** に依頼する。

## フェーズ3: テストコードの実装

`.claude/rules/backend-testing.md` の「共通ルール」「インフラのテスト」に従う。

- DB のテストは、テストケースごとに「データを投入 → 検証 → データを壊す」のフローをトランザクションで実装し、テスト終了時にロールバックする。
- DB は既存の `infra/postgres/testmain_test.go`(testcontainers でマイグレーション適用済みの Postgres を起動)を使う。新しいパッケージで DB テストが必要な場合は、同じ形の `TestMain` を用意する。

判定は **test-code-reviewer** に依頼する。

## フェーズ4: 本実装

テストが通るように中身を実装する(`go test ./infra/...`)。

- 実装は外部リソースとのやりとりと型変換に徹する。業務判断の条件分岐や計算を持たせない。
- DB から取得した値をドメインモデルに戻すときは、ドメインモデルのコンストラクタを経由し、フィールドを直接組み立てない。

層の完了時、**architecture-reviewer** に再度判定を依頼し、インフラへのロジック混入がないかを重点的に見てもらう。OK かつユーザー確認が取れたら Infra 層は完了。次は `backend-implementation-handler` スキルを使うことをユーザーに案内する。
