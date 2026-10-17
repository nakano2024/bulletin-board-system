package postgres_test

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/nakanokota/bulletin-board-system/backend/domain/image_file"
	domainthread "github.com/nakanokota/bulletin-board-system/backend/domain/thread"
	"github.com/nakanokota/bulletin-board-system/backend/infra/postgres"
)

type seedReply struct {
	id       string
	threadID string
	userID   string
	body     string
	// fileID is empty for a reply without an image (stored as NULL).
	fileID    string
	createdAt time.Time
}

func TestPostgresThreadRepository_FetchThreadDetail_正常系(t *testing.T) {
	pool := newTestPool(t)
	fixedTime := time.Date(2026, 10, 17, 12, 0, 0, 0, time.UTC)

	threadFile := seedFile{id: "file-thread", name: "thread.png"}
	aliveThread := seedThread{id: "thread-1", body: "hello", fileID: "file-thread", isAlive: true, createdAt: fixedTime}

	replyWithImage := mustNewReply(t, "reply-1", "user-2", "with image", mustFilePath(t, "reply.png"), fixedTime.Add(time.Minute))
	replyWithoutImage := mustNewReply(t, "reply-1", "user-2", "no image", nil, fixedTime.Add(time.Minute))
	replyOldest := mustNewReply(t, "reply-old", "user-2", "oldest", nil, fixedTime.Add(time.Minute))
	replyMiddle := mustNewReply(t, "reply-mid", "user-2", "middle", nil, fixedTime.Add(2*time.Minute))
	replyNewest := mustNewReply(t, "reply-new", "user-2", "newest", nil, fixedTime.Add(3*time.Minute))
	replySameA := mustNewReply(t, "reply-a", "user-2", "a", nil, fixedTime.Add(time.Minute))
	replySameB := mustNewReply(t, "reply-b", "user-2", "b", nil, fixedTime.Add(time.Minute))
	replyMine := mustNewReply(t, "reply-mine", "user-2", "mine", nil, fixedTime.Add(time.Minute))

	manySeedReplies := make([]seedReply, 0, 101)
	manyWantReplies := make([]*domainthread.Reply, 0, 101)
	for i := 100; i >= 0; i-- {
		id := fmt.Sprintf("reply-%03d", i)
		createdAt := fixedTime.Add(time.Duration(i+1) * time.Minute)
		manySeedReplies = append(manySeedReplies, seedReply{id: id, threadID: "thread-1", userID: "user-2", body: "body", createdAt: createdAt})
		manyWantReplies = append(manyWantReplies, mustNewReply(t, id, "user-2", "body", nil, createdAt))
	}

	tests := []struct {
		name        string
		seedFiles   []seedFile
		seedThreads []seedThread
		seedReplies []seedReply
		id          string
		want        *domainthread.ThreadDetail
	}{
		{
			name:        "is_aliveがtrueのスレッドを指定したとき、ID/UserID/Body/CreatedAtがthreadsの値で、FilePathがbasePathとfile_idで結合したfilesのnameから組み立てられたThreadDetailが返り、IsDead()がfalseであること",
			seedFiles:   []seedFile{threadFile},
			seedThreads: []seedThread{aliveThread},
			seedReplies: nil,
			id:          "thread-1",
			want:        mustNewThreadDetail(t, "thread-1", "hello", "thread.png", fixedTime, true, []*domainthread.Reply{}),
		},
		{
			name:        "is_aliveがfalseのスレッドを指定したとき、エラーにならず、IsDead()がtrueのThreadDetailが返ること",
			seedFiles:   []seedFile{threadFile},
			seedThreads: []seedThread{{id: "thread-1", body: "hello", fileID: "file-thread", isAlive: false, createdAt: fixedTime}},
			seedReplies: nil,
			id:          "thread-1",
			want:        mustNewThreadDetail(t, "thread-1", "hello", "thread.png", fixedTime, false, []*domainthread.Reply{}),
		},
		{
			name:        "スレッドにレスが1件も登録されていないとき、Replies()が空配列(nilではない)であること",
			seedFiles:   []seedFile{threadFile},
			seedThreads: []seedThread{aliveThread},
			seedReplies: nil,
			id:          "thread-1",
			want:        mustNewThreadDetail(t, "thread-1", "hello", "thread.png", fixedTime, true, []*domainthread.Reply{}),
		},
		{
			name:        "画像ありのレスが登録されているとき、ReplyのID/UserID/Body/CreatedAtがrepliesの値で、FilePathがbasePathとfile_idで結合したfilesのnameから組み立てられた値であること",
			seedFiles:   []seedFile{threadFile, {id: "file-reply", name: "reply.png"}},
			seedThreads: []seedThread{aliveThread},
			seedReplies: []seedReply{{id: "reply-1", threadID: "thread-1", userID: "user-2", body: "with image", fileID: "file-reply", createdAt: fixedTime.Add(time.Minute)}},
			id:          "thread-1",
			want:        mustNewThreadDetail(t, "thread-1", "hello", "thread.png", fixedTime, true, []*domainthread.Reply{replyWithImage}),
		},
		{
			name:        "画像なし(file_idがNULL)のレスが登録されているとき、ReplyのFilePath()がnilであること",
			seedFiles:   []seedFile{threadFile},
			seedThreads: []seedThread{aliveThread},
			seedReplies: []seedReply{{id: "reply-1", threadID: "thread-1", userID: "user-2", body: "no image", fileID: "", createdAt: fixedTime.Add(time.Minute)}},
			id:          "thread-1",
			want:        mustNewThreadDetail(t, "thread-1", "hello", "thread.png", fixedTime, true, []*domainthread.Reply{replyWithoutImage}),
		},
		{
			name:        "created_atが異なる複数のレスが登録順と無関係に登録されているとき、Replies()が投稿日時の降順(新しい順)であること",
			seedFiles:   []seedFile{threadFile},
			seedThreads: []seedThread{aliveThread},
			seedReplies: []seedReply{
				{id: "reply-mid", threadID: "thread-1", userID: "user-2", body: "middle", createdAt: fixedTime.Add(2 * time.Minute)},
				{id: "reply-old", threadID: "thread-1", userID: "user-2", body: "oldest", createdAt: fixedTime.Add(time.Minute)},
				{id: "reply-new", threadID: "thread-1", userID: "user-2", body: "newest", createdAt: fixedTime.Add(3 * time.Minute)},
			},
			id:   "thread-1",
			want: mustNewThreadDetail(t, "thread-1", "hello", "thread.png", fixedTime, true, []*domainthread.Reply{replyNewest, replyMiddle, replyOldest}),
		},
		{
			name:        "created_atが同じ複数のレスが登録されているとき、Replies()がidの降順であること",
			seedFiles:   []seedFile{threadFile},
			seedThreads: []seedThread{aliveThread},
			seedReplies: []seedReply{
				{id: "reply-a", threadID: "thread-1", userID: "user-2", body: "a", createdAt: fixedTime.Add(time.Minute)},
				{id: "reply-b", threadID: "thread-1", userID: "user-2", body: "b", createdAt: fixedTime.Add(time.Minute)},
			},
			id:   "thread-1",
			want: mustNewThreadDetail(t, "thread-1", "hello", "thread.png", fixedTime, true, []*domainthread.Reply{replySameB, replySameA}),
		},
		{
			name:      "別のスレッドのレスが登録されているとき、指定したスレッドのレスだけがReplies()に含まれること",
			seedFiles: []seedFile{threadFile, {id: "file-other", name: "other.png"}},
			seedThreads: []seedThread{
				aliveThread,
				{id: "thread-other", body: "other", fileID: "file-other", isAlive: true, createdAt: fixedTime},
			},
			seedReplies: []seedReply{
				{id: "reply-mine", threadID: "thread-1", userID: "user-2", body: "mine", createdAt: fixedTime.Add(time.Minute)},
				{id: "reply-other", threadID: "thread-other", userID: "user-2", body: "other", createdAt: fixedTime.Add(2 * time.Minute)},
			},
			id:   "thread-1",
			want: mustNewThreadDetail(t, "thread-1", "hello", "thread.png", fixedTime, true, []*domainthread.Reply{replyMine}),
		},
		{
			name:      "複数のスレッドが登録されているとき、指定したidのスレッドのID/UserID/Body/CreatedAt/FilePathが返ること",
			seedFiles: []seedFile{threadFile, {id: "file-other", name: "other.png"}},
			seedThreads: []seedThread{
				{id: "thread-other", body: "other", fileID: "file-other", isAlive: true, createdAt: fixedTime.Add(time.Minute)},
				aliveThread,
			},
			seedReplies: nil,
			id:          "thread-1",
			want:        mustNewThreadDetail(t, "thread-1", "hello", "thread.png", fixedTime, true, []*domainthread.Reply{}),
		},
		{
			name:        "スレッドに101件のレスが登録されているとき、Replies()が101件全て返ること",
			seedFiles:   []seedFile{threadFile},
			seedThreads: []seedThread{aliveThread},
			seedReplies: manySeedReplies,
			id:          "thread-1",
			want:        mustNewThreadDetail(t, "thread-1", "hello", "thread.png", fixedTime, true, manyWantReplies),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := context.Background()
			tx := beginTx(t, pool)
			seedFilesAndThreads(t, tx, tt.seedFiles, tt.seedThreads)
			seedReplies(t, tx, tt.seedReplies)

			sut := postgres.NewPostgresThreadRepository(tx, "thread_images/")
			got, err := sut.FetchThreadDetail(ctx, tt.id)

			require.NoError(t, err)
			assert.Equal(t, tt.want, normalizeThreadDetail(t, got))
		})
	}
}

