package image_file

import (
	"context"
	"errors"
	"net/http"
	"path/filepath"

	"github.com/labstack/echo/v4"

	applicationimagefile "github.com/nakanokota/bulletin-board-system/backend/application/usecase/image_file"
)

type UploadImageFileRequestBody struct {
	OriginalFileName string `json:"original_file_name" validate:"required,excludesall=/\\"`
	ContentType      string `json:"content_type"`
	SizeBytes        int64  `json:"size_bytes"`
}

type UploadImageFileRequest struct {
	File UploadImageFileRequestBody `json:"file"`
}

// ContentType and ContentLength are the headers the client's PUT to SignedURL must send exactly,
// since they are bound into the presigned URL's signature.
type UploadImageFileResponseBody struct {
	FileID        string `json:"file_id"`
	ContentType   string `json:"content_type"`
	ContentLength int64  `json:"content_length"`
}

type UploadImageFileResponse struct {
	File      UploadImageFileResponseBody `json:"file"`
	SignedURL string                      `json:"signed_url"`
}

// uploadImageFileUsecase is satisfied by *applicationimagefile.UploadImageFileUsecase; the seam exists for mocking in handler tests.
type uploadImageFileUsecase interface {
	Exec(ctx context.Context, cmd applicationimagefile.UploadImageFileCommand) (*applicationimagefile.UploadImageFileOutput, error)
}

type UploadImageFileHandler struct {
	uploadImageFileUsecase uploadImageFileUsecase
}

func NewUploadImageFileHandler(uploadImageFileUsecase uploadImageFileUsecase) *UploadImageFileHandler {
	return &UploadImageFileHandler{uploadImageFileUsecase: uploadImageFileUsecase}
}

func (h *UploadImageFileHandler) UploadImageFile(c echo.Context) error {
	var req UploadImageFileRequest
	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"message": "リクエストの形式が不正です。"})
	}
	if err := c.Validate(&req); err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"message": "ファイル名が不正です。"})
	}

	output, err := h.uploadImageFileUsecase.Exec(c.Request().Context(), applicationimagefile.UploadImageFileCommand{
		Extension:   filepath.Ext(req.File.OriginalFileName),
		ContentType: req.File.ContentType,
		SizeBytes:   req.File.SizeBytes,
	})
	if err != nil {
		return h.toErrorResponse(c, err)
	}

	return c.JSON(http.StatusCreated, UploadImageFileResponse{
		File: UploadImageFileResponseBody{
			FileID:        output.File.FileID,
			ContentType:   output.File.ContentType,
			ContentLength: output.File.ContentLength,
		},
		SignedURL: output.SignedURL,
	})
}

// toErrorResponse maps the usecase's errors to a status code and a fixed user-facing message;
// the error detail goes to the log only.
func (h *UploadImageFileHandler) toErrorResponse(c echo.Context, err error) error {
	switch {
	case errors.Is(err, applicationimagefile.ErrInvalidExtension):
		return c.JSON(http.StatusBadRequest, map[string]string{"message": "対応していない画像形式です。"})
	case errors.Is(err, applicationimagefile.ErrContentTypeRequired):
		return c.JSON(http.StatusBadRequest, map[string]string{"message": "ファイルの形式を入力してください。"})
	case errors.Is(err, applicationimagefile.ErrSizeBytesInvalid):
		return c.JSON(http.StatusBadRequest, map[string]string{"message": "ファイルサイズが不正です。"})
	default:
		c.Logger().Error(err)
		return c.JSON(http.StatusInternalServerError, map[string]string{"message": "ファイルのアップロードに失敗しました。"})
	}
}
