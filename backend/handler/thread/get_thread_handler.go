package thread

import (
	"context"
	"errors"
	"net/http"
	"time"

	"github.com/labstack/echo/v4"

	applicationthread "github.com/nakanokota/bulletin-board-system/backend/application/usecase/thread"
)

type GetThreadResponseReply struct {
	ID     string `json:"id"`
	UserID string `json:"user_id"`
	Body   string `json:"body"`
	// ImagePath is empty when the reply has no attached image.
	ImagePath string    `json:"image_path"`
	CreatedAt time.Time `json:"created_at"`
}

type GetThreadResponseThread struct {
	ID        string                   `json:"id"`
	UserID    string                   `json:"user_id"`
	Body      string                   `json:"body"`
	ImagePath string                   `json:"image_path"`
	CreatedAt time.Time                `json:"created_at"`
	Replies   []GetThreadResponseReply `json:"replies"`
}

type GetThreadResponse struct {
	Thread GetThreadResponseThread `json:"thread"`
}

// getThreadUsecase is satisfied by *applicationthread.GetThreadUsecase; the seam exists for mocking in handler tests.
type getThreadUsecase interface {
	Exec(ctx context.Context, cmd applicationthread.GetThreadCommand) (*applicationthread.GetThreadOutput, error)
}

type GetThreadHandler struct {
	getThreadUsecase getThreadUsecase
}

func NewGetThreadHandler(getThreadUsecase getThreadUsecase) *GetThreadHandler {
	return &GetThreadHandler{getThreadUsecase: getThreadUsecase}
}

// GetThread handles GET /threads/:id.
func (h *GetThreadHandler) GetThread(c echo.Context) error {
	output, err := h.getThreadUsecase.Exec(c.Request().Context(), applicationthread.GetThreadCommand{ThreadID: c.Param("id")})
	if err != nil {
		return h.toErrorResponse(c, err)
	}

	return c.JSON(http.StatusOK, GetThreadResponse{Thread: toGetThreadResponseThread(output.Thread)})
}

// toErrorResponse maps the usecase's errors to a status code and a fixed user-facing message;
// the error detail goes to the log only. A missing thread and a fallen-off (dat落ち) one look the same to the user.
func (h *GetThreadHandler) toErrorResponse(c echo.Context, err error) error {
	switch {
	case errors.Is(err, applicationthread.ErrThreadNotFound), errors.Is(err, applicationthread.ErrThreadDead):
		return c.JSON(http.StatusNotFound, map[string]string{"message": "スレッドが存在しません。"})
	default:
		c.Logger().Error(err)
		return c.JSON(http.StatusInternalServerError, map[string]string{"message": "スレッドを開くのに失敗しました。"})
	}
}

func toGetThreadResponseThread(t applicationthread.GetThreadOutputThread) GetThreadResponseThread {
	replies := make([]GetThreadResponseReply, 0, len(t.Replies))
	for _, r := range t.Replies {
		replies = append(replies, GetThreadResponseReply{ID: r.ID, UserID: r.UserID, Body: r.Body, ImagePath: r.ImagePath, CreatedAt: r.CreatedAt})
	}

	return GetThreadResponseThread{
		ID:        t.ID,
		UserID:    t.UserID,
		Body:      t.Body,
		ImagePath: t.ImagePath,
		CreatedAt: t.CreatedAt,
		Replies:   replies,
	}
}
