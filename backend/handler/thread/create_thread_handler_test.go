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
	"github.com/nakanokota/bulletin-board-system/backend/handler"
	"github.com/nakanokota/bulletin-board-system/backend/handler/thread"
	"github.com/nakanokota/bulletin-board-system/backend/handler/thread/mock_thread"
)

const validCreateThreadRequestBody = `{"thread":{"body":"hello","file_id":"file-1"}}`

func newCreateThreadContext(e *echo.Echo, body string) (echo.Context, *httptest.ResponseRecorder) {
	req := httptest.NewRequest(http.MethodPost, "/threads", strings.NewReader(body))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	req.RemoteAddr = "192.0.2.1:12345"
	rec := httptest.NewRecorder()
	return e.NewContext(req, rec), rec
}

func TestCreateThreadHandler_CreateThread_正常系(t *testing.T) {
	tests := []struct {
		name       string
		body       string
		setupMock  func(*mock_thread.MockcreateThreadUsecase)
		wantStatus int
		wantBody   string
	}{
		{
			name: "リクエストが正常なとき、ステータス201かつ作成されたThreadのIDを含むレスポンスが返ること",
			body: validCreateThreadRequestBody,
			setupMock: func(m *mock_thread.MockcreateThreadUsecase) {
				m.EXPECT().Exec(gomock.Any(), gomock.Any()).Return(&applicationthread.CreateThreadOutput{ThreadID: "thread-1"}, nil)
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
			c, rec := newCreateThreadContext(echo.New(), tt.body)

			sut := thread.NewCreateThreadHandler(usecase)
			err := sut.CreateThread(c)

			require.NoError(t, err)
			assert.Equal(t, tt.wantStatus, rec.Code)
			assert.Equal(t, tt.wantBody, rec.Body.String())
		})
	}
}

func TestCreateThreadHandler_CreateThread_Usecaseに渡す値(t *testing.T) {
	tests := []struct {
		name      string
		body      string
		setupMock func(*mock_thread.MockcreateThreadUsecase)
	}{
		{
			name: "リクエストが正常なとき、Usecaseに、リクエストのbody・file_idと、リクエスト元のIPを持つCreateThreadCommandが渡されること",
			body: validCreateThreadRequestBody,
			setupMock: func(m *mock_thread.MockcreateThreadUsecase) {
				m.EXPECT().Exec(gomock.Any(), applicationthread.CreateThreadCommand{
					IP:     "192.0.2.1",
					Body:   "hello",
					FileID: "file-1",
				}).Return(&applicationthread.CreateThreadOutput{ThreadID: "thread-1"}, nil)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			usecase := mock_thread.NewMockcreateThreadUsecase(ctrl)
			tt.setupMock(usecase)
			c, _ := newCreateThreadContext(echo.New(), tt.body)

			sut := thread.NewCreateThreadHandler(usecase)
			err := sut.CreateThread(c)

			require.NoError(t, err)
		})
	}
}

