---
name: backend-implementation-handler
description: Go バックエンドの API 実装のうち Handler 層(Echo のハンドラ、リクエスト / レスポンス型、エラーのステータスコード・メッセージ変換)を実装する手順。Infra 層の実装が終わった後、またはハンドラ・レスポンス形式・エラーハンドリングの追加や変更を行うときに、ユーザーが手順を指定しなくても必ず使う。バックエンド実装の5スキル(domain → usecase → infra → handler → integration)の4番目。スケルトン実装 → テストケース洗い出し → テストコード実装 → 本実装 を進め、各段階でサブエージェントによるルール適合判定とユーザー確認を挟む。
---

# バックエンド実装: Handler 層

## 最初に読むもの

作業を始める前に、必ず `.claude/skills/backend-implementation-shared/workflow.md` を読み、そこに書かれた共通手順(実装対象の確認、常に守ること、判定ループ、ユーザー確認の出し方)に従う。このファイルには Handler 層特有の手順だけを書いている。

## 前提

- 実装対象の API の `XxUsecase` と DTO(`XxCommand` / `XxOutput`)、ハンドラが判定に使うセンチネルエラーが確定していること。
- `backend-implementation-infra` が完了していること(`go test ./domain/... ./application/... ./infra/...` が通ること)。
- 満たしていない場合は、この層で補わずにユーザーへ伝え、前の層のスキルに戻るかどうかを確認する。

## 対象

ハンドラ、リクエスト / レスポンス型、エラー変換(`handler/`)。

`main.go` でのルーティング登録・DI 配線はこの層では行わない(`backend-implementation-integration` で扱う)。Usecase の DTO やエラーを変えたくなった場合は、この層で直さずユーザーに相談する。

## フェーズ1: スケルトン実装

作るもの:

- ハンドラの型、コンストラクタ、メソッドシグネチャ(`func (h *XxHandler) Xx(c echo.Context) error`)
- ハンドラが依存する Usecase のインターフェース。既存ハンドラに倣い、handler パッケージ内に非公開のインターフェース(`Exec()` のみ)として定義し、テストでモックできるようにする
- リクエスト型(バリデーションタグを含む)とレスポンス型 `XxResponse`

`.claude/rules/backend-architecture.md` の「プレゼンテーション層(ハンドラ)」に従う。

- パスは REST の原則(複数形のリソース名、動詞を含めない、ネストは親パスにぶら下げる)に沿う。パスはこのフェーズで確定させ、ルーティング登録は `backend-implementation-integration` で行う。
- レスポンスの JSON はルートにリソース名のキー(一覧は複数形、単体は単数形)を置く。
- `XxResponse` に持たせるフィールドは、CLAUDE.md の仕様で要求されている項目に限定する。
- `XxCommand` には、リクエストから取り出した値をプリミティブ型に変換して詰める。

判定は **architecture-reviewer** に依頼する。

ユーザー確認では、パス、リクエスト / レスポンスの JSON の形、エラー種別ごとのステータスコードとユーザー向けメッセージの対応表を重点的に見てもらう。

## フェーズ2: テストケースの洗い出し

`.claude/rules/backend-testing.md` の「ハンドラのテスト」に従い、「HTTP ステータスコード」「レスポンスボディ(JSON)」を基準にケースを洗い出す。

- 正常時のステータスコードとレスポンスボディ
- リクエストの形式不正・バリデーションエラーのときの 400
- Usecase が返すエラーの種類ごとのステータスコード(400 / 404 / 500)と固定メッセージ
- ハンドラが明示的にログを出す場合は、そのログ内容

業務ルール自体の検証(Domain / Usecase 層で検証済みのもの)は重複して洗い出さない。

判定は **testcase-reviewer** に依頼する。

## フェーズ3: テストコードの実装

`.claude/rules/backend-testing.md` の「共通ルール」「モックの使い方」「ハンドラのテスト」に従う。

- `httptest.NewRequest` / `httptest.NewRecorder` と `e.NewContext(req, rec)` で `echo.Context` を組み立て、ハンドラのメソッドを直接呼ぶ。ルーティング設定はテスト対象に含めない。
- Usecase は gomock でモックする(既存の `handler/thread/mock_thread/` に倣う)。
- ログを検証する場合は、`e.Logger.SetOutput(&buf)` で出力先を差し替えて `assert.Contains` で確認する。

判定は **test-code-reviewer** に依頼する。

## フェーズ4: 本実装

テストが通るように中身を実装する(`go test ./handler/...`)。

- Usecase のエラーは `errors.Is` / `errors.As` で判定し、エラーメッセージの文字列比較で分岐しない。
- `err.Error()` やスタックトレースをレスポンスに含めず、エラー種別ごとの固定メッセージ(例: 「スレッドが存在しません。」)に変換する。詳細はログにのみ残す。
- `XxOutput` → `XxResponse` の詰め替えはこの層で行う。

層の完了時、**architecture-reviewer** に再度判定を依頼する。OK かつユーザー確認が取れたら Handler 層は完了。次は `backend-implementation-integration` スキルを使うことをユーザーに案内する。
