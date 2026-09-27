package thread

import (
	"context"
	"errors"
	"net/http"

	"github.com/labstack/echo/v4"

	applicationthread "github.com/nakanokota/bulletin-board-system/backend/application/usecase/thread"
)

type CreateThreadRequestBody struct {
	Body   string `json:"body"`
	FileID string `json:"file_id"`
}

type CreateThreadRequest struct {
	Thread CreateThreadRequestBody `json:"thread"`
}

type CreateThreadResponseBody struct {
	ID string `json:"id"`
}

type CreateThreadResponse struct {
	Thread CreateThreadResponseBody `json:"thread"`
}

// createThreadUsecase is satisfied by *applicationthread.CreateThreadUsecase; the seam exists for mocking in handler tests.
type createThreadUsecase interface {
	Exec(ctx context.Context, cmd applicationthread.CreateThreadCommand) (*applicationthread.CreateThreadOutput, error)
}

type CreateThreadHandler struct {
	createThreadUsecase createThreadUsecase
}

func NewCreateThreadHandler(createThreadUsecase createThreadUsecase) *CreateThreadHandler {
	return &CreateThreadHandler{createThreadUsecase: createThreadUsecase}
}

func (h *CreateThreadHandler) CreateThread(c echo.Context) error {
	var req CreateThreadRequest
	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"message": "リクエストの形式が不正です。"})
	}

	output, err := h.createThreadUsecase.Exec(c.Request().Context(), applicationthread.CreateThreadCommand{
		IP:     c.RealIP(),
		Body:   req.Thread.Body,
		FileID: req.Thread.FileID,
	})
	if err != nil {
		return h.toErrorResponse(c, err)
	}

	return c.JSON(http.StatusCreated, CreateThreadResponse{Thread: CreateThreadResponseBody{ID: output.ThreadID}})
}

// toErrorResponse maps the usecase's errors to a status code and a fixed user-facing message;
// the error detail goes to the log only.
func (h *CreateThreadHandler) toErrorResponse(c echo.Context, err error) error {
	switch {
	case errors.Is(err, applicationthread.ErrInvalidBody):
		return c.JSON(http.StatusBadRequest, map[string]string{"message": "本文を入力してください。"})
	case errors.Is(err, applicationthread.ErrInvalidFile):
		return c.JSON(http.StatusBadRequest, map[string]string{"message": "画像ファイルを添付してください。"})
	case errors.Is(err, applicationthread.ErrFileAlreadyUsed):
		return c.JSON(http.StatusBadRequest, map[string]string{"message": "この画像ファイルは既に使用されています。"})
	default:
		c.Logger().Error(err)
		return c.JSON(http.StatusInternalServerError, map[string]string{"message": "スレッドの作成に失敗しました。"})
	}
}