func TestPostgresThreadRepository_FetchThreadDetail_異常系(t *testing.T) {
	pool := newTestPool(t)
	fixedTime := time.Date(2026, 10, 17, 12, 0, 0, 0, time.UTC)

	tests := []struct {
		name        string
		seedFiles   []seedFile
		seedThreads []seedThread
		id          string
		wantErr     error
	}{
		{
			name:        "存在しないidを指定したとき、ErrThreadNotFoundが返ること",
			seedFiles:   []seedFile{{id: "file-thread", name: "thread.png"}},
			seedThreads: []seedThread{{id: "thread-1", body: "hello", fileID: "file-thread", isAlive: true, createdAt: fixedTime}},
			id:          "thread-missing",
			wantErr:     domainthread.ErrThreadNotFound,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := context.Background()
			tx := beginTx(t, pool)
			seedFilesAndThreads(t, tx, tt.seedFiles, tt.seedThreads)

			sut := postgres.NewPostgresThreadRepository(tx, "thread_images/")
			_, err := sut.FetchThreadDetail(ctx, tt.id)

			require.ErrorIs(t, err, tt.wantErr)
		})
	}
}

func seedReplies(t *testing.T, tx pgx.Tx, replies []seedReply) {
	t.Helper()
	ctx := context.Background()

	for _, r := range replies {
		_, err := tx.Exec(ctx, "INSERT INTO replies (id, thread_id, user_id, body, file_id, created_at) VALUES ($1, $2, $3, $4, NULLIF($5, ''), $6)", r.id, r.threadID, r.userID, r.body, r.fileID, r.createdAt)
		require.NoError(t, err)
	}
}

