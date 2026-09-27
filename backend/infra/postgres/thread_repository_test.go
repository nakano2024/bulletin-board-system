package postgres_test

import (
	"context"
	"os"
	"testing"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	domainthread "github.com/nakanokota/bulletin-board-system/backend/domain/thread"
	"github.com/nakanokota/bulletin-board-system/backend/infra/postgres"
)

type seedFile struct {
	id   string
	name string
}

type seedThread struct {
	id        string
	body      string
	fileID    string
	isAlive   bool
	createdAt time.Time
}

func TestPostgresThreadRepository_FetchActiveThreadListNewestFirst_正常系(t *testing.T) {
	pool := newTestPool(t)
	fixedTime := time.Date(2026, 9, 3, 12, 0, 0, 0, time.UTC)

	threadHello, _ := domainthread.NewThread("thread-1", "hello", mustFilePath(t, "sample.png"), fixedTime)
	threadCat, _ := domainthread.NewThread("thread-cat", "cat", mustFilePath(t, "cat.png"), fixedTime)
	threadDog, _ := domainthread.NewThread("thread-dog", "dog", mustFilePath(t, "dog.jpg"), fixedTime.Add(time.Minute))
	threadFirst, _ := domainthread.NewThread("thread-1", "first", mustFilePath(t, "first.png"), fixedTime)
	threadSecond, _ := domainthread.NewThread("thread-2", "second", mustFilePath(t, "second.png"), fixedTime.Add(time.Minute))
	threadOldest, _ := domainthread.NewThread("thread-old", "oldest", mustFilePath(t, "old.png"), fixedTime)
	threadMiddle, _ := domainthread.NewThread("thread-mid", "middle", mustFilePath(t, "mid.png"), fixedTime.Add(time.Minute))
	threadNewest, _ := domainthread.NewThread("thread-new", "newest", mustFilePath(t, "new.png"), fixedTime.Add(2*time.Minute))
	threadAlive, _ := domainthread.NewThread("thread-1", "alive", mustFilePath(t, "alive.png"), fixedTime)

	tests := []struct {
		name        string
		seedFiles   []seedFile
		seedThreads []seedThread
		want        []*domainthread.Thread
	}{
		{
			name:        "スレッドが1件登録されているとき、ThreadのFilePathが、basePathと、file_idで結合したfilesのnameから組み立てられた値であること",
			seedFiles:   []seedFile{{id: "file-1", name: "sample.png"}},
			seedThreads: []seedThread{{id: "thread-1", body: "hello", fileID: "file-1", isAlive: true, createdAt: fixedTime}},
			want:        []*domainthread.Thread{threadHello},
		},
		{
			name: "複数のスレッドがそれぞれ異なるファイルを参照しているとき、各ThreadのFilePathが、それぞれのfile_idに対応するfilesのnameから組み立てられた値であること",
			seedFiles: []seedFile{
				{id: "file-cat", name: "cat.png"},
				{id: "file-dog", name: "dog.jpg"},
			},
			seedThreads: []seedThread{
				{id: "thread-cat", body: "cat", fileID: "file-cat", isAlive: true, createdAt: fixedTime},
				{id: "thread-dog", body: "dog", fileID: "file-dog", isAlive: true, createdAt: fixedTime.Add(time.Minute)},
			},
			want: []*domainthread.Thread{threadDog, threadCat},
		},
		{
			name: "スレッドが複数件登録されているとき、登録件数と同じ件数のThreadが返ること",
			seedFiles: []seedFile{
				{id: "file-1", name: "first.png"},
				{id: "file-2", name: "second.png"},
			},
			seedThreads: []seedThread{
				{id: "thread-1", body: "first", fileID: "file-1", isAlive: true, createdAt: fixedTime},
				{id: "thread-2", body: "second", fileID: "file-2", isAlive: true, createdAt: fixedTime.Add(time.Minute)},
			},
			want: []*domainthread.Thread{threadSecond, threadFirst},
		},
		{
			name: "createdAtが異なる複数のスレッドが登録順と無関係に登録されているとき、作成日時の降順(新しい順)で返ること",
			seedFiles: []seedFile{
				{id: "file-mid", name: "mid.png"},
				{id: "file-old", name: "old.png"},
				{id: "file-new", name: "new.png"},
			},
			seedThreads: []seedThread{
				{id: "thread-mid", body: "middle", fileID: "file-mid", isAlive: true, createdAt: fixedTime.Add(time.Minute)},
				{id: "thread-old", body: "oldest", fileID: "file-old", isAlive: true, createdAt: fixedTime},
				{id: "thread-new", body: "newest", fileID: "file-new", isAlive: true, createdAt: fixedTime.Add(2 * time.Minute)},
			},
			want: []*domainthread.Thread{threadNewest, threadMiddle, threadOldest},
		},
		{
			name:        "スレッドが1件も登録されていないとき、ThreadListのThreads()が空スライスであること",
			seedFiles:   nil,
			seedThreads: nil,
			want:        []*domainthread.Thread{},
		},
		{
			name: "is_aliveがfalseのスレッドを含むとき、そのスレッドはThreadListに含まれないこと",
			seedFiles: []seedFile{
				{id: "file-alive", name: "alive.png"},
				{id: "file-dead", name: "dead.png"},
			},
			seedThreads: []seedThread{
				{id: "thread-1", body: "alive", fileID: "file-alive", isAlive: true, createdAt: fixedTime},
				{id: "thread-2", body: "dead", fileID: "file-dead", isAlive: false, createdAt: fixedTime.Add(time.Minute)},
			},
			want: []*domainthread.Thread{threadAlive},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := context.Background()
			tx := beginTx(t, pool)
			seedFilesAndThreads(t, tx, tt.seedFiles, tt.seedThreads)

			sut := postgres.NewPostgresThreadRepository(tx, "thread_images/")
			got, err := sut.FetchActiveThreadListNewestFirst(ctx)

			require.NoError(t, err)
			assert.Equal(t, normalizeThreads(tt.want), normalizeThreads(got.Threads()))
		})
	}
}

