---
name: backend-implementation-domain
description: Go バックエンドの API 実装のうち Domain 層(Entity / Aggregate、DomainService、IRepository インターフェース、センチネルエラー)を実装する手順。新しい API の実装を始めるとき、またはドメインモデル・IRepository の追加や変更を行うときに、ユーザーが手順を指定しなくても必ず使う。バックエンド実装の5スキル(domain → usecase → infra → handler → integration)の1番目。スケルトン実装 → テストケース洗い出し → テストコード実装 → 本実装 を進め、各段階でサブエージェントによるルール適合判定とユーザー確認を挟む。
---

# バックエンド実装: Domain 層

## 最初に読むもの

作業を始める前に、必ず `.claude/skills/backend-implementation-shared/workflow.md` を読み、そこに書かれた共通手順(実装対象の確認、常に守ること、判定ループ、ユーザー確認の出し方)に従う。このファイルには Domain 層特有の手順だけを書いている。

## 前提

- Domain 層は5層の最初なので、前の層の完了確認は不要。実装対象の API は、共通手順の「実装対象の確認」に従ってユーザーに尋ねる。
- 既存の `domain/` に流用できるモデルや `IRepository` がないかを先に確認し、新規作成と既存の拡張のどちらにするかをフェーズ1のユーザー確認で示す。

## 対象

Entity / Aggregate、DomainService、`IRepository` インターフェース、センチネルエラー(`domain/`)。Usecase 型・DTO・Infra 実装・Handler・`main.go` など他層のコードはこの層では作らない・変更しない。

`domain/` は標準ライブラリのみを import する。`time.Now()`、`uuid.New()`、乱数、DB・ファイル I/O を直接書かない(必要な値は引数で受け取る)。

## フェーズ1: スケルトン実装

作るもの:

- ドメインモデル(Entity / Aggregate)の型とコンストラクタのシグネチャ
- DomainService の型とメソッドシグネチャ(必要な場合のみ。まず Entity や値オブジェクトのメソッドで表現できないかを検討する)
- `IRepository` インターフェース(集約ルート単位。メソッドは業務単位で閉じ、汎用的な検索条件を引数に取らない)
- ハンドラで 404 / 400 に振り分ける必要があるエラーのセンチネルエラー(例: `ErrThreadNotFound`)

判定は **architecture-reviewer** に依頼する。

ユーザー確認では、集約の切り方と `IRepository` のメソッドの切り方(業務単位で閉じているか)を重点的に見てもらう。

## フェーズ2: テストケースの洗い出し

CLAUDE.md の仕様のうち、ドメインモデル・DomainService の振る舞いに関わるケースを洗い出す。Usecase の入出力の組み立てなど Usecase 層の責務にあたるケースは、ここでは含めない(`backend-implementation-usecase` で洗い出す)。

判定は **testcase-reviewer** に依頼する。

## フェーズ3: テストコードの実装

`.claude/rules/backend-testing.md` の「共通ルール」「モックの使い方」「ドメインモデルのテスト」「ドメインサービスのテスト」に従う。

- ドメインモデルはモックなしでそのままテストする。
- DomainService がインフラ依存(`IRepository` など)を持つ場合は、その依存を gomock でモックする。

判定は **test-code-reviewer** に依頼する。

## フェーズ4: 本実装

テストが通るように中身を実装する(`go test ./domain/...`)。

層の完了時、**architecture-reviewer** に再度判定を依頼する。OK かつユーザー確認が取れたら Domain 層は完了。次は `backend-implementation-usecase` スキルを使うことをユーザーに案内する。

Domain 層の変更によって既存の Usecase / Infra / Handler / `main.go` がコンパイルできなくなった場合は、直さずに影響範囲(ファイルと、どの層のスキルで扱うか)をユーザーに伝える。