// mustNewThreadDetail builds the expected ThreadDetail; seedFilesAndThreads always stores user_id "user-1".
func mustNewThreadDetail(t *testing.T, id, body, fileName string, createdAt time.Time, isAlive bool, replies []*domainthread.Reply) *domainthread.ThreadDetail {
	t.Helper()
	td, err := domainthread.NewThreadDetail(id, "user-1", body, mustFilePath(t, fileName), createdAt, isAlive, replies)
	require.NoError(t, err)
	return td
}

func mustNewReply(t *testing.T, id, userID, body string, filePath *image_file.FilePath, createdAt time.Time) *domainthread.Reply {
	t.Helper()
	r, err := domainthread.NewReply(id, userID, body, filePath, createdAt)
	require.NoError(t, err)
	return r
}

// normalizeThreadDetail rebuilds td with UTC times, since timestamps read back from Postgres carry the session's location.
func normalizeThreadDetail(t *testing.T, td *domainthread.ThreadDetail) *domainthread.ThreadDetail {
	t.Helper()
	replies := make([]*domainthread.Reply, 0, len(td.Replies()))
	for _, r := range td.Replies() {
		replies = append(replies, mustNewReply(t, r.ID(), r.UserID(), r.Body(), r.FilePath(), r.CreatedAt().UTC()))
	}
	normalized, err := domainthread.NewThreadDetail(td.ID(), td.UserID(), td.Body(), td.FilePath(), td.CreatedAt().UTC(), !td.IsDead(), replies)
	require.NoError(t, err)
	return normalized
}
