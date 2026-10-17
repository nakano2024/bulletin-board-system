package thread_test

import (
	"bytes"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/labstack/echo/v4"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"

	applicationthread "github.com/nakanokota/bulletin-board-system/backend/application/usecase/thread"
	"github.com/nakanokota/bulletin-board-system/backend/handler/thread"
	"github.com/nakanokota/bulletin-board-system/backend/handler/thread/mock_thread"
)

func TestGetThreadHandler_GetThread_正常系(t *testing.T) {
	fixedTime := time.Date(2026, 10, 17, 12, 0, 0, 0, time.UTC)

	tests := []struct {
		name       string
		id         string
		setupMock  func(*mock_thread.MockgetThreadUsecase)
		wantStatus int
		wantBody   string
	}{
		{
			name: "Usecaseがレス0件のOutputを返すとき、ステータス200かつレスポンスボディがスレッドのid/user_id/body/image_path/created_atと空のrepliesを含むこと",
			id:   "thread-1",
			setupMock: func(m *mock_thread.MockgetThreadUsecase) {
				m.EXPECT().Exec(gomock.Any(), gomock.Any()).Return(&applicationthread.GetThreadOutput{
					Thread: applicationthread.GetThreadOutputThread{
						ID: "thread-1", UserID: "user-1", Body: "hello", ImagePath: "thread_images/thread.png", CreatedAt: fixedTime,
						Replies: []applicationthread.GetThreadOutputReply{},
					},
				}, nil)
			},
			wantStatus: http.StatusOK,
			wantBody:   `{"thread":{"id":"thread-1","user_id":"user-1","body":"hello","image_path":"thread_images/thread.png","created_at":"2026-10-17T12:00:00Z","replies":[]}}` + "\n",
		},
		{
			name: "Usecaseがレス2件(画像あり・画像なし)のOutputを返すとき、ステータス200かつrepliesがOutputの順序どおりに各レスのid/user_id/body/image_path/created_atを含み、画像なしのレスのimage_pathが空文字であること",
			id:   "thread-1",
			setupMock: func(m *mock_thread.MockgetThreadUsecase) {
				m.EXPECT().Exec(gomock.Any(), gomock.Any()).Return(&applicationthread.GetThreadOutput{
					Thread: applicationthread.GetThreadOutputThread{
						ID: "thread-1", UserID: "user-1", Body: "hello", ImagePath: "thread_images/thread.png", CreatedAt: fixedTime,
						Replies: []applicationthread.GetThreadOutputReply{
							{ID: "reply-2", UserID: "user-3", Body: "second", ImagePath: "", CreatedAt: fixedTime.Add(2 * time.Minute)},
							{ID: "reply-1", UserID: "user-2", Body: "first", ImagePath: "thread_images/reply.png", CreatedAt: fixedTime.Add(time.Minute)},
						},
					},
				}, nil)
			},
			wantStatus: http.StatusOK,
			wantBody: `{"thread":{"id":"thread-1","user_id":"user-1","body":"hello","image_path":"thread_images/thread.png","created_at":"2026-10-17T12:00:00Z","replies":[` +
				`{"id":"reply-2","user_id":"user-3","body":"second","image_path":"","created_at":"2026-10-17T12:02:00Z"},` +
				`{"id":"reply-1","user_id":"user-2","body":"first","image_path":"thread_images/reply.png","created_at":"2026-10-17T12:01:00Z"}]}}` + "\n",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			usecase := mock_thread.NewMockgetThreadUsecase(ctrl)
			tt.setupMock(usecase)

			e := echo.New()
			req := httptest.NewRequest(http.MethodGet, "/threads/"+tt.id, nil)
			rec := httptest.NewRecorder()
			c := e.NewContext(req, rec)
			c.SetParamNames("id")
			c.SetParamValues(tt.id)

			sut := thread.NewGetThreadHandler(usecase)
			err := sut.GetThread(c)

			require.NoError(t, err)
			assert.Equal(t, tt.wantStatus, rec.Code)
			assert.Equal(t, tt.wantBody, rec.Body.String())
		})
	}
}

