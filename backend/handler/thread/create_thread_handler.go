package thread

import (
	"context"
	"errors"
	"net/http"

	"github.com/labstack/echo/v4"

	applicationthread "github.com/nakanokota/bulletin-board-system/backend/application/usecase/thread"
	domainthread "github.com/nakanokota/bulletin-board-system/backend/domain/thread"
)

type CreateThreadRequestBody struct {
	Body     string `json:"body"`
	FileName string `json:"file_name"`
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
		IP:       c.RealIP(),
		Body:     req.Thread.Body,
		FileName: req.Thread.FileName,
	})
	if err != nil {
		if errors.Is(err, domainthread.ErrPendingThreadBodyEmpty) {
			return c.JSON(http.StatusBadRequest, map[string]string{"message": "本文を入力してください。"})
		}
		if errors.Is(err, domainthread.ErrPendingThreadFileNameMissing) || errors.Is(err, domainthread.ErrFileNameInvalid) {
			return c.JSON(http.StatusBadRequest, map[string]string{"message": "画像ファイルを添付してください。"})
		}
		c.Logger().Error(err)
		return c.JSON(http.StatusInternalServerError, map[string]string{"message": "スレッドの作成に失敗しました。"})
	}

	return c.JSON(http.StatusCreated, CreateThreadResponse{Thread: CreateThreadResponseBody{ID: output.ThreadID}})
}
