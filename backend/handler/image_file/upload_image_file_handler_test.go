package image_file_test

import (
	"bytes"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/labstack/echo/v4"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"

	applicationimagefile "github.com/nakanokota/bulletin-board-system/backend/application/usecase/image_file"
	"github.com/nakanokota/bulletin-board-system/backend/handler"
	"github.com/nakanokota/bulletin-board-system/backend/handler/image_file"
	"github.com/nakanokota/bulletin-board-system/backend/handler/image_file/mock_image_file"
)

const validUploadImageFileRequestBody = `{"file":{"original_file_name":"sample.png","content_type":"image/png","size_bytes":1024}}`

func newUploadImageFileContext(e *echo.Echo, body string) (echo.Context, *httptest.ResponseRecorder) {
	e.Validator = handler.NewRequestValidator()
	req := httptest.NewRequest(http.MethodPost, "/files", strings.NewReader(body))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rec := httptest.NewRecorder()
	return e.NewContext(req, rec), rec
}

func TestUploadImageFileHandler_UploadImageFile_正常系(t *testing.T) {
	tests := []struct {
		name       string
		body       string
		setupMock  func(*mock_image_file.MockuploadImageFileUsecase)
		wantStatus int
		wantBody   string
	}{
		{
			name: "リクエストが正常なとき、ステータス201かつUsecaseが返したFileID・ContentType・ContentLength・SignedURLを含むレスポンスが返ること",
			body: validUploadImageFileRequestBody,
			setupMock: func(m *mock_image_file.MockuploadImageFileUsecase) {
				m.EXPECT().Exec(gomock.Any(), gomock.Any()).Return(&applicationimagefile.UploadImageFileOutput{
					File: applicationimagefile.UploadImageFileOutputFile{
						FileID:        "file-1",
						ContentType:   "image/png",
						ContentLength: 1024,
					},
					SignedURL: "https://example.com/signed",
				}, nil)
			},
			wantStatus: http.StatusCreated,
			wantBody:   `{"file":{"file_id":"file-1","content_type":"image/png","content_length":1024},"signed_url":"https://example.com/signed"}` + "\n",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			usecase := mock_image_file.NewMockuploadImageFileUsecase(ctrl)
			tt.setupMock(usecase)
			c, rec := newUploadImageFileContext(echo.New(), tt.body)

			sut := image_file.NewUploadImageFileHandler(usecase)
			err := sut.UploadImageFile(c)

			require.NoError(t, err)
			assert.Equal(t, tt.wantStatus, rec.Code)
			assert.Equal(t, tt.wantBody, rec.Body.String())
		})
	}
}

func TestUploadImageFileHandler_UploadImageFile_Usecaseに渡す値(t *testing.T) {
	tests := []struct {
		name      string
		body      string
		setupMock func(*mock_image_file.MockuploadImageFileUsecase)
	}{
		{
			name: "リクエストが正常なとき、Usecaseに、original_file_nameから抽出した拡張子・content_type・size_bytesを持つUploadImageFileCommandが渡されること",
			body: validUploadImageFileRequestBody,
			setupMock: func(m *mock_image_file.MockuploadImageFileUsecase) {
				m.EXPECT().Exec(gomock.Any(), applicationimagefile.UploadImageFileCommand{
					Extension:   ".png",
					ContentType: "image/png",
					SizeBytes:   1024,
				}).Return(&applicationimagefile.UploadImageFileOutput{File: applicationimagefile.UploadImageFileOutputFile{FileID: "file-1"}, SignedURL: "https://example.com/signed"}, nil)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			usecase := mock_image_file.NewMockuploadImageFileUsecase(ctrl)
			tt.setupMock(usecase)
			c, _ := newUploadImageFileContext(echo.New(), tt.body)

			sut := image_file.NewUploadImageFileHandler(usecase)
			err := sut.UploadImageFile(c)

			require.NoError(t, err)
		})
	}
}

