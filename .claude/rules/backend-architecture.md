---
paths:
  - "handler/**/*.go"
  - "application/**/*.go"
  - "domain/**/*.go"
  - "infra/**/*.go"
---

# バックエンド実装ルール（オニオンアーキテクチャ + DDD / Go）

## ディレクトリ構成

プロジェクトルート直下に、層と1対1で対応するディレクトリを置く。

```text
myapp/
├── handler/       # プレゼンテーション層
├── application/   # アプリケーション層
├── domain/        # ドメイン層
├── infra/         # インフラ層
└── main.go        # 依存の組み立て（DI）
```

層とディレクトリは常に1対1で対応させる。複数の層にまたがるディレクトリを新設しない。

## 依存の方向

| ディレクトリ | import してよい |
| --- | --- |
| `handler/` | `application/` |
| `application/` | `domain/` |
| `domain/` | 標準ライブラリのみ |
| `infra/` | `domain/`, `application/`（インターフェイスの実装のため） |

- `domain/` は他のどの層も import しない。
- `application/` は `infra/` を import しない。必ず対応するインターフェイスを介する（依存性逆転）。
- `handler/` は `domain/` を直接 import しない。やりとりは後述の DTO に限る。
- 具象実装の組み立ては `main.go` でのみ行い、コンストラクタ経由で注入する。

## ユースケース

- ビジネスロジックのオーケストレータは `XxUsecase` とし、`application/` に置く。
- 実行単位は `Exec()` メソッド1つ。1ユースケースにつき公開メソッドは `Exec()` のみとする。
- `Exec()` は `XxCommand` を受け取り `XxOutput` を返す。
- **1つの `Exec()` は、原子性を保つべき処理のまとまり単位で切る。** 途中まで実行して一部だけ確定してよい処理を1つの `Exec()` に混在させない。逆に、失敗時に途中経過を残してはいけない一連の処理は、複数のユースケースに分割せず1つの `Exec()` にまとめる。

```go
// application/thread/create_thread_usecase.go
type CreateThreadUsecase struct {
	threadRepo thread.IThreadRepository // domain 層のインターフェイス
	uuidGen    IUuidGenerator            // application 層のインターフェイス
	timeGetter ITimeGetter
}

func (u *CreateThreadUsecase) Exec(ctx context.Context, cmd CreateThreadCommand) (*CreateThreadOutput, error) {
	// 取得 → ドメインモデルに委譲 → 保存 の組み立てに徹する
}
```

`Exec()` の中に業務判断の分岐が増えてきたら、その判断は domain 層に移せないか見直す。

### エラー発生時のログ出力

`Exec()` は、依存（`IRepository` / `IQueryService` / 汎用インフラ）がエラーを返したとき、`ILogger`（application 層の汎用インターフェイス、実装は infra に置く）経由でログに出力してからそのままエラーを返してよい。これは「取得 → 委譲 → 保存」という Exec() の組み立てに業務ロジックを追加するものではなく、失敗の発生元をユースケース単位で特定できるようにするための観測用の副作用として扱う。handler 層でも同じエラーに対して別途ログを出す場合があるが、責務が異なる（usecase 層はユースケース単位の失敗検知、handler 層はレスポンスへの変換に付随する記録）ため、両方に出力があってよい。

## 層間の I/O ルール

| 経路 | やりとりに使う型 |
| --- | --- |
| handler → application | `XxCommand`（`application/` に配置） |
| application → handler | `XxOutput`（`application/` に配置） |
| usecase → IRepository | ドメインモデルを直接使ってよい |
| usecase → IQueryService | `application/` の DTO、またはプリミティブ型 |
| usecase → 汎用インフラ | `application/` の DTO、またはプリミティブ型 |

### DTO（XxCommand / XxOutput）

- 配置は `application/`。`handler/` 側や `domain/` 側には置かない。
- **ドメインモデル型をフィールドに持たない。** プリミティブ型と DTO のみで構成する軽量データモデルとする。
- 振る舞い（業務ロジック）を持たせない。ドメインモデルからの変換関数は持ってよい。
- ドメインモデル → `XxOutput` の変換は application 層で行う。handler にドメインモデルを渡さない。
- **`XxOutput` は、`IRepository` / `IQueryService` が返す型（ドメインモデル、`IQueryService` の DTO）をそのまま使い回さない。** フィールド構成が同じであっても、`XxOutput` 専用の型を別途定義し、`Exec()` の中で変換する。これにより、`IQueryService` 側の戻り値の型を変更しても `XxOutput` の型定義に直接波及しない。
- **`XxCommand` は、Web 層の型（`http.Header`、`*http.Request`、`echo.Context` など）をそのままプロパティに持たない。** ビジネスロジックが必要とする値だけを、handler 側でプリミティブ型に変換してから詰める。「念のため渡しておく」形で Web 層の値をまるごと持たせない。
- `XxCommand` のプロパティは、そのユースケースのビジネスロジックが実際に使う必要最小限の値に絞る。使うかどうか分からない値を先回りして持たせない。

