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

	applicationthread "github.com/nakanokota/bulletin-board-system/backend/application/thread"
	handlerthread "github.com/nakanokota/bulletin-board-system/backend/handler/thread"
	infralog "github.com/nakanokota/bulletin-board-system/backend/infra/log"
	"github.com/nakanokota/bulletin-board-system/backend/infra/postgres"
)

// TestE2E_GetThreads_正常系 wires the real Handler -> Usecase -> Repository -> DB stack
// (no mocks) and checks GET /threads against one seeded row. Kept to a single case
// on purpose: full-stack DB tests are the most brittle in the suite.
func TestE2E_GetThreads_正常系(t *testing.T) {
	ctx := context.Background()

	pool, err := pgxpool.New(ctx, os.Getenv("E2E_DATABASE_URL"))
	require.NoError(t, err)
	t.Cleanup(pool.Close)

	fixedTime := time.Date(2026, 9, 3, 12, 0, 0, 0, time.UTC)
	_, err = pool.Exec(ctx,
		"INSERT INTO threads (id, body, file_name, is_alive, created_at) VALUES ($1, $2, $3, $4, $5)",
		"thread-1", "hello e2e", "sample123", true, fixedTime)
	require.NoError(t, err)
	t.Cleanup(func() {
		_, _ = pool.Exec(context.Background(), "DELETE FROM threads WHERE id = $1", "thread-1")
	})

	threadRepository := postgres.NewPostgresThreadRepository(pool, "thread_images/")
	logger := infralog.NewStdLogger()
	listThreadsUsecase := applicationthread.NewListThreadsUsecase(threadRepository, logger)
	threadHandler := handlerthread.NewThreadHandler(listThreadsUsecase)

	e := echo.New()
	req := httptest.NewRequest(http.MethodGet, "/threads", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	err = threadHandler.ListThreads(c)
	require.NoError(t, err)

	assert.Equal(t, http.StatusOK, rec.Code)

	var got handlerthread.ListThreadsResponse
	require.NoError(t, json.Unmarshal(rec.Body.Bytes(), &got))

	require.Len(t, got.Threads, 1)
	assert.Equal(t, "thread-1", got.Threads[0].ID)
	assert.Equal(t, "hello e2e", got.Threads[0].Body)
	assert.Equal(t, "thread_images/sample123", got.Threads[0].ImagePath)
	assert.True(t, got.Threads[0].CreatedAt.Equal(fixedTime))
}