func TestCreateThreadHandler_CreateThread_Usecaseに渡すIP(t *testing.T) {
	tests := []struct {
		name              string
		trustedProxyCIDRs []string
		remoteAddr        string
		xForwardedFor     []string
		setupMock         func(*mock_thread.MockcreateThreadUsecase)
	}{
		{
			name:              "信用するCIDR外の接続元からX-Forwarded-Forに偽のIPを付けて送ったとき、接続元IPを持つCreateThreadCommandがUsecaseに渡されること",
			trustedProxyCIDRs: []string{"172.28.0.0/24"},
			remoteAddr:        "203.0.113.9:12345",
			xForwardedFor:     []string{"198.51.100.1"},
			setupMock: func(m *mock_thread.MockcreateThreadUsecase) {
				m.EXPECT().Exec(gomock.Any(), applicationthread.CreateThreadCommand{
					IP:     "203.0.113.9",
					Body:   "hello",
					FileID: "file-1",
				}).Return(&applicationthread.CreateThreadOutput{ThreadID: "thread-1"}, nil)
			},
		},
		{
			name:              "信用するCIDR内の接続元(BFF)からX-Forwarded-For付きで送ったとき、X-Forwarded-ForのIPを持つCreateThreadCommandがUsecaseに渡されること",
			trustedProxyCIDRs: []string{"172.28.0.0/24"},
			remoteAddr:        "172.28.0.5:12345",
			xForwardedFor:     []string{"198.51.100.1"},
			setupMock: func(m *mock_thread.MockcreateThreadUsecase) {
				m.EXPECT().Exec(gomock.Any(), applicationthread.CreateThreadCommand{
					IP:     "198.51.100.1",
					Body:   "hello",
					FileID: "file-1",
				}).Return(&applicationthread.CreateThreadOutput{ThreadID: "thread-1"}, nil)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			usecase := mock_thread.NewMockcreateThreadUsecase(ctrl)
			tt.setupMock(usecase)
			ipExtractor, err := handler.NewIPExtractor(tt.trustedProxyCIDRs)
			require.NoError(t, err)
			e := echo.New()
			e.IPExtractor = ipExtractor
			c, _ := newCreateThreadContext(e, validCreateThreadRequestBody)
			c.Request().RemoteAddr = tt.remoteAddr
			c.Request().Header[echo.HeaderXForwardedFor] = tt.xForwardedFor

			sut := thread.NewCreateThreadHandler(usecase)
			err = sut.CreateThread(c)

			require.NoError(t, err)
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
			name:       "リクエストボディのJSONの形式が不正なとき、ステータス400かつ「リクエストの形式が不正です。」が返ること",
			body:       `{"thread":`,
			setupMock:  func(m *mock_thread.MockcreateThreadUsecase) {},
			wantStatus: http.StatusBadRequest,
			wantBody:   `{"message":"リクエストの形式が不正です。"}` + "\n",
		},
		{
			name: "UsecaseがErrInvalidBodyを返すとき、ステータス400かつ「本文を入力してください。」が返ること",
			body: validCreateThreadRequestBody,
			setupMock: func(m *mock_thread.MockcreateThreadUsecase) {
				m.EXPECT().Exec(gomock.Any(), gomock.Any()).Return(nil, applicationthread.ErrInvalidBody)
			},
			wantStatus: http.StatusBadRequest,
			wantBody:   `{"message":"本文を入力してください。"}` + "\n",
		},
		{
			name: "UsecaseがErrFileRequiredを返すとき、ステータス400かつ「画像ファイルを添付してください。」が返ること",
			body: validCreateThreadRequestBody,
			setupMock: func(m *mock_thread.MockcreateThreadUsecase) {
				m.EXPECT().Exec(gomock.Any(), gomock.Any()).Return(nil, applicationthread.ErrFileRequired)
			},
			wantStatus: http.StatusBadRequest,
			wantBody:   `{"message":"画像ファイルを添付してください。"}` + "\n",
		},
		{
			name: "UsecaseがErrFileNotFoundを返すとき、ステータス400かつ「存在するファイルを指定してください。」が返ること",
			body: validCreateThreadRequestBody,
			setupMock: func(m *mock_thread.MockcreateThreadUsecase) {
				m.EXPECT().Exec(gomock.Any(), gomock.Any()).Return(nil, applicationthread.ErrFileNotFound)
			},
			wantStatus: http.StatusBadRequest,
			wantBody:   `{"message":"存在するファイルを指定してください。"}` + "\n",
		},
		{
			name: "UsecaseがErrFileAlreadyUsedを返すとき、ステータス400かつ「この画像ファイルは既に使用されています。」が返ること",
			body: validCreateThreadRequestBody,
			setupMock: func(m *mock_thread.MockcreateThreadUsecase) {
				m.EXPECT().Exec(gomock.Any(), gomock.Any()).Return(nil, applicationthread.ErrFileAlreadyUsed)
			},
			wantStatus: http.StatusBadRequest,
			wantBody:   `{"message":"この画像ファイルは既に使用されています。"}` + "\n",
		},
		{
			name: "Usecaseが上記以外のエラーを返すとき、ステータス500かつ「スレッドの作成に失敗しました。」が返り、エラー詳細はレスポンスに含まれないこと",
			body: validCreateThreadRequestBody,
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
			c, rec := newCreateThreadContext(echo.New(), tt.body)

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
		body          string
		setupMock     func(*mock_thread.MockcreateThreadUsecase)
		wantLogSubstr string
	}{
		{
			name: "Usecaseが上記以外のエラーを返すとき、エラー詳細がログに出力されること",
			body: validCreateThreadRequestBody,
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
			c, _ := newCreateThreadContext(e, tt.body)

			sut := thread.NewCreateThreadHandler(usecase)
			err := sut.CreateThread(c)

			require.NoError(t, err)
			assert.Contains(t, logBuf.String(), tt.wantLogSubstr)
		})
	}
}
