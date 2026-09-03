package thread

import (
	"context"
	"net/http"
	"time"

	"github.com/labstack/echo/v4"

	applicationthread "github.com/nakanokota/bulletin-board-system/backend/application/thread"
)

type ThreadResponse struct {
	ID        string    `json:"id"`
	Body      string    `json:"body"`
	ImagePath *string   `json:"image_path"`
	CreatedAt time.Time `json:"created_at"`
}

type ListThreadsResponse struct {
	Threads []ThreadResponse `json:"threads"`
}

// listThreadsUsecase is satisfied by *applicationthread.ListThreadsUsecase; the seam exists for mocking in handler tests.
type listThreadsUsecase interface {
	Exec(ctx context.Context, cmd applicationthread.ListThreadsCommand) (*applicationthread.ListThreadsOutput, error)
}

type ThreadHandler struct {
	listThreadsUsecase listThreadsUsecase
}

func NewThreadHandler(listThreadsUsecase listThreadsUsecase) *ThreadHandler {
	return &ThreadHandler{listThreadsUsecase: listThreadsUsecase}
}

func (h *ThreadHandler) ListThreads(c echo.Context) error {
	output, err := h.listThreadsUsecase.Exec(c.Request().Context(), applicationthread.ListThreadsCommand{})
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"message": "スレッド一覧の取得に失敗しました。"})
	}

	return c.JSON(http.StatusOK, ListThreadsResponse{Threads: toThreadResponses(output.Threads)})
}

func toThreadResponses(threads []applicationthread.ListThreadsOutputThread) []ThreadResponse {
	responses := make([]ThreadResponse, 0, len(threads))
	for _, t := range threads {
		responses = append(responses, ThreadResponse{ID: t.ID, Body: t.Body, ImagePath: t.ImagePath, CreatedAt: t.CreatedAt})
	}
	return responses
}
