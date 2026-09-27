---
name: backend-implementation-integration
description: Go バックエンドの API 実装の最後に、main.go での DI 配線・ルーティング登録と、Handler → Usecase → Infra → DB を実物で通す E2E テストを実装する手順。Handler 層の実装が終わった後、または main.go の配線や e2e/ のテストを追加・変更するときに、ユーザーが手順を指定しなくても必ず使う。バックエンド実装の5スキル(domain → usecase → infra → handler → integration)の5番目。配線 → E2E テストケース洗い出し → E2E テストコード実装 → 全体検証 を進め、各段階でサブエージェントによるルール適合判定とユーザー確認を挟む。
---

# バックエンド実装: 繋ぎこみ(DI 配線・E2E)

## 最初に読むもの

作業を始める前に、必ず `.claude/skills/backend-implementation-shared/workflow.md` を読み、そこに書かれた共通手順(実装対象の確認、常に守ること、判定ループ、ユーザー確認の出し方)に従う。このファイルには繋ぎこみ特有の手順だけを書いている。

## 前提

- 実装対象の API の Domain / Usecase / Infra / Handler がすべて揃い、`go test ./domain/... ./application/... ./infra/... ./handler/...` が通ること。
- 満たしていない場合は、この段階で補わずにユーザーへ伝え、該当する層のスキルに戻るかどうかを確認する。

## 対象

- `main.go`: 具象実装の組み立て(コンストラクタ経由の注入)とルーティング登録
- `e2e/`: 実物の Handler → Usecase → Infra → DB を通す E2E テスト

各層のプロダクトコード(`domain/` `application/` `infra/` `handler/`)は、この段階では変更しない。配線や E2E で不具合が見つかった場合は、原因の層をユーザーに伝え、その層のスキルに戻るかどうかを確認する。

この段階のフェーズは、他の層の「スケルトン → テストケース → テストコード → 本実装」ではなく、次の4フェーズで進める。

## フェーズ1: DI 配線とルーティング登録

`main.go` に次を追加する。

- Infra の具象実装 → DomainService → Usecase → Handler の順に、コンストラクタ経由で組み立てる。すでに生成済みのインスタンス(リポジトリ、ロガーなど)があれば使い回す。
- `backend-implementation-handler` で確定したメソッド + パスで、ルーティングを登録する(`e.GET` / `e.POST` など)。
- 環境変数など新しい設定値が必要な場合は、`docker-compose.yml` などへの追加が必要かをユーザーに確認する(勝手に追加しない)。

`go build ./...` が通ることを確認する。

判定は **architecture-reviewer** に依頼する(具象実装の組み立てが `main.go` のみで行われているか、パスが REST の原則に沿っているか)。

## フェーズ2: E2E テストケースの洗い出し

**コードは書かない。** 実物の全層を通したときに、API の主要な結果(ステータスコード、レスポンスボディ、DB に保存・取得される値)が正しいことを確認するケースを洗い出す。

- E2E は全層 + DB を通すため最も壊れやすく、実行コストも高い。**基本は正常系1ケースに絞る。** 各層の単体テストで検証済みの分岐(バリデーション、エラー種別ごとのステータスコードなど)は E2E で重複させない。
- 層をまたいだ組み合わせでしか確認できない振る舞いがある場合に限り、ケースを追加する。追加する場合は理由を添えてユーザー確認で示す。

`.claude/rules/backend-testing.md` の「テストケースの定義」に従い、「XXであること」の形式で書く。判定は **testcase-reviewer** に依頼する(E2E 専用のルールはテストルールにないため、上記の方針と既存の `e2e/` のテストを参照先として渡す)。

## フェーズ3: E2E テストコードの実装

既存の `e2e/thread_list_test.go` / `e2e/thread_create_test.go` に倣う。

- ファイルは `e2e/xxx_test.go`、関数名は `TestE2E_Xxx_正常系` の形式にする。
- DB は `e2e/testmain_test.go` の `TestMain`(testcontainers で Postgres を起動し `E2E_DATABASE_URL` に設定)を使う。
- モックを使わず、`main.go` と同じ順序で実物の Infra → Usecase → Handler を組み立てる。
- `httptest.NewRequest` / `httptest.NewRecorder` と `e.NewContext(req, rec)` で `echo.Context` を作り、ハンドラのメソッドを直接呼ぶ。`e.Validator` など `main.go` で設定している Echo の設定のうち、ハンドラの動作に必要なものは同じく設定する。
- 事前データの投入と後始末は `pool.Exec` と `t.Cleanup` で行う(ハンドラが内部で DB に書き込むため、インフラのテストのようなトランザクションのロールバックは使えない)。後始末は、他のテストのデータを消さないよう、そのテストで投入・作成したデータに限定する。
- 検証は、ステータスコード、レスポンスボディ、必要に応じて DB に保存された値の3点で行う。`.claude/rules/backend-testing.md` の「共通ルール」(if を書かない、`require` / `assert` で検証する)に従う。

判定は **test-code-reviewer** に依頼する(参照先として既存の `e2e/` のテストも渡す)。

## フェーズ4: 全体検証

1. `go build ./...` と `go test ./...`(E2E を含む全テスト)を実行し、すべて通ることを確認する。失敗した場合は、原因の層を特定してユーザーに伝える。E2E のテストコード自体の誤りでない限り、この段階でプロダクトコードを直さない。
2. **architecture-reviewer** に、今回実装した API に関わる全ファイル(Domain / Usecase / Infra / Handler / `main.go`)を通しで判定依頼する。各層単体では見えなかった違反(インフラへのロジック混入、ドメイン層での `time.Now()` 呼び出し、handler からの `domain/` の直接 import など)が、繋ぎこみ後に初めて見えるため。
3. 違反が見つかった場合は、どの層のどのファイルかを示してユーザーに相談し、該当する層のスキルに戻って直すかどうかを確認する。

最後のユーザー確認をもって、実装対象の API は完了とする。
