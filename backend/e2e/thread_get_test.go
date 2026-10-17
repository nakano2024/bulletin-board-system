package e2e

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/labstack/echo/v4"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	applicationthread "github.com/nakanokota/bulletin-board-system/backend/application/usecase/thread"
	handlerthread "github.com/nakanokota/bulletin-board-system/backend/handler/thread"
	infralog "github.com/nakanokota/bulletin-board-system/backend/infra/log"
	"github.com/nakanokota/bulletin-board-system/backend/infra/postgres"
)

// TestE2E_GetThread_正常系 wires the real Handler -> Usecase -> Repository -> DB stack (no mocks) and checks
// GET /threads/:id against one seeded thread with two replies. The replies (one with an image, one without) are
// included because the NULL file_id -> nil FilePath -> empty image_path hand-off only shows up across all layers.
func TestE2E_GetThread_正常系(t *testing.T) {
	ctx := context.Background()

	pool, err := pgxpool.New(ctx, os.Getenv("E2E_DATABASE_URL"))
	require.NoError(t, err)
	t.Cleanup(pool.Close)

	fixedTime := time.Date(2026, 10, 17, 12, 0, 0, 0, time.UTC)
	_, err = pool.Exec(ctx, "INSERT INTO files (id, name, content_type, size_bytes) VALUES ($1, $2, $3, $4), ($5, $6, $7, $8)",
		"file-e2e-get-thread", "thread.png", "image/png", 1024,
		"file-e2e-get-reply", "reply.png", "image/png", 1024)
	require.NoError(t, err)
	_, err = pool.Exec(ctx,
		"INSERT INTO threads (id, user_id, body, file_id, is_alive, created_at) VALUES ($1, $2, $3, $4, $5, $6)",
		"thread-e2e-get", "user-1", "hello e2e", "file-e2e-get-thread", true, fixedTime)
	require.NoError(t, err)
	_, err = pool.Exec(ctx,
		"INSERT INTO replies (id, thread_id, user_id, body, file_id, created_at) VALUES ($1, $2, $3, $4, $5, $6), ($7, $8, $9, $10, NULL, $11)",
		"reply-e2e-with-image", "thread-e2e-get", "user-2", "with image", "file-e2e-get-reply", fixedTime.Add(time.Minute),
		"reply-e2e-no-image", "thread-e2e-get", "user-3", "no image", fixedTime.Add(2*time.Minute))
	require.NoError(t, err)
	t.Cleanup(func() {
		_, _ = pool.Exec(context.Background(), "DELETE FROM replies WHERE thread_id = $1", "thread-e2e-get")
		_, _ = pool.Exec(context.Background(), "DELETE FROM threads WHERE id = $1", "thread-e2e-get")
		_, _ = pool.Exec(context.Background(), "DELETE FROM files WHERE id IN ($1, $2)", "file-e2e-get-thread", "file-e2e-get-reply")
	})

	threadRepository := postgres.NewPostgresThreadRepository(pool, "thread_images/")
	logger := infralog.NewStdLogger()
	getThreadUsecase := applicationthread.NewGetThreadUsecase(threadRepository, logger)
	getThreadHandler := handlerthread.NewGetThreadHandler(getThreadUsecase)

	e := echo.New()
	req := httptest.NewRequest(http.MethodGet, "/threads/thread-e2e-get", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.SetParamNames("id")
	c.SetParamValues("thread-e2e-get")

	err = getThreadHandler.GetThread(c)
	require.NoError(t, err)

	assert.Equal(t, http.StatusOK, rec.Code)

	var got handlerthread.GetThreadResponse
	require.NoError(t, json.Unmarshal(rec.Body.Bytes(), &got))

	assert.Equal(t, "thread-e2e-get", got.Thread.ID)
	assert.Equal(t, "user-1", got.Thread.UserID)
	assert.Equal(t, "hello e2e", got.Thread.Body)
	assert.Equal(t, "thread_images/thread.png", got.Thread.ImagePath)
	assert.True(t, got.Thread.CreatedAt.Equal(fixedTime))

	require.Len(t, got.Thread.Replies, 2)
	assert.Equal(t, "reply-e2e-no-image", got.Thread.Replies[0].ID)
	assert.Equal(t, "user-3", got.Thread.Replies[0].UserID)
	assert.Equal(t, "no image", got.Thread.Replies[0].Body)
	assert.Equal(t, "", got.Thread.Replies[0].ImagePath)
	assert.True(t, got.Thread.Replies[0].CreatedAt.Equal(fixedTime.Add(2*time.Minute)))
	assert.Equal(t, "reply-e2e-with-image", got.Thread.Replies[1].ID)
	assert.Equal(t, "user-2", got.Thread.Replies[1].UserID)
	assert.Equal(t, "with image", got.Thread.Replies[1].Body)
	assert.Equal(t, "thread_images/reply.png", got.Thread.Replies[1].ImagePath)
	assert.True(t, got.Thread.Replies[1].CreatedAt.Equal(fixedTime.Add(time.Minute)))
}
