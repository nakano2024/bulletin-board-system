package e2e

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/labstack/echo/v4"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	applicationthread "github.com/nakanokota/bulletin-board-system/backend/application/usecase/thread"
	domainthread "github.com/nakanokota/bulletin-board-system/backend/domain/thread"
	domainuser "github.com/nakanokota/bulletin-board-system/backend/domain/user"
	handlerthread "github.com/nakanokota/bulletin-board-system/backend/handler/thread"
	"github.com/nakanokota/bulletin-board-system/backend/infra/clock"
	infralog "github.com/nakanokota/bulletin-board-system/backend/infra/log"
	"github.com/nakanokota/bulletin-board-system/backend/infra/postgres"
)

// TestE2E_CreateThread_正常系 wires the real Handler -> Usecase -> UserService/Repositories -> DB stack
// (no mocks) and checks POST /threads, given a registered unused file, persists both the new user and the new thread. Kept to a single
// case on purpose: full-stack DB tests are the most brittle in the suite.
func TestE2E_CreateThread_正常系(t *testing.T) {
	ctx := context.Background()
	ip := "203.0.113.50"
	fileID := "file-e2e-create"

	pool, err := pgxpool.New(ctx, os.Getenv("E2E_DATABASE_URL"))
	require.NoError(t, err)
	t.Cleanup(pool.Close)

	_, err = pool.Exec(ctx, "INSERT INTO files (id, name, content_type, size_bytes) VALUES ($1, $2, $3, $4)", fileID, "sample.png", "image/png", 1024)
	require.NoError(t, err)
	t.Cleanup(func() {
		_, _ = pool.Exec(context.Background(), "DELETE FROM threads WHERE file_id = $1", fileID)
		_, _ = pool.Exec(context.Background(), "DELETE FROM files WHERE id = $1", fileID)
		_, _ = pool.Exec(context.Background(), "DELETE FROM users WHERE ip = $1", ip)
	})

	createThreadHandler := newE2ECreateThreadHandler(pool)

	e := echo.New()
	req := httptest.NewRequest(http.MethodPost, "/threads", strings.NewReader(`{"thread":{"body":"hello e2e","file_id":"`+fileID+`"}}`))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	req.RemoteAddr = ip + ":12345"
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	err = createThreadHandler.CreateThread(c)
	require.NoError(t, err)

	assert.Equal(t, http.StatusCreated, rec.Code)

	var got handlerthread.CreateThreadResponse
	require.NoError(t, json.Unmarshal(rec.Body.Bytes(), &got))
	require.NotEmpty(t, got.Thread.ID)

	var storedBody, storedFileID, storedUserID string
	err = pool.QueryRow(ctx, "SELECT body, file_id, user_id FROM threads WHERE id = $1", got.Thread.ID).Scan(&storedBody, &storedFileID, &storedUserID)
	require.NoError(t, err)
	assert.Equal(t, "hello e2e", storedBody)
	assert.Equal(t, fileID, storedFileID)

	var storedIP string
	err = pool.QueryRow(ctx, "SELECT ip FROM users WHERE id = $1", storedUserID).Scan(&storedIP)
	require.NoError(t, err)
	assert.Equal(t, ip, storedIP)
}

// TestE2E_CreateThread_異常系_ロールバック checks that the whole creation runs in one transaction: the poster's user row
// is created before the file check, so a request with an unknown file_id must not leave that row behind.
func TestE2E_CreateThread_異常系_ロールバック(t *testing.T) {
	ctx := context.Background()
	ip := "203.0.113.51"

	pool, err := pgxpool.New(ctx, os.Getenv("E2E_DATABASE_URL"))
	require.NoError(t, err)
	t.Cleanup(pool.Close)
	t.Cleanup(func() {
		_, _ = pool.Exec(context.Background(), "DELETE FROM users WHERE ip = $1", ip)
	})

	createThreadHandler := newE2ECreateThreadHandler(pool)

	e := echo.New()
	req := httptest.NewRequest(http.MethodPost, "/threads", strings.NewReader(`{"thread":{"body":"hello e2e","file_id":"file-e2e-missing"}}`))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	req.RemoteAddr = ip + ":12345"
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	err = createThreadHandler.CreateThread(c)
	require.NoError(t, err)

	assert.Equal(t, http.StatusBadRequest, rec.Code)

	var userCount int
	err = pool.QueryRow(ctx, "SELECT count(*) FROM users WHERE ip = $1", ip).Scan(&userCount)
	require.NoError(t, err)
	assert.Equal(t, 0, userCount)
}

// newE2ECreateThreadHandler assembles POST /threads the same way main.go does, with no mocks.
func newE2ECreateThreadHandler(pool *pgxpool.Pool) *handlerthread.CreateThreadHandler {
	threadRepository := postgres.NewPostgresThreadRepository(pool, "thread_images/")
	threadFileChecker := postgres.NewPostgresThreadFileChecker(pool)
	userRepository := postgres.NewPostgresUserRepository(pool)
	txManager := postgres.NewPostgresTransactionManager(pool)
	threadCreationService := domainthread.NewThreadCreationService(threadFileChecker, threadRepository)
	userService := domainuser.NewUserService(userRepository)
	logger := infralog.NewStdLogger()
	timeGetter := clock.NewStdTimeGetter()
	createThreadUsecase := applicationthread.NewCreateThreadUsecase(threadCreationService, userService, timeGetter, txManager, logger)
	return handlerthread.NewCreateThreadHandler(createThreadUsecase)
}
