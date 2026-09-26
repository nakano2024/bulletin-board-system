package postgres_test

import (
	"context"
	"os"
	"testing"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	domainthread "github.com/nakanokota/bulletin-board-system/backend/domain/thread"
	"github.com/nakanokota/bulletin-board-system/backend/infra/postgres"
)

type seedThread struct {
	id        string
	body      string
	fileName  *string
	isAlive   bool
	createdAt time.Time
}

func TestPostgresThreadRepository_FetchActiveThreadListNewestFirst_正常系(t *testing.T) {
	pool := newTestPool(t)
	fixedTime := time.Date(2026, 9, 3, 12, 0, 0, 0, time.UTC)
	fileName := "sample.png"

	filePath, _ := domainthread.NewFilePath("thread_images/", "sample.png")
	threadHello, _ := domainthread.NewThread("thread-1", "hello", filePath, fixedTime)
	threadFirst, _ := domainthread.NewThread("thread-1", "first", filePath, fixedTime)
	threadSecond, _ := domainthread.NewThread("thread-2", "second", filePath, fixedTime.Add(time.Minute))
	threadOldest, _ := domainthread.NewThread("thread-old", "oldest", filePath, fixedTime)
	threadMiddle, _ := domainthread.NewThread("thread-mid", "middle", filePath, fixedTime.Add(time.Minute))
	threadNewest, _ := domainthread.NewThread("thread-new", "newest", filePath, fixedTime.Add(2*time.Minute))
	threadAlive, _ := domainthread.NewThread("thread-1", "alive", filePath, fixedTime)

	tests := []struct {
		name string
		seed []seedThread
		want []*domainthread.Thread
	}{
		{
			name: "スレッドが1件登録されている(file_nameあり)とき、ThreadListに含まれるThreadのFilePathがbasePathとfile_nameから再構築された値であること",
			seed: []seedThread{
				{id: "thread-1", body: "hello", fileName: &fileName, isAlive: true, createdAt: fixedTime},
			},
			want: []*domainthread.Thread{threadHello},
		},
		{
			name: "スレッドが複数件登録されている(いずれもfile_nameあり)とき、登録件数と同じ件数のThreadが返ること",
			seed: []seedThread{
				{id: "thread-1", body: "first", fileName: &fileName, isAlive: true, createdAt: fixedTime},
				{id: "thread-2", body: "second", fileName: &fileName, isAlive: true, createdAt: fixedTime.Add(time.Minute)},
			},
			want: []*domainthread.Thread{threadSecond, threadFirst},
		},
		{
			name: "createdAtが異なる複数のスレッドが登録順と無関係に登録されているとき、作成日時の降順(新しい順)で返ること",
			seed: []seedThread{
				{id: "thread-mid", body: "middle", fileName: &fileName, isAlive: true, createdAt: fixedTime.Add(time.Minute)},
				{id: "thread-old", body: "oldest", fileName: &fileName, isAlive: true, createdAt: fixedTime},
				{id: "thread-new", body: "newest", fileName: &fileName, isAlive: true, createdAt: fixedTime.Add(2 * time.Minute)},
			},
			want: []*domainthread.Thread{threadNewest, threadMiddle, threadOldest},
		},
		{
			name: "スレッドが1件も登録されていないとき、ThreadListのThreads()が空スライスであること",
			seed: nil,
			want: []*domainthread.Thread{},
		},
		{
			name: "is_aliveがfalseのスレッドを含むとき、そのスレッドはThreadListに含まれないこと",
			seed: []seedThread{
				{id: "thread-1", body: "alive", fileName: &fileName, isAlive: true, createdAt: fixedTime},
				{id: "thread-2", body: "dead", fileName: &fileName, isAlive: false, createdAt: fixedTime.Add(time.Minute)},
			},
			want: []*domainthread.Thread{threadAlive},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := context.Background()

			tx, err := pool.Begin(ctx)
			require.NoError(t, err)
			t.Cleanup(func() { tx.Rollback(ctx) })

			for _, s := range tt.seed {
				_, err := tx.Exec(ctx, "INSERT INTO threads (id, user_id, body, file_name, is_alive, created_at) VALUES ($1, $2, $3, $4, $5, $6)", s.id, "user-1", s.body, s.fileName, s.isAlive, s.createdAt)
				require.NoError(t, err)
			}

			sut := postgres.NewPostgresThreadRepository(tx, "thread_images/")
			got, err := sut.FetchActiveThreadListNewestFirst(ctx)

			require.NoError(t, err)
			assert.Equal(t, normalizeThreads(tt.want), normalizeThreads(got.Threads()))
		})
	}
}

func TestPostgresThreadRepository_CreateThread_正常系(t *testing.T) {
	pool := newTestPool(t)
	fileName, err := domainthread.NewFileName("sample.png")
	require.NoError(t, err)

	tests := []struct {
		name          string
		pendingThread *domainthread.PendingThread
	}{
		{
			name:          "有効なPendingThreadを渡したとき、id・createdAtが採番されたThreadが返り、DBにも保存されること",
			pendingThread: mustNewPendingThread(t, "user-1", "hello", fileName),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := context.Background()

			tx, err := pool.Begin(ctx)
			require.NoError(t, err)
			t.Cleanup(func() { tx.Rollback(ctx) })

			sut := postgres.NewPostgresThreadRepository(tx, "thread_images/")
			got, err := sut.CreateThread(ctx, tt.pendingThread)

			require.NoError(t, err)
			assert.NotEmpty(t, got.ID())
			assert.Equal(t, tt.pendingThread.Body(), got.Body())
			assert.Equal(t, "thread_images/sample.png", got.FilePathValue())
			assert.False(t, got.CreatedAt().IsZero())

			var userID, body string
			err = tx.QueryRow(ctx, "SELECT user_id, body FROM threads WHERE id = $1", got.ID()).Scan(&userID, &body)
			require.NoError(t, err)
			assert.Equal(t, tt.pendingThread.UserID(), userID)
			assert.Equal(t, tt.pendingThread.Body(), body)
		})
	}
}

func mustNewPendingThread(t *testing.T, userID, body string, fileName *domainthread.FileName) *domainthread.PendingThread {
	t.Helper()
	pt, err := domainthread.NewPendingThread(userID, body, fileName)
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