良い例:

```go
// application/thread/dto.go
type CreateThreadCommand struct {
	Title    string
	AuthorID string
}

type CreateThreadOutput struct {
	ID        string
	Title     string
	CreatedAt time.Time
}
```

悪い例:

```go
type CreateThreadCommand struct {
	Title   string
	Header  http.Header // Web 層の型がそのまま漏れている。必要な値だけ抽出してプリミティブ型で渡す
}

type CreateThreadOutput struct {
	Thread *thread.Thread // ドメインモデルが handler に漏れている
}
```

## インフラ層に閉じ込める処理

次のいずれかに該当する処理は、必ず `infra/` のモジュールに置く。

- DB への書き込みがある
- ファイルの I/O が存在する
- 乱数要素が存在する（タイムスタンプ、UUID などのランダム文字列を含む）

したがって `domain/` `application/` のコードに以下を直接書かない。インターフェイス経由で呼ぶこと。

- `database/sql`、ORM のクエリ実行
- `os.Open` / `os.ReadFile` / `os.WriteFile` などのファイル操作
- `time.Now()`
- `uuid.New()`、`math/rand`、`crypto/rand`

### インフラにロジックを持たせない

`infra/` の実装は、外部リソースとのやりとりと型変換に徹する。業務判断となる条件分岐や計算を持たせない。

インフラにロジックが入りそうになったら、それを `domain/` に寄せられないか見直す。インフラのロジックが少ないほど、インフラのテストは不要になる。テストしづらいモジュールにロジックが溜まっている状態は、設計の見直し信号として扱う。

## インターフェイスの分類

| 種類 | 定義場所 | 用途 |
| --- | --- | --- |
| `IRepository` | `domain/` | ビジネスロジックに強く関係する永続化・取得。ドメインモデルをやりとりする |
| `IQueryService` | `application/` | 表示や集計のための参照。DTO を返す |
| その他（`ITimeGetter`, `IUuidGenerator`, `IRandomGenerator`, `IFileUploader` など） | `application/` | 汎用的な副作用の抽象化（値の生成・取得） |
| `ILogger` | `application/` | ログ出力という観測用の副作用の抽象化。値を生成せずExec()に副作用を戻さない点で上記とは性質が異なるが、配置場所・実装先（`infra/`）の扱いは同じ |

実装はすべて `infra/` に置く。インターフェイス名には `I` プレフィックスを付ける（Go の一般的な慣習とは異なるが、本プロジェクトの規約とする）。実装側は `PostgresThreadRepository` のように技術要素を含む名前にする。

### IRepository のメソッドはビジネスルールごとに閉じる

汎用的な検索条件を引数で受け取らない。ユースケースが必要とする単位でメソッドを切る。

```go
// good — domain/thread/repository.go
type IThreadRepository interface {
	FetchAllActiveThreads(ctx context.Context) ([]*Thread, error)
	Save(ctx context.Context, t *Thread) error
}

// bad — 「アクティブとは何か」という業務知識が呼び出し側に漏れる
FetchAllThreads(ctx context.Context, status string) ([]*Thread, error)
```

## ロジックの置き場所

- **アプリケーション固有のロジック（ビジネスロジック）** → `domain/`
- **汎用的なロジック** → `application/`
- **ドメインモデルで表現できないロジック** → DomainService（`domain/`）

DomainService は「複数の集約にまたがる」「単一のモデルに責務を割り当てると不自然になる」場合の手段であり、第一選択にしない。まず Entity や値オブジェクトのメソッドとして表現できないかを検討する。

## ドメインモデル

### Entity

- 識別子を持ち、同一性は識別子で判定する（フィールド値の一致では判定しない）
- フィールドは非公開にし、コンストラクタと振る舞いメソッドを通してのみ変更する
- コンストラクタで不変条件を検証し、不正な状態のインスタンスを作らせない

### Aggregate

- 整合性を保ちたいドメインモデルをまとめて管理する
- 参照・操作は集約ルート経由に限定する。内部の子モデルを直接取り出して更新しない
- リポジトリは集約ルート単位で用意する。子モデル専用のリポジトリは作らない
- トランザクション境界は集約単位に合わせる

## プレゼンテーション層（ハンドラ）

### ルーティング

パスは REST の原則に従う。リソース名は複数形の名詞とし、パスに動詞を含めない。

| メソッド | パス | 用途 |
| --- | --- | --- |
| GET | `/threads` | スレッド一覧取得 |
| GET | `/threads/:id` | スレッド詳細取得 |
| POST | `/threads` | スレッド作成 |

以降にエンドポイントを追加する場合も同じ原則（リソース単位の複数形パス、ネストする場合は親リソースのパスにぶら下げる）に従う。