func TestUploadImageFileHandler_UploadImageFile_異常系(t *testing.T) {
	tests := []struct {
		name       string
		body       string
		setupMock  func(*mock_image_file.MockuploadImageFileUsecase)
		wantStatus int
		wantBody   string
	}{
		{
			name:       "リクエストボディのJSONの形式が不正なとき、ステータス400かつ「リクエストの形式が不正です。」が返ること",
			body:       `{"file":`,
			setupMock:  func(m *mock_image_file.MockuploadImageFileUsecase) {},
			wantStatus: http.StatusBadRequest,
			wantBody:   `{"message":"リクエストの形式が不正です。"}` + "\n",
		},
		{
			name:       "original_file_nameが空文字のとき、ステータス400かつ「ファイル名が不正です。」が返ること",
			body:       `{"file":{"original_file_name":"","content_type":"image/png","size_bytes":1024}}`,
			setupMock:  func(m *mock_image_file.MockuploadImageFileUsecase) {},
			wantStatus: http.StatusBadRequest,
			wantBody:   `{"message":"ファイル名が不正です。"}` + "\n",
		},
		{
			name:       "original_file_nameにパス区切り文字(\"/\")が含まれるとき、ステータス400かつ「ファイル名が不正です。」が返ること",
			body:       `{"file":{"original_file_name":"../sample.png","content_type":"image/png","size_bytes":1024}}`,
			setupMock:  func(m *mock_image_file.MockuploadImageFileUsecase) {},
			wantStatus: http.StatusBadRequest,
			wantBody:   `{"message":"ファイル名が不正です。"}` + "\n",
		},
		{
			name: "UsecaseがErrInvalidExtensionを返すとき、ステータス400かつ「対応していない画像形式です。」が返ること",
			body: validUploadImageFileRequestBody,
			setupMock: func(m *mock_image_file.MockuploadImageFileUsecase) {
				m.EXPECT().Exec(gomock.Any(), gomock.Any()).Return(nil, applicationimagefile.ErrInvalidExtension)
			},
			wantStatus: http.StatusBadRequest,
			wantBody:   `{"message":"対応していない画像形式です。"}` + "\n",
		},
		{
			name: "UsecaseがErrContentTypeRequiredを返すとき、ステータス400かつ「ファイルの形式を入力してください。」が返ること",
			body: validUploadImageFileRequestBody,
			setupMock: func(m *mock_image_file.MockuploadImageFileUsecase) {
				m.EXPECT().Exec(gomock.Any(), gomock.Any()).Return(nil, applicationimagefile.ErrContentTypeRequired)
			},
			wantStatus: http.StatusBadRequest,
			wantBody:   `{"message":"ファイルの形式を入力してください。"}` + "\n",
		},
		{
			name: "UsecaseがErrSizeBytesInvalidを返すとき、ステータス400かつ「ファイルサイズが不正です。」が返ること",
			body: validUploadImageFileRequestBody,
			setupMock: func(m *mock_image_file.MockuploadImageFileUsecase) {
				m.EXPECT().Exec(gomock.Any(), gomock.Any()).Return(nil, applicationimagefile.ErrSizeBytesInvalid)
			},
			wantStatus: http.StatusBadRequest,
			wantBody:   `{"message":"ファイルサイズが不正です。"}` + "\n",
		},
		{
			name: "Usecaseが上記以外のエラーを返すとき、ステータス500かつ「ファイルのアップロードに失敗しました。」が返り、エラー詳細はレスポンスに含まれないこと",
			body: validUploadImageFileRequestBody,
			setupMock: func(m *mock_image_file.MockuploadImageFileUsecase) {
				m.EXPECT().Exec(gomock.Any(), gomock.Any()).Return(nil, errors.New("signer failed"))
			},
			wantStatus: http.StatusInternalServerError,
			wantBody:   `{"message":"ファイルのアップロードに失敗しました。"}` + "\n",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			usecase := mock_image_file.NewMockuploadImageFileUsecase(ctrl)
			tt.setupMock(usecase)
			c, rec := newUploadImageFileContext(echo.New(), tt.body)

			sut := image_file.NewUploadImageFileHandler(usecase)
			err := sut.UploadImageFile(c)

			require.NoError(t, err)
			assert.Equal(t, tt.wantStatus, rec.Code)
			assert.Equal(t, tt.wantBody, rec.Body.String())
		})
	}
}

func TestUploadImageFileHandler_UploadImageFile_ログ出力(t *testing.T) {
	tests := []struct {
		name          string
		body          string
		setupMock     func(*mock_image_file.MockuploadImageFileUsecase)
		wantLogSubstr string
	}{
		{
			name: "Usecaseが上記以外のエラーを返すとき、エラー詳細がログに出力されること",
			body: validUploadImageFileRequestBody,
			setupMock: func(m *mock_image_file.MockuploadImageFileUsecase) {
				m.EXPECT().Exec(gomock.Any(), gomock.Any()).Return(nil, errors.New("signer failed"))
			},
			wantLogSubstr: "signer failed",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			usecase := mock_image_file.NewMockuploadImageFileUsecase(ctrl)
			tt.setupMock(usecase)

			var logBuf bytes.Buffer
			e := echo.New()
			e.Logger.SetOutput(&logBuf)
			c, _ := newUploadImageFileContext(e, tt.body)

			sut := image_file.NewUploadImageFileHandler(usecase)
			err := sut.UploadImageFile(c)

			require.NoError(t, err)
			assert.Contains(t, logBuf.String(), tt.wantLogSubstr)
		})
	}
}
