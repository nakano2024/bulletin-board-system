package main

import (
	"context"
	"net/http"
	"os"

	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"

	applicationthread "github.com/nakanokota/bulletin-board-system/backend/application/thread"
	"github.com/nakanokota/bulletin-board-system/backend/handler/thread"
	"github.com/nakanokota/bulletin-board-system/backend/infra/postgres"
)

func main() {
	ctx := context.Background()

	pool, err := postgres.NewPool(ctx, os.Getenv("DATABASE_URL"))
	if err != nil {
		panic(err)
	}

	threadQueryService := postgres.NewPostgresThreadQueryService(pool)
	listThreadsUsecase := applicationthread.NewListThreadsUsecase(threadQueryService)
	threadHandler := thread.NewThreadHandler(listThreadsUsecase)

	e := echo.New()
	e.Use(middleware.Logger())
	e.Use(middleware.Recover())

	e.GET("/health", func(c echo.Context) error {
		return c.JSON(http.StatusOK, map[string]string{"status": "ok"})
	})
	e.GET("/threads", threadHandler.ListThreads)

	e.Logger.Fatal(e.Start(":8080"))
}
