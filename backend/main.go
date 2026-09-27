package main

import (
	"context"
	"net/http"
	"os"

	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"

	applicationthread "github.com/nakanokota/bulletin-board-system/backend/application/usecase/thread"
	domainthread "github.com/nakanokota/bulletin-board-system/backend/domain/thread"
	domainuser "github.com/nakanokota/bulletin-board-system/backend/domain/user"
	"github.com/nakanokota/bulletin-board-system/backend/handler"
	"github.com/nakanokota/bulletin-board-system/backend/handler/thread"
	"github.com/nakanokota/bulletin-board-system/backend/infra/clock"
	infralog "github.com/nakanokota/bulletin-board-system/backend/infra/log"
	"github.com/nakanokota/bulletin-board-system/backend/infra/postgres"
)

func main() {
	ctx := context.Background()

	pool, err := postgres.NewPool(ctx, os.Getenv("DATABASE_URL"))
	if err != nil {
		panic(err)
	}

	threadRepository := postgres.NewPostgresThreadRepository(pool, os.Getenv("THREAD_IMAGE_BASE_PATH"))
	threadFileChecker := postgres.NewPostgresThreadFileChecker(pool)
	userRepository := postgres.NewPostgresUserRepository(pool)
	threadCreationService := domainthread.NewThreadCreationService(threadFileChecker, threadRepository)
	userService := domainuser.NewUserService(userRepository)
	logger := infralog.NewStdLogger()
	timeGetter := clock.NewStdTimeGetter()

	listThreadsUsecase := applicationthread.NewListThreadsUsecase(threadRepository, logger)
	createThreadUsecase := applicationthread.NewCreateThreadUsecase(threadCreationService, userService, timeGetter, logger)
	listThreadsHandler := thread.NewListThreadsHandler(listThreadsUsecase)
	createThreadHandler := thread.NewCreateThreadHandler(createThreadUsecase)

	e := echo.New()
	e.Use(middleware.Logger())
	e.Use(middleware.Recover())
	e.Validator = handler.NewRequestValidator()

	e.GET("/health", func(c echo.Context) error {
		return c.JSON(http.StatusOK, map[string]string{"status": "ok"})
	})
	e.GET("/threads", listThreadsHandler.ListThreads)
	e.POST("/threads", createThreadHandler.CreateThread)

	e.Logger.Fatal(e.Start(":8080"))
}
