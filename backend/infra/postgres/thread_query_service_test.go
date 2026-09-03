package postgres_test

import (
	"cmp"
	"context"
	"os"
	"testing"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	applicationthread "github.com/nakanokota/bulletin-board-system/backend/application/thread"
	"github.com/nakanokota/bulletin-board-system/backend/infra/postgres"
)

type seedThread struct {
	id        string
	body      string
	imagePath *string
	createdAt time.Time
}

func TestPostgresThreadQueryService_FetchThreadList_正常系(t *testing.T) {
	pool := newTestPool(t)
	fixedTime := time.Date(2026, 9, 3, 12, 0, 0, 0, time.UTC)
	imagePath := "images/thread-1.png"

	tests := []struct {
		name string
		seed []seedThread
		want []applicationthread.ThreadListItem
	}{
		{
			name: "スレッドが1件登録されているとき、そのスレッドのID/Body/ImagePath/CreatedAtを保持した要素が1件返ること",
			seed: []seedThread{
				{id: "thread-1", body: "hello", imagePath: &imagePath, createdAt: fixedTime},
			},
			want: []applicationthread.ThreadListItem{
				{ID: "thread-1", Body: "hello", ImagePath: &imagePath, CreatedAt: fixedTime},
			},
		},
		{
			name: "スレッドが複数件登録されているとき、登録件数と同じ件数の要素(image_pathがnullの要素含む)が返ること",
			seed: []seedThread{
				{id: "thread-1", body: "first", imagePath: nil, createdAt: fixedTime},
				{id: "thread-2", body: "second", imagePath: &imagePath, createdAt: fixedTime.Add(time.Minute)},
			},
			want: []applicationthread.ThreadListItem{
				{ID: "thread-1", Body: "first", ImagePath: nil, CreatedAt: fixedTime},
				{ID: "thread-2", Body: "second", ImagePath: &imagePath, CreatedAt: fixedTime.Add(time.Minute)},
			},
		},
		{
			name: "スレッドが1件も登録されていないとき、空スライスが返ること",
			seed: nil,
			want: []applicationthread.ThreadListItem{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := context.Background()

			tx, err := pool.Begin(ctx)
			require.NoError(t, err)
			t.Cleanup(func() { tx.Rollback(ctx) })

			for _, s := range tt.seed {
				_, err := tx.Exec(ctx, "INSERT INTO threads (id, body, image_path, created_at) VALUES ($1, $2, $3, $4)", s.id, s.body, s.imagePath, s.createdAt)
				require.NoError(t, err)
			}

			sut := postgres.NewPostgresThreadQueryService(tx)
			got, err := sut.FetchThreadList(ctx)

			require.NoError(t, err)
			assert.Equal(t, normalizeThreadListItems(tt.want), normalizeThreadListItems(got))
		})
	}
}

func normalizeThreadListItems(items []applicationthread.ThreadListItem) []applicationthread.ThreadListItem {
	normalized := make([]applicationthread.ThreadListItem, len(items))
	for i, item := range items {
		normalized[i] = applicationthread.ThreadListItem{
			ID:        item.ID,
			Body:      item.Body,
			ImagePath: item.ImagePath,
			CreatedAt: item.CreatedAt.UTC(),
		}
	}
	return normalized
}

func newTestPool(t *testing.T) *pgxpool.Pool {
	t.Helper()

	dsn := cmp.Or(os.Getenv("TEST_DATABASE_URL"), "postgres://postgres:postgres@localhost:5432/bulletin_board?sslmode=disable")

	pool, err := pgxpool.New(context.Background(), dsn)
	require.NoError(t, err)
	t.Cleanup(pool.Close)

	return pool
}
