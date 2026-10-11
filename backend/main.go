package main

import (
	"context"
	"net/http"
	"os"

	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"

	applicationimagefile "github.com/nakanokota/bulletin-board-system/backend/application/usecase/image_file"
	applicationthread "github.com/nakanokota/bulletin-board-system/backend/application/usecase/thread"
	domainthread "github.com/nakanokota/bulletin-board-system/backend/domain/thread"
	domainuser "github.com/nakanokota/bulletin-board-system/backend/domain/user"
	"github.com/nakanokota/bulletin-board-system/backend/handler"
	"github.com/nakanokota/bulletin-board-system/backend/handler/image_file"
	"github.com/nakanokota/bulletin-board-system/backend/handler/thread"
	"github.com/nakanokota/bulletin-board-system/backend/infra/clock"
	infralog "github.com/nakanokota/bulletin-board-system/backend/infra/log"
	"github.com/nakanokota/bulletin-board-system/backend/infra/objectstorage"
	"github.com/nakanokota/bulletin-board-system/backend/infra/postgres"
	infrauuid "github.com/nakanokota/bulletin-board-system/backend/infra/uuid"
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
	imageFileRepository := postgres.NewPostgresImageFileRepository(pool)
	txManager := postgres.NewPostgresTransactionManager(pool)
	threadCreationService := domainthread.NewThreadCreationService(threadFileChecker, threadRepository)
	userService := domainuser.NewUserService(userRepository)
	logger := infralog.NewStdLogger()
	timeGetter := clock.NewStdTimeGetter()
	uuidGenerator := infrauuid.NewGoogleUuidGenerator()
	signedURLIssuer := newSignedURLIssuer()

	listThreadsUsecase := applicationthread.NewListThreadsUsecase(threadRepository, logger)
	createThreadUsecase := applicationthread.NewCreateThreadUsecase(threadCreationService, userService, timeGetter, txManager, logger)
	uploadImageFileUsecase := applicationimagefile.NewUploadImageFileUsecase(imageFileRepository, uuidGenerator, signedURLIssuer, logger)
	listThreadsHandler := thread.NewListThreadsHandler(listThreadsUsecase)
	createThreadHandler := thread.NewCreateThreadHandler(createThreadUsecase)
	uploadImageFileHandler := image_file.NewUploadImageFileHandler(uploadImageFileUsecase)

	e := echo.New()
	e.Use(middleware.Logger())
	e.Use(middleware.Recover())
	e.Validator = handler.NewRequestValidator()

	e.GET("/health", func(c echo.Context) error {
		return c.JSON(http.StatusOK, map[string]string{"status": "ok"})
	})
	e.GET("/threads", listThreadsHandler.ListThreads)
	e.POST("/threads", createThreadHandler.CreateThread)
	e.POST("/files", uploadImageFileHandler.UploadImageFile)

	e.Logger.Fatal(e.Start(":8080"))
}

// newSignedURLIssuer picks the ISignedURLIssuer implementation for the running environment:
// APP_ENV=local uses RustFS (the local S3-compatible object storage), anything else uses the
// not-yet-implemented deploy environment's implementation.
func newSignedURLIssuer() applicationimagefile.ISignedURLIssuer {
	if os.Getenv("APP_ENV") == "local" {
		return objectstorage.NewRustFSSignedURLIssuer(
			os.Getenv("RUSTFS_ENDPOINT"),
			os.Getenv("RUSTFS_ACCESS_KEY"),
			os.Getenv("RUSTFS_SECRET_KEY"),
			os.Getenv("RUSTFS_BUCKET"),
		)
	}
	return objectstorage.NewDeploySignedURLIssuer()
}
