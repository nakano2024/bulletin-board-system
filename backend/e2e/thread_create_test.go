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
	domainuser "github.com/nakanokota/bulletin-board-system/backend/domain/user"
	"github.com/nakanokota/bulletin-board-system/backend/handler"
	handlerthread "github.com/nakanokota/bulletin-board-system/backend/handler/thread"
	"github.com/nakanokota/bulletin-board-system/backend/infra/clock"
	infralog "github.com/nakanokota/bulletin-board-system/backend/infra/log"
	"github.com/nakanokota/bulletin-board-system/backend/infra/postgres"
)

// TestE2E_CreateThread_正常系 wires the real Handler -> Usecase -> UserService/Repositories -> DB stack
// (no mocks) and checks POST /threads persists both the new user and the new thread. Kept to a single
// case on purpose: full-stack DB tests are the most brittle in the suite.
func TestE2E_CreateThread_正常系(t *testing.T) {
	ctx := context.Background()
	ip := "203.0.113.50"

	pool, err := pgxpool.New(ctx, os.Getenv("E2E_DATABASE_URL"))
	require.NoError(t, err)
	t.Cleanup(pool.Close)
	t.Cleanup(func() {
		_, _ = pool.Exec(context.Background(), "DELETE FROM threads WHERE user_id IN (SELECT id FROM users WHERE ip = $1)", ip)
		_, _ = pool.Exec(context.Background(), "DELETE FROM users WHERE ip = $1", ip)
	})

	threadRepository := postgres.NewPostgresThreadRepository(pool, "thread_images/")
	userRepository := postgres.NewPostgresUserRepository(pool)
	userService := domainuser.NewUserService(userRepository)
	logger := infralog.NewStdLogger()
	timeGetter := clock.NewStdTimeGetter()
	createThreadUsecase := applicationthread.NewCreateThreadUsecase(threadRepository, userService, timeGetter, logger)
	createThreadHandler := handlerthread.NewCreateThreadHandler(createThreadUsecase)

	e := echo.New()
	e.Validator = handler.NewRequestValidator()
	req := httptest.NewRequest(http.MethodPost, "/threads", strings.NewReader(`{"thread":{"body":"hello e2e","file_name":"sample.png"}}`))
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

	var storedBody, storedFileName, storedUserID string
	err = pool.QueryRow(ctx, "SELECT body, file_name, user_id FROM threads WHERE id = $1", got.Thread.ID).Scan(&storedBody, &storedFileName, &storedUserID)
	require.NoError(t, err)
	assert.Equal(t, "hello e2e", storedBody)
	assert.Equal(t, "sample.png", storedFileName)

	var storedIP string
	err = pool.QueryRow(ctx, "SELECT ip FROM users WHERE id = $1", storedUserID).Scan(&storedIP)
	require.NoError(t, err)
	assert.Equal(t, ip, storedIP)
}
