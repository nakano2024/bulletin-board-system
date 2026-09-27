---
name: backend-implementation-usecase
description: Go バックエンドの API 実装のうち Usecase 層(XxUsecase、DTO の XxCommand / XxOutput、IQueryService・ITimeGetter・ILogger など application 層の汎用インターフェース)を実装する手順。Domain 層の実装が終わった後、またはユースケースの追加や変更を行うときに、ユーザーが手順を指定しなくても必ず使う。バックエンド実装の5スキル(domain → usecase → infra → handler → integration)の2番目。スケルトン実装 → テストケース洗い出し → テストコード実装 → 本実装 を進め、各段階でサブエージェントによるルール適合判定とユーザー確認を挟む。
---

# バックエンド実装: Usecase 層

## 最初に読むもの

作業を始める前に、必ず `.claude/skills/backend-implementation-shared/workflow.md` を読み、そこに書かれた共通手順(実装対象の確認、常に守ること、判定ループ、ユーザー確認の出し方)に従う。このファイルには Usecase 層特有の手順だけを書いている。

## 前提

- 実装対象の API に必要なドメインモデル・`IRepository` インターフェース・センチネルエラーが `domain/` に存在し、`go test ./domain/...` が通ること。
- 満たしていない場合は、この層で補わずにユーザーへ伝え、`backend-implementation-domain` に戻るかどうかを確認する。

## 対象

`XxUsecase`、DTO(`XxCommand` / `XxOutput`)、`IQueryService`・`ITimeGetter`・`IUuidGenerator`・`ILogger` など application 層の汎用インターフェース(`application/`)。

`IRepository` の実装、Handler、`main.go` はこの層では作らない。Domain 層で定義した `IRepository` インターフェースをそのまま使う。`domain/` の変更が必要になった場合は、この層で直さずユーザーに相談する。

## フェーズ1: スケルトン実装

作るもの:

- ユースケース型 `XxUsecase`、コンストラクタ、`Exec()` のシグネチャ(公開メソッドは `Exec()` のみ)
- DTO(`XxCommand` / `XxOutput`)
- `IQueryService`、`ITimeGetter`、`IUuidGenerator`、`ILogger` など、今回の API に必要な application 層の汎用インターフェース(既存のものがあれば流用する)

DTO は次を守る。

- `XxCommand` / `XxOutput` にドメインモデル型を持たせない。
- `XxCommand` に Web 層の型(`http.Header`、`echo.Context` など)を持たせず、ビジネスロジックが実際に使う最小限のプリミティブ値に絞る。
- `XxOutput` は `IRepository` / `IQueryService` の戻り値の型を使い回さず、専用の型を定義する。

判定は **architecture-reviewer** に依頼する。

ユーザー確認では、DTO にドメインモデル型が混ざっていないか、`Exec()` が「取得 → ドメインモデルに委譲 → 保存」の組み立てに徹する形になっているか、`Exec()` の切り方が原子性を保つべき処理のまとまりになっているかを重点的に見てもらう。

## フェーズ2: テストケースの洗い出し

`Exec()` の入力に対してどの出力・エラーを返すかを基準にケースを洗い出す。ドメインモデル自体の判定ロジック(不変条件の境界値など)は Domain 層のテストで検証済みなので、同じ観点を重複して洗い出さない。ただし、仕様で利用者起因のエラーとされている入力(本文が空など)について、`Exec()` がそのエラーをそのまま返すことは Usecase の出力として洗い出す(handler がそのエラーでステータスコードを判定するため)。依存(`IRepository` など)がエラーを返したときに、エラーが返ることのケースは含める。

判定は **testcase-reviewer** に依頼する。

## フェーズ3: テストコードの実装

`.claude/rules/backend-testing.md` の「共通ルール」「モックの使い方」「ユースケースのテスト」に従う。

- `IRepository` / `IQueryService` / 汎用インフラ(`ITimeGetter`、`ILogger` など)は gomock でモックする。具象実装がまだ存在しなくても、インターフェースが確定していればモックできる。
- ドメインモデル・DomainService はモックせず実物を使う。DomainService に注入する依存はモックしてよい。
- 正常系と異常系はテスト関数を分ける。

この時点では Infra の実装がないため、`main.go` の DI 配線などは不要。判定は **test-code-reviewer** に依頼する。

## フェーズ4: 本実装

テストが通るように中身を実装する(`go test ./application/...`)。

- `Exec()` は「取得 → ドメインモデルに委譲 → 保存」の組み立てに徹し、業務判断の分岐が増えたら Domain 層に移せないか見直す(移す場合はユーザーに相談する)。
- 依存がエラーを返したときは、`ILogger` 経由でログを出してからエラーを返してよい。
- ドメインモデル → `XxOutput` の変換はこの層で行う。

層の完了時、**architecture-reviewer** に再度判定を依頼する。OK かつユーザー確認が取れたら Usecase 層は完了。次は `backend-implementation-infra` スキルを使うことをユーザーに案内する。