func TestGetThreadHandler_GetThread_異常系(t *testing.T) {
	tests := []struct {
		name          string
		id            string
		setupMock     func(*mock_thread.MockgetThreadUsecase)
		wantStatus    int
		wantBody      string
		wantLogSubstr string
	}{
		{
			name: "UsecaseがErrThreadNotFoundを返すとき、ステータス404かつ「スレッドが存在しません。」が返ること",
			id:   "thread-missing",
			setupMock: func(m *mock_thread.MockgetThreadUsecase) {
				m.EXPECT().Exec(gomock.Any(), gomock.Any()).Return(nil, applicationthread.ErrThreadNotFound)
			},
			wantStatus:    http.StatusNotFound,
			wantBody:      `{"message":"スレッドが存在しません。"}` + "\n",
			wantLogSubstr: "",
		},
		{
			name: "UsecaseがErrThreadDeadを返すとき、ステータス404かつ「スレッドが存在しません。」が返ること",
			id:   "thread-dead",
			setupMock: func(m *mock_thread.MockgetThreadUsecase) {
				m.EXPECT().Exec(gomock.Any(), gomock.Any()).Return(nil, applicationthread.ErrThreadDead)
			},
			wantStatus:    http.StatusNotFound,
			wantBody:      `{"message":"スレッドが存在しません。"}` + "\n",
			wantLogSubstr: "",
		},
		{
			name: "Usecaseがその他のエラーを返すとき、ステータス500かつ「スレッドを開くのに失敗しました。」が返り、エラー詳細がログに出力されること",
			id:   "thread-1",
			setupMock: func(m *mock_thread.MockgetThreadUsecase) {
				m.EXPECT().Exec(gomock.Any(), gomock.Any()).Return(nil, errors.New("query failed"))
			},
			wantStatus:    http.StatusInternalServerError,
			wantBody:      `{"message":"スレッドを開くのに失敗しました。"}` + "\n",
			wantLogSubstr: "query failed",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			usecase := mock_thread.NewMockgetThreadUsecase(ctrl)
			tt.setupMock(usecase)

			var logBuf bytes.Buffer
			e := echo.New()
			e.Logger.SetOutput(&logBuf)
			req := httptest.NewRequest(http.MethodGet, "/threads/"+tt.id, nil)
			rec := httptest.NewRecorder()
			c := e.NewContext(req, rec)
			c.SetParamNames("id")
			c.SetParamValues(tt.id)

			sut := thread.NewGetThreadHandler(usecase)
			err := sut.GetThread(c)

			require.NoError(t, err)
			assert.Equal(t, tt.wantStatus, rec.Code)
			assert.Equal(t, tt.wantBody, rec.Body.String())
			assert.Contains(t, logBuf.String(), tt.wantLogSubstr)
		})
	}
}

func TestGetThreadHandler_GetThread_依存への入力(t *testing.T) {
	tests := []struct {
		name      string
		id        string
		setupMock func(*mock_thread.MockgetThreadUsecase)
	}{
		{
			name: "パスパラメータidの値が、GetThreadCommandのThreadIDにそのまま渡されること",
			id:   "thread-1",
			setupMock: func(m *mock_thread.MockgetThreadUsecase) {
				m.EXPECT().Exec(gomock.Any(), applicationthread.GetThreadCommand{ThreadID: "thread-1"}).Return(&applicationthread.GetThreadOutput{
					Thread: applicationthread.GetThreadOutputThread{Replies: []applicationthread.GetThreadOutputReply{}},
				}, nil)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			usecase := mock_thread.NewMockgetThreadUsecase(ctrl)
			tt.setupMock(usecase)

			e := echo.New()
			req := httptest.NewRequest(http.MethodGet, "/threads/"+tt.id, nil)
			rec := httptest.NewRecorder()
			c := e.NewContext(req, rec)
			c.SetParamNames("id")
			c.SetParamValues(tt.id)

			sut := thread.NewGetThreadHandler(usecase)
			err := sut.GetThread(c)

			require.NoError(t, err)
		})
	}
}
