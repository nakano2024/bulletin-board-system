# セキュリティレビュー（2026-11-10）

- 対象: `backend/`（ブランチ `feature/thread-details`、コミット `fdd7f86` 時点）
- 方法: コードの目視確認。実際の攻撃による検証は未実施。
- 分類: 緊急 / 重要 / 警告 / 注意

| # | 分類 | 概要 | 状態 |
|---|---|---|---|
| 1 | 重要 | `X-Forwarded-For` 偽装による他ユーザーIDのなりすまし | 未対応 |
| 2 | 重要 | Content-Type 未検証・SVG 許可による Stored XSS | 未対応 |
| 3 | 重要 | アップロードファイルサイズの上限なし | 未対応 |
| 4 | 警告 | スレッド本文の1000文字制限なし・リクエストボディ上限なし | 未対応 |
| 5 | 警告 | レートリミットなし | 未対応 |
| 6 | 注意 | ローカル環境の DB 等が LAN に公開されている | 未対応 |
| 7 | 注意 | 未アップロードのファイルでスレッドを作成できる | 未対応 |

緊急に該当するものはなし。

---

## 重要

### 1. `X-Forwarded-For` 偽装による他ユーザーIDのなりすまし

- 該当箇所: `backend/handler/thread/create_thread_handler.go`（`c.RealIP()`）、`backend/main.go`（`IPExtractor` 未設定）
- 内容: Echo v4.15.4 の `RealIP()` は `IPExtractor` が未設定だと `X-Forwarded-For` → `X-Real-IP` → `RemoteAddr` の順に無条件で信用する。
- 影響: 攻撃者は `X-Forwarded-For: <任意のIP>` を付けるだけで、そのIPに紐づく当日のユーザーIDで投稿できる（なりすまし）。また毎回異なるIPを指定すればID管理を回避できる。匿名掲示板ではIDが唯一の同一性の手がかりのため影響が大きい。
- 対策:
  - `e.IPExtractor = echo.ExtractIPFromXFFHeader(echo.TrustIPRange(<BFFのネットワーク>))` のように、BFF からのヘッダのみ信用する。
  - BFF 構成のため、BFF 側でも実クライアントIPを正しく `X-Forwarded-For` に付与して転送する必要がある。

### 2. Content-Type 未検証・SVG 許可による Stored XSS

- 該当箇所: `backend/domain/image_file/file_name.go`（許可拡張子に `.svg`）、`backend/domain/image_file/image_file.go`（`contentType` は空チェックのみ）
- 内容:
  - 拡張子は `.png` のまま `content_type: "text/html"` を申告でき、その Content-Type で署名付き PUT URL が発行される。
  - `.svg` が許可されており、`<script>` を含む SVG をアップロードできる。
  - バケットは public-read（`docker-compose.yml` の `rustfs-init`）。
- 影響: 画像URLを開いたユーザーのブラウザでスクリプトが実行される。画像配信元がアプリと同一オリジンの構成なら XSS、別オリジンでもフィッシングページの設置場所として悪用され得る。
- 対策:
  - `content_type` を許可リスト（`image/png`, `image/jpeg`, `image/gif`, `image/webp` 等）に限定し、拡張子との整合も検証する。
  - SVG を許可対象から外す。
  - 配信時に `X-Content-Type-Options: nosniff`、必要に応じて `Content-Disposition` を付与する。

### 3. アップロードファイルサイズの上限なし

- 該当箇所: `backend/domain/image_file/image_file.go`（`sizeBytes > 0` のみ検証）
- 内容: `size_bytes` に巨大な値を申告すると、そのサイズの署名付きURLが発行される。
- 影響: ストレージ枯渇・コスト増加。
- 対策: Domain 層で上限（例: 10MB）を設けて検証する。

---

## 警告

### 4. スレッド本文の1000文字制限なし・リクエストボディ上限なし

- 該当箇所: `backend/domain/thread/pending_thread.go`（空チェックのみ）、`backend/main.go`（`middleware.BodyLimit` 未使用）
- 内容: レスは `MaxReplyBodyLength`（1000文字）を検証しているが、スレッド本文は未検証。仕様「本文は1000文字以内で入力してください。」が未実装。加えてリクエストボディのサイズ上限もない。
- 影響: 数百MB規模の本文送信によるメモリ・DB 圧迫。
- 対策: `PendingThread` に文字数チェックを追加する。`e.Use(middleware.BodyLimit("1M"))` 等でリクエスト全体も制限する。

### 5. レートリミットなし

- 該当箇所: `POST /files`、`POST /threads`
- 内容: 両APIとも無制限に呼び出せる。特に `POST /files` は呼ぶたびに `files` レコードと署名付きURLが作られ、未使用ファイルの掃除処理もない。
- 影響: スパム投稿、不要データの蓄積。
- 対策: 起動済みの Redis を使い IP 単位のレートリミットを導入する。なお #1 が未対応だと IP 偽装で回避されるため、#1 の修正が前提。

---

## 注意

### 6. ローカル環境の DB 等が LAN に公開されている

- 該当箇所: `backend/docker-compose.yml`
- 内容: Postgres（`postgres/postgres`）、Redis（認証なし）、RustFS（`rustfsadmin`、コンソール 9001 含む）が `0.0.0.0` で公開されている。Docker のポート公開は ufw をバイパスする。
- 影響: 共有ネットワーク（カフェ等）上で、同一ネットワーク内の第三者から接続され得る。
- 対策: `"127.0.0.1:5432:5432"` のようにループバックにバインドする。

### 7. 未アップロードのファイルでスレッドを作成できる

- 該当箇所: `backend/domain/thread/thread_creation_service.go`、`backend/infra/postgres/thread_file_checker.go`
- 内容: `POST /files` でレコードだけ作成し実体を PUT していない `file_id` でもスレッドを作成できる。
- 影響: 画像が表示されないスレッドが作られる（セキュリティ影響は小）。
- 対策: スレッド作成時にストレージ上の実体存在（HEAD）を確認する、等。

---

## 推奨対応順

1 → 2 → 3 → 4（仕様の実装漏れでもあるため併せて対応）→ 5 → 6 → 7