### レスポンスのデータ構造

レスポンスの JSON は、配列やオブジェクトを直接ルートに置かず、**必ずルートにリソース名のキーを置く。** 一覧は複数形、単体は単数形のキーにする。

```json
// GET /threads
{
  "threads": [
    { "id": "1", "title": "..." }
  ]
}
```

```json
// GET /threads/:id
{
  "thread": {
    "id": "1",
    "title": "..."
  }
}
```

- このルート直下のキー付けは `handler/` の責務とする。`XxOutput`（`application/`）はキー付けを持たず、`handler/` 側でリソース名をキーにした `XxResponse` 型に詰め替えてから返す。
- レスポンスに含めるフィールドは、CLAUDE.md などの仕様で要求されている項目に限定する。**ドメインモデルや `XxOutput` が持つフィールドをそのまま返さない。** 仕様上不要なフィールドは `XxResponse` に持たせない。

```go
// handler/thread/response.go
type ThreadResponse struct {
	ID    string `json:"id"`
	Title string `json:"title"`
}

type ListThreadsResponse struct {
	Threads []ThreadResponse `json:"threads"`
}

type GetThreadResponse struct {
	Thread ThreadResponse `json:"thread"`
}
```

### エラーのステータスコード

ハンドラは、ユースケースが返したエラーの種類を判定し、次の基準でステータスコードを決める。

| エラーの種類 | ステータスコード |
| --- | --- |
| サーバー内部起因のエラー（想定外のエラーを含む） | 500 |
| ユーザー起因のエラー（入力値不正など） | 400 |
| ユースケースが「見つからない」エラーを返したとき | 404 |

ハンドラが種類を判定できるように、`domain/` にセンチネルエラー（例: `ErrThreadNotFound`）を定義し、`application/` はそれをラップして返す。ハンドラは `errors.Is` / `errors.As` で判定し、エラーメッセージの文字列比較で分岐しない。

```go
// domain/thread/errors.go
var ErrThreadNotFound = errors.New("thread not found")
```

```go
// handler/thread/handler.go
func (h *ThreadHandler) GetThread(c echo.Context) error {
	output, err := h.usecase.Exec(c.Request().Context(), cmd)
	if err != nil {
		return h.toHTTPError(err) // ステータスコードとメッセージへの変換に徹する
	}
	return c.JSON(http.StatusOK, GetThreadResponse{
		Thread: ThreadResponse{ID: output.ID, Title: output.Title},
	})
}
```

### エラーメッセージの変換

- ハンドラは、ユースケースから返ってきたエラーをそのままレスポンスに含めない。スタックトレースや `err.Error()` の文字列をそのまま返さない。
- エラーの種類ごとに、あらかじめ定義したユーザー向けの固定メッセージへ変換してから返す（例: `ErrThreadNotFound` → 「スレッドが存在しません。」）。変換前の詳細なエラーは、フロントエンドが参照しない前提とする。
- 変換ロジックは `handler/` に置く。`application/` や `domain/` にレスポンス用の文言を持たせない。
- 詳細なエラー情報を残したい場合はログにのみ出力し、レスポンスボディには含めない。

## 実装前チェックリスト

1. その処理は DB 書き込み / ファイル I/O / 乱数要素を含むか → 含むなら `infra/` + インターフェイス
2. そのロジックはアプリケーション固有か汎用か → 固有なら `domain/`、汎用なら `application/`
3. ドメインモデルのメソッドとして表現できるか → できないなら DomainService
4. 追加したリポジトリメソッドは、業務上の意味を名前で表現できているか
5. `XxOutput` にドメインモデル型や `IQueryService` の DTO 型がそのまま混ざっていないか（`XxOutput` 専用の型に変換しているか）
6. `infra/` の実装に業務判断の分岐が入っていないか
7. `domain/` の import に、他レイヤーや外部ライブラリが混ざっていないか
8. 追加したエンドポイントのパスは REST の原則（複数形リソース名、ネストは親パスにぶら下げる）に沿っているか
9. ハンドラが返すエラーは、エラー種別ごとに 400 / 404 / 500 へ正しくマッピングされているか
10. ハンドラのレスポンスに、変換前の詳細なエラー文字列やスタックトレースが含まれていないか
11. レスポンスの JSON がルートにリソース名のキーを持っているか（配列やオブジェクトを直接ルートに置いていないか）
12. レスポンスに含めるフィールドは、仕様で要求されている項目に限定されているか（ドメインモデルや `XxOutput` の全フィールドをそのまま返していないか）
13. `Exec()` は原子性を保つべき処理のまとまり単位で切られているか（一部だけ確定してよい処理が混在していないか）
14. `XxCommand` に Web 層の型（`http.Header` など）がそのまま含まれておらず、ビジネスロジックに必要な最小限のプリミティブ型のみになっているか