func TestPostgresThreadRepository_CreateThread_正常系_返り値(t *testing.T) {
	pool := newTestPool(t)

	tests := []struct {
		name          string
		seedFiles     []seedFile
		pendingThread *domainthread.PendingThread
		wantBody      string
		wantFilePath  string
	}{
		{
			name:          "登録済みでどのスレッドにも使われていないファイルのfile_idを持つPendingThreadを渡したとき、id・createdAtが採番され、FilePathがbasePathとそのファイルのnameから組み立てられたThreadが返ること",
			seedFiles:     []seedFile{{id: "file-1", name: "sample.png"}},
			pendingThread: mustNewPendingThread(t, "user-1", "hello", "file-1"),
			wantBody:      "hello",
			wantFilePath:  "thread_images/sample.png",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := context.Background()
			tx := beginTx(t, pool)
			seedFilesAndThreads(t, tx, tt.seedFiles, nil)

			sut := postgres.NewPostgresThreadRepository(tx, "thread_images/")
			got, err := sut.CreateThread(ctx, tt.pendingThread)

			require.NoError(t, err)
			assert.NotEmpty(t, got.ID())
			assert.False(t, got.CreatedAt().IsZero())
			assert.Equal(t, tt.wantBody, got.Body())
			assert.Equal(t, tt.wantFilePath, got.FilePathValue())
		})
	}
}

func TestPostgresThreadRepository_CreateThread_正常系_保存内容(t *testing.T) {
	pool := newTestPool(t)

	tests := []struct {
		name          string
		seedFiles     []seedFile
		pendingThread *domainthread.PendingThread
		wantUserID    string
		wantBody      string
		wantFileID    string
	}{
		{
			name:          "登録済みでどのスレッドにも使われていないファイルのfile_idを持つPendingThreadを渡したとき、threadsにPendingThreadのuser_id・body・file_idが保存されること",
			seedFiles:     []seedFile{{id: "file-1", name: "sample.png"}},
			pendingThread: mustNewPendingThread(t, "user-1", "hello", "file-1"),
			wantUserID:    "user-1",
			wantBody:      "hello",
			wantFileID:    "file-1",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := context.Background()
			tx := beginTx(t, pool)
			seedFilesAndThreads(t, tx, tt.seedFiles, nil)

			sut := postgres.NewPostgresThreadRepository(tx, "thread_images/")
			got, err := sut.CreateThread(ctx, tt.pendingThread)
			require.NoError(t, err)

			var userID, body, fileID string
			err = tx.QueryRow(ctx, "SELECT user_id, body, file_id FROM threads WHERE id = $1", got.ID()).Scan(&userID, &body, &fileID)
			require.NoError(t, err)
			assert.Equal(t, tt.wantUserID, userID)
			assert.Equal(t, tt.wantBody, body)
			assert.Equal(t, tt.wantFileID, fileID)
		})
	}
}

