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

func TestPostgresThreadRepository_FetchActiveThreadList_正常系(t *testing.T) {
	pool := newTestPool(t)
	fixedTime := time.Date(2026, 9, 3, 12, 0, 0, 0, time.UTC)
	fileName := "sample123"

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
			want: []*domainthread.Thread{
				mustNewThread(t, "thread-1", "hello", mustNewFilePath(t, "thread_images/", "sample123"), fixedTime),
			},
		},
		{
			name: "スレッドが複数件登録されている(いずれもfile_nameあり)とき、登録件数と同じ件数のThreadが返ること",
			seed: []seedThread{
				{id: "thread-1", body: "first", fileName: &fileName, isAlive: true, createdAt: fixedTime},
				{id: "thread-2", body: "second", fileName: &fileName, isAlive: true, createdAt: fixedTime.Add(time.Minute)},
			},
			want: []*domainthread.Thread{
				mustNewThread(t, "thread-1", "first", mustNewFilePath(t, "thread_images/", "sample123"), fixedTime),
				mustNewThread(t, "thread-2", "second", mustNewFilePath(t, "thread_images/", "sample123"), fixedTime.Add(time.Minute)),
			},
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
			want: []*domainthread.Thread{
				mustNewThread(t, "thread-1", "alive", mustNewFilePath(t, "thread_images/", "sample123"), fixedTime),
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := context.Background()

			tx, err := pool.Begin(ctx)
			require.NoError(t, err)
			t.Cleanup(func() { tx.Rollback(ctx) })

			for _, s := range tt.seed {
				_, err := tx.Exec(ctx, "INSERT INTO threads (id, body, file_name, is_alive, created_at) VALUES ($1, $2, $3, $4, $5)", s.id, s.body, s.fileName, s.isAlive, s.createdAt)
				require.NoError(t, err)
			}

			sut := postgres.NewPostgresThreadRepository(tx, "thread_images/")
			got, err := sut.FetchActiveThreadList(ctx)

			require.NoError(t, err)
			assert.Equal(t, normalizeThreads(tt.want), normalizeThreads(got.Threads()))
		})
	}
}

func TestPostgresThreadRepository_FetchActiveThreadList_異常系(t *testing.T) {
	pool := newTestPool(t)
	fixedTime := time.Date(2026, 9, 3, 12, 0, 0, 0, time.UTC)

	tests := []struct {
		name    string
		seed    []seedThread
		wantErr error
	}{
		{
			name: "file_name列の値が半角英数字・アンダースコア・ハイフン以外の文字を含む(FilePathの命名規則に合致しない)とき、FetchActiveThreadListがエラーを返すこと",
			seed: []seedThread{
				{id: "thread-1", body: "hello", fileName: strPtr("invalid@name.png"), isAlive: true, createdAt: fixedTime},
			},
			wantErr: domainthread.ErrFilePathFileNameInvalid,
		},
		{
			name: "file_name列がNULL(画像未添付)のスレッドを含むとき、FetchActiveThreadListがErrThreadFilePathMissingを返すこと",
			seed: []seedThread{
				{id: "thread-1", body: "hello", fileName: nil, isAlive: true, createdAt: fixedTime},
			},
			wantErr: domainthread.ErrThreadFilePathMissing,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := context.Background()

			tx, err := pool.Begin(ctx)
			require.NoError(t, err)
			t.Cleanup(func() { tx.Rollback(ctx) })

			for _, s := range tt.seed {
				_, err := tx.Exec(ctx, "INSERT INTO threads (id, body, file_name, is_alive, created_at) VALUES ($1, $2, $3, $4, $5)", s.id, s.body, s.fileName, s.isAlive, s.createdAt)
				require.NoError(t, err)
			}

			sut := postgres.NewPostgresThreadRepository(tx, "thread_images/")
			_, err = sut.FetchActiveThreadList(ctx)

			require.ErrorIs(t, err, tt.wantErr)
		})
	}
}

func mustNewFilePath(t *testing.T, basePath, fileName string) *domainthread.FilePath {
	t.Helper()
	filePath, err := domainthread.NewFilePath(basePath, fileName)
	require.NoError(t, err)
	return filePath
}

func mustNewThread(t *testing.T, id, body string, filePath *domainthread.FilePath, createdAt time.Time) *domainthread.Thread {
	t.Helper()
	th, err := domainthread.NewThread(id, body, filePath, createdAt)
	require.NoError(t, err)
	return th
}

func strPtr(s string) *string {
	return &s
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
