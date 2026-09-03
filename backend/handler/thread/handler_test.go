package thread_test

import (
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/labstack/echo/v4"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"

	applicationthread "github.com/nakanokota/bulletin-board-system/backend/application/thread"
	"github.com/nakanokota/bulletin-board-system/backend/handler/thread"
	"github.com/nakanokota/bulletin-board-system/backend/handler/thread/mock_thread"
)

func TestThreadHandler_ListThreads_正常系(t *testing.T) {
	fixedTime := time.Date(2026, 9, 3, 12, 0, 0, 0, time.UTC)
	imagePath := "images/thread-1.png"

	tests := []struct {
		name       string
		setupMock  func(*mock_thread.MocklistThreadsUsecase)
		wantStatus int
		wantBody   string
	}{
		{
			name: "Usecaseが1件のOutputを返すとき、ステータス200かつレスポンスボディが1件の要素(image_path含む)を含むこと",
			setupMock: func(m *mock_thread.MocklistThreadsUsecase) {
				m.EXPECT().Exec(gomock.Any(), applicationthread.ListThreadsCommand{}).Return(&applicationthread.ListThreadsOutput{
					Threads: []applicationthread.ListThreadsOutputThread{
						{ID: "thread-1", Body: "hello", ImagePath: &imagePath, CreatedAt: fixedTime},
					},
				}, nil)
			},
			wantStatus: http.StatusOK,
			wantBody:   `{"threads":[{"id":"thread-1","body":"hello","image_path":"images/thread-1.png","created_at":"2026-09-03T12:00:00Z"}]}` + "\n",
		},
		{
			name: "Usecaseが複数件のOutputを返すとき、ステータス200かつレスポンスボディが各要素(image_pathがnullの要素含む)を含むこと",
			setupMock: func(m *mock_thread.MocklistThreadsUsecase) {
				m.EXPECT().Exec(gomock.Any(), applicationthread.ListThreadsCommand{}).Return(&applicationthread.ListThreadsOutput{
					Threads: []applicationthread.ListThreadsOutputThread{
						{ID: "thread-1", Body: "first", ImagePath: nil, CreatedAt: fixedTime},
						{ID: "thread-2", Body: "second", ImagePath: &imagePath, CreatedAt: fixedTime.Add(time.Minute)},
					},
				}, nil)
			},
			wantStatus: http.StatusOK,
			wantBody:   `{"threads":[{"id":"thread-1","body":"first","image_path":null,"created_at":"2026-09-03T12:00:00Z"},{"id":"thread-2","body":"second","image_path":"images/thread-1.png","created_at":"2026-09-03T12:01:00Z"}]}` + "\n",
		},
		{
			name: "Usecaseが空のOutputを返すとき、ステータス200かつレスポンスボディが空配列であること",
			setupMock: func(m *mock_thread.MocklistThreadsUsecase) {
				m.EXPECT().Exec(gomock.Any(), applicationthread.ListThreadsCommand{}).Return(&applicationthread.ListThreadsOutput{
					Threads: []applicationthread.ListThreadsOutputThread{},
				}, nil)
			},
			wantStatus: http.StatusOK,
			wantBody:   `{"threads":[]}` + "\n",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			usecase := mock_thread.NewMocklistThreadsUsecase(ctrl)
			tt.setupMock(usecase)

			e := echo.New()
			req := httptest.NewRequest(http.MethodGet, "/threads", nil)
			rec := httptest.NewRecorder()
			c := e.NewContext(req, rec)

			sut := thread.NewThreadHandler(usecase)
			err := sut.ListThreads(c)

			require.NoError(t, err)
			assert.Equal(t, tt.wantStatus, rec.Code)
			assert.Equal(t, tt.wantBody, rec.Body.String())
		})
	}
}

func TestThreadHandler_ListThreads_異常系(t *testing.T) {
	tests := []struct {
		name       string
		setupMock  func(*mock_thread.MocklistThreadsUsecase)
		wantStatus int
		wantBody   string
	}{
		{
			name: "Usecaseがエラーを返すとき、ステータス500かつ固定メッセージを含むレスポンスが返ること",
			setupMock: func(m *mock_thread.MocklistThreadsUsecase) {
				m.EXPECT().Exec(gomock.Any(), applicationthread.ListThreadsCommand{}).Return(nil, errors.New("query failed"))
			},
			wantStatus: http.StatusInternalServerError,
			wantBody:   `{"message":"スレッド一覧の取得に失敗しました。"}` + "\n",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			usecase := mock_thread.NewMocklistThreadsUsecase(ctrl)
			tt.setupMock(usecase)

			e := echo.New()
			req := httptest.NewRequest(http.MethodGet, "/threads", nil)
			rec := httptest.NewRecorder()
			c := e.NewContext(req, rec)

			sut := thread.NewThreadHandler(usecase)
			err := sut.ListThreads(c)

			require.NoError(t, err)
			assert.Equal(t, tt.wantStatus, rec.Code)
			assert.Equal(t, tt.wantBody, rec.Body.String())
		})
	}
}