func TestPostgresThreadRepository_CreateThread_異常系(t *testing.T) {
	pool := newTestPool(t)
	fixedTime := time.Date(2026, 9, 3, 12, 0, 0, 0, time.UTC)

	tests := []struct {
		name          string
		seedFiles     []seedFile
		seedThreads   []seedThread
		pendingThread *domainthread.PendingThread
		wantErr       error
	}{
		{
			name:          "filesに存在しないfile_idを持つPendingThreadを渡したとき、ErrFileNotFoundが返ること",
			seedFiles:     nil,
			seedThreads:   nil,
			pendingThread: mustNewPendingThread(t, "user-1", "hello", "file-missing"),
			wantErr:       domainthread.ErrFileNotFound,
		},
		{
			name:          "他のスレッドが既に参照しているfile_idを持つPendingThreadを渡したとき、ErrFileAlreadyUsedが返ること",
			seedFiles:     []seedFile{{id: "file-1", name: "sample.png"}},
			seedThreads:   []seedThread{{id: "thread-1", body: "existing", fileID: "file-1", isAlive: true, createdAt: fixedTime}},
			pendingThread: mustNewPendingThread(t, "user-1", "hello", "file-1"),
			wantErr:       domainthread.ErrFileAlreadyUsed,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := context.Background()
			tx := beginTx(t, pool)
			seedFilesAndThreads(t, tx, tt.seedFiles, tt.seedThreads)

			sut := postgres.NewPostgresThreadRepository(tx, "thread_images/")
			_, err := sut.CreateThread(ctx, tt.pendingThread)

			require.ErrorIs(t, err, tt.wantErr)
		})
	}
}

func beginTx(t *testing.T, pool *pgxpool.Pool) pgx.Tx {
	t.Helper()
	ctx := context.Background()

	tx, err := pool.Begin(ctx)
	require.NoError(t, err)
	t.Cleanup(func() { tx.Rollback(ctx) })

	return tx
}

func seedFilesAndThreads(t *testing.T, tx pgx.Tx, files []seedFile, threads []seedThread) {
	t.Helper()
	ctx := context.Background()

	for _, f := range files {
		_, err := tx.Exec(ctx, "INSERT INTO files (id, name, content_type, size_bytes) VALUES ($1, $2, $3, $4)", f.id, f.name, "image/png", 1024)
		require.NoError(t, err)
	}
	for _, s := range threads {
		_, err := tx.Exec(ctx, "INSERT INTO threads (id, user_id, body, file_id, is_alive, created_at) VALUES ($1, $2, $3, $4, $5, $6)", s.id, "user-1", s.body, s.fileID, s.isAlive, s.createdAt)
		require.NoError(t, err)
	}
}

func mustFilePath(t *testing.T, fileName string) *domainthread.FilePath {
	t.Helper()
	fp, err := domainthread.NewFilePath("thread_images/", fileName)
	require.NoError(t, err)
	return fp
}

func mustNewPendingThread(t *testing.T, userID, body, fileID string) *domainthread.PendingThread {
	t.Helper()
	pt, err := domainthread.NewPendingThread(userID, body, fileID)
	require.NoError(t, err)
	return pt
}

func normalizeThreads(threads []*domainthread.Thread) []*domainthread.Thread {
	normalized := make([]*domainthread.Thread, len(threads))
	for i, th := range threads {
		normalized[i] = mustNewThreadFrom(th)
	}
	return normalized
}

func mustNewThreadFrom(th *domainthread.Thread) *domainthread.Thread {
	normalized, _ := domainthread.NewThread(th.ID(), th.Body(), th.FilePath(), th.CreatedAt().UTC())
	return normalized
}

func newTestPool(t *testing.T) *pgxpool.Pool {
	t.Helper()

	pool, err := pgxpool.New(context.Background(), os.Getenv("TEST_DATABASE_URL"))
	require.NoError(t, err)
	t.Cleanup(pool.Close)

	return pool
}
