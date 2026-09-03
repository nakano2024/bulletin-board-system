---
paths:
  - "handler/**/*_test.go"
  - "application/**/*_test.go"
  - "domain/**/*_test.go"
  - "infra/**/*_test.go"
---

# バックエンドテスト実装ルール（Go）

設計ルールは `.claude/rules/backend-architecture.md` に従う。

## テストケースの定義

CLAUDE.md に書かれた仕様をもとにケースを洗い出す。

- **「ユーザーの入力に対して、どの値を返すか」** を基準に定義する。
- ケース名は **「XXであること」** の形式で書く。
- 境界値チェックと代表値チェックが必要な箇所では、必ず両方を行う。

```text
good: タイトルが1文字のとき、スレッドが作成されること
good: タイトルが101文字のとき、バリデーションエラーが返ること
bad:  タイトルのテスト
bad:  正常に動作すること
```

「どう動くか」ではなく「何が返るか」で書くと、実装を変えてもテストが壊れにくくなる。

## 共通ルール

### if 文を書かない

テストコードの中に `if` を含めない。分岐が必要になった時点で、テスト関数の粒度が粗い。

- **1つのテスト関数は、`if` なしで書ける粒度にする。**
- 正常系と異常系は別のテスト関数に分ける（`wantErr` で分岐させない）。
- エラー検証は `require.NoError` / `require.ErrorIs` などのヘルパーで行い、自前で分岐しない。

### Table Driven Test

テストケースは Table Driven Test で管理する。テーブルは次の3プロパティを持つ。

| プロパティ | 内容 |
| --- | --- |
| 入力 | テスト対象に渡す値 |
| モックの挙動定義 | スタブ / モックの設定関数 |
| 期待値 | 返却されるべき値 |

```go
func TestCreateThreadUsecase_Exec_正常系(t *testing.T) {
	tests := []struct {
		name     string
		input    CreateThreadCommand
		setupMock func(*mock_thread.MockIThreadRepository)
		want     *CreateThreadOutput
	}{
		{
			name:  "タイトルが1文字のとき、スレッドが作成されること",
			input: CreateThreadCommand{Title: "a", AuthorID: "u1"},
			setupMock: func(m *mock_thread.MockIThreadRepository) {
				m.EXPECT().Save(gomock.Any(), gomock.Any()).Return(nil)
			},
			want: &CreateThreadOutput{ID: "fixed-uuid", Title: "a"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			repo := mock_thread.NewMockIThreadRepository(ctrl)
			tt.setupMock(repo)

			sut := NewCreateThreadUsecase(repo, stubUuidGen, stubTimeGetter)
			got, err := sut.Exec(context.Background(), tt.input)

			require.NoError(t, err)
			assert.Equal(t, tt.want, got)
		})
	}
}
```

## モックの使い方

| 確かめたいこと | 使う手法 |
| --- | --- |
| 依存が返す値によって、テスト対象の挙動がどう変わるか | **スタブ**。値を静的に返すだけにする |
| テスト対象が依存に書き込む値が正しいか | **モック**。渡された引数を検証する |
| — | **スパイは基本的に使わない** |

1つのテストで両方を混ぜない。値の変化を見たいなら引数検証は書かず、引数を見たいなら戻り値は固定にする。検証対象が2つあるテストは、失敗したときに原因が特定できない。

## テスト対象ごとのルール

### ユースケースのテスト

- gomock を用いて、インフラ依存のあるモジュール（`IRepository`、`IQueryService`、汎用インフラ）をモックする。
- モックかスタブかは「モックの使い方」の基準で決める。
- **ドメインサービスおよびドメインモデルの振る舞いはモックしない。** 実物を使う。
- ドメインサービスがインフラ依存を持つ場合、そのドメインサービスに注入する依存をモックするのは可。

### ドメインサービスのテスト

- gomock を用いて、インフラ依存のあるモジュールをモックする。
- モックかスタブかは「モックの使い方」の基準で決める。

### ドメインモデルのテスト

- 基本的にそのままテストしてよい。モックは不要。

### インフラのテスト

**DB の場合**

テストケースごとに「データを投入 → 検証 → データを壊す」のフローで行う。軽量なのでトランザクションで実装し、テスト終了時にロールバックする。

**DB 以外の場合**

- 無理にテストしなくてよい。
- 外部依存や API 呼び出しをモックする手段があれば、それを使ってテストする。
- 乱数関連は、標準的な手段があれば使う（時間取得なら時刻を固定してテストするなど）。
- そもそも設計ルールどおり、インフラにロジックを持たせないことでテストの必要性を減らす。テストしたくなるロジックがインフラにある場合は、ドメインに寄せられないか見直す。

### ハンドラのテスト

- echo の `e.NewContext(req, rec)` で `echo.Context` を組み立て、ハンドラのメソッドを直接呼び出す（`httptest.NewRequest` / `httptest.NewRecorder` を使う）。ルーティング設定（`main.go` の `e.GET` など）自体はテスト対象に含めない。
- Usecase は gomock でモックする。モックかスタブかは「モックの使い方」の基準で決める。
- 検証観点は「HTTPステータスコード」「レスポンスボディ（JSON）」の2点に絞る。
