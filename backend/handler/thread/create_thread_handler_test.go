package thread_test

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

	applicationthread "github.com/nakanokota/bulletin-board-system/backend/application/usecase/thread"
	domainthread "github.com/nakanokota/bulletin-board-system/backend/domain/thread"
	"github.com/nakanokota/bulletin-board-system/backend/handler/thread"
	"github.com/nakanokota/bulletin-board-system/backend/handler/thread/mock_thread"
)

func TestCreateThreadHandler_CreateThread_正常系(t *testing.T) {
	tests := []struct {
		name       string
		setupMock  func(*mock_thread.MockcreateThreadUsecase)
		wantStatus int
		wantBody   string
	}{
		{
			name: "リクエストが正常なとき、ステータス201かつ作成されたThreadのIDを含むレスポンスが返ること",
			setupMock: func(m *mock_thread.MockcreateThreadUsecase) {
				m.EXPECT().Exec(gomock.Any(), applicationthread.CreateThreadCommand{
					IP:       "192.0.2.1",
					Body:     "hello",
					FileName: "sample.png",
				}).Return(&applicationthread.CreateThreadOutput{ThreadID: "thread-1"}, nil)
			},
			wantStatus: http.StatusCreated,
			wantBody:   `{"thread":{"id":"thread-1"}}` + "\n",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			usecase := mock_thread.NewMockcreateThreadUsecase(ctrl)
			tt.setupMock(usecase)

			e := echo.New()
			req := httptest.NewRequest(http.MethodPost, "/threads", strings.NewReader(`{"thread":{"body":"hello","file_name":"sample.png"}}`))
			req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
			req.RemoteAddr = "192.0.2.1:12345"
			rec := httptest.NewRecorder()
			c := e.NewContext(req, rec)

			sut := thread.NewCreateThreadHandler(usecase)
			err := sut.CreateThread(c)

			require.NoError(t, err)
			assert.Equal(t, tt.wantStatus, rec.Code)
			assert.Equal(t, tt.wantBody, rec.Body.String())
		})
	}
}

func TestCreateThreadHandler_CreateThread_異常系(t *testing.T) {
	tests := []struct {
		name       string
		body       string
		setupMock  func(*mock_thread.MockcreateThreadUsecase)
		wantStatus int
		wantBody   string
	}{
		{
			name:       "リクエストボディの形式が不正なとき、ステータス400が返ること",
			body:       `{invalid`,
			setupMock:  func(m *mock_thread.MockcreateThreadUsecase) {},
			wantStatus: http.StatusBadRequest,
			wantBody:   `{"message":"リクエストの形式が不正です。"}` + "\n",
		},
		{
			name: "本文が空のとき、ステータス400かつ「本文を入力してください。」が返ること",
			body: `{"thread":{"body":"","file_name":"sample.png"}}`,
			setupMock: func(m *mock_thread.MockcreateThreadUsecase) {
				m.EXPECT().Exec(gomock.Any(), gomock.Any()).Return(nil, domainthread.ErrPendingThreadBodyEmpty)
			},
			wantStatus: http.StatusBadRequest,
			wantBody:   `{"message":"本文を入力してください。"}` + "\n",
		},
		{
			name: "画像ファイル名が未指定のとき、ステータス400かつ「画像ファイルを添付してください。」が返ること",
			body: `{"thread":{"body":"hello","file_name":""}}`,
			setupMock: func(m *mock_thread.MockcreateThreadUsecase) {
				m.EXPECT().Exec(gomock.Any(), gomock.Any()).Return(nil, domainthread.ErrPendingThreadFileNameMissing)
			},
			wantStatus: http.StatusBadRequest,
			wantBody:   `{"message":"画像ファイルを添付してください。"}` + "\n",
		},
		{
			name: "画像ファイル名の形式が不正なとき、ステータス400かつ「画像ファイルを添付してください。」が返ること",
			body: `{"thread":{"body":"hello","file_name":"sample.exe"}}`,
			setupMock: func(m *mock_thread.MockcreateThreadUsecase) {
				m.EXPECT().Exec(gomock.Any(), gomock.Any()).Return(nil, domainthread.ErrFileNameInvalid)
			},
			wantStatus: http.StatusBadRequest,
			wantBody:   `{"message":"画像ファイルを添付してください。"}` + "\n",
		},
		{
			name: "その他のエラーが返るとき、ステータス500かつ「スレッドの作成に失敗しました。」が返ること",
			body: `{"thread":{"body":"hello","file_name":"sample.png"}}`,
			setupMock: func(m *mock_thread.MockcreateThreadUsecase) {
				m.EXPECT().Exec(gomock.Any(), gomock.Any()).Return(nil, errors.New("repository failed"))
			},
			wantStatus: http.StatusInternalServerError,
			wantBody:   `{"message":"スレッドの作成に失敗しました。"}` + "\n",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			usecase := mock_thread.NewMockcreateThreadUsecase(ctrl)
			tt.setupMock(usecase)

			e := echo.New()
			req := httptest.NewRequest(http.MethodPost, "/threads", strings.NewReader(tt.body))
			req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
			rec := httptest.NewRecorder()
			c := e.NewContext(req, rec)

			sut := thread.NewCreateThreadHandler(usecase)
			err := sut.CreateThread(c)

			require.NoError(t, err)
			assert.Equal(t, tt.wantStatus, rec.Code)
			assert.Equal(t, tt.wantBody, rec.Body.String())
		})
	}
}

func TestCreateThreadHandler_CreateThread_ログ出力(t *testing.T) {
	tests := []struct {
		name          string
		setupMock     func(*mock_thread.MockcreateThreadUsecase)
		wantLogSubstr string
	}{
		{
			name: "Usecaseがドメインバリデーション以外のエラーを返すとき、エラー詳細がログに出力されること",
			setupMock: func(m *mock_thread.MockcreateThreadUsecase) {
				m.EXPECT().Exec(gomock.Any(), gomock.Any()).Return(nil, errors.New("repository failed"))
			},
			wantLogSubstr: "repository failed",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			usecase := mock_thread.NewMockcreateThreadUsecase(ctrl)
			tt.setupMock(usecase)

			var logBuf bytes.Buffer
			e := echo.New()
			e.Logger.SetOutput(&logBuf)
			req := httptest.NewRequest(http.MethodPost, "/threads", strings.NewReader(`{"thread":{"body":"hello","file_name":"sample.png"}}`))
			req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
			rec := httptest.NewRecorder()
			c := e.NewContext(req, rec)

			sut := thread.NewCreateThreadHandler(usecase)
			err := sut.CreateThread(c)

			require.NoError(t, err)
			assert.Contains(t, logBuf.String(), tt.wantLogSubstr)
		})
	}
}
