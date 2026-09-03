package thread_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"

	"github.com/nakanokota/bulletin-board-system/backend/application/thread"
	"github.com/nakanokota/bulletin-board-system/backend/application/thread/mock_thread"
)

func TestListThreadsUsecase_Exec_正常系(t *testing.T) {
	fixedTime := time.Date(2026, 9, 3, 12, 0, 0, 0, time.UTC)

	tests := []struct {
		name      string
		setupMock func(*mock_thread.MockIThreadQueryService)
		want      *thread.ListThreadsOutput
	}{
		{
			name: "QueryServiceが1件のスレッドを返すとき、Outputの要素にID/Body/CreatedAtが正しく変換されて含まれること",
			setupMock: func(m *mock_thread.MockIThreadQueryService) {
				m.EXPECT().FetchThreadList(gomock.Any()).Return([]thread.ThreadListItem{
					{ID: "thread-1", Body: "hello", CreatedAt: fixedTime},
				}, nil)
			},
			want: &thread.ListThreadsOutput{
				Threads: []thread.ListThreadsOutputThread{
					{ID: "thread-1", Body: "hello", CreatedAt: fixedTime},
				},
			},
		},
		{
			name: "QueryServiceが複数件のスレッドを返すとき、Outputの件数・順序がQueryServiceの返却順と一致すること",
			setupMock: func(m *mock_thread.MockIThreadQueryService) {
				m.EXPECT().FetchThreadList(gomock.Any()).Return([]thread.ThreadListItem{
					{ID: "thread-1", Body: "first", CreatedAt: fixedTime},
					{ID: "thread-2", Body: "second", CreatedAt: fixedTime.Add(time.Minute)},
				}, nil)
			},
			want: &thread.ListThreadsOutput{
				Threads: []thread.ListThreadsOutputThread{
					{ID: "thread-1", Body: "first", CreatedAt: fixedTime},
					{ID: "thread-2", Body: "second", CreatedAt: fixedTime.Add(time.Minute)},
				},
			},
		},
		{
			name: "QueryServiceが0件を返すとき、Outputの Threads が空スライスであること",
			setupMock: func(m *mock_thread.MockIThreadQueryService) {
				m.EXPECT().FetchThreadList(gomock.Any()).Return([]thread.ThreadListItem{}, nil)
			},
			want: &thread.ListThreadsOutput{
				Threads: []thread.ListThreadsOutputThread{},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			queryService := mock_thread.NewMockIThreadQueryService(ctrl)
			tt.setupMock(queryService)

			sut := thread.NewListThreadsUsecase(queryService)
			got, err := sut.Exec(context.Background(), thread.ListThreadsCommand{})

			require.NoError(t, err)
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestListThreadsUsecase_Exec_異常系(t *testing.T) {
	errQueryFailed := errors.New("query failed")

	tests := []struct {
		name      string
		setupMock func(*mock_thread.MockIThreadQueryService)
		wantErr   error
	}{
		{
			name: "QueryServiceがエラーを返すとき、Execはエラーを返すこと",
			setupMock: func(m *mock_thread.MockIThreadQueryService) {
				m.EXPECT().FetchThreadList(gomock.Any()).Return(nil, errQueryFailed)
			},
			wantErr: errQueryFailed,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			queryService := mock_thread.NewMockIThreadQueryService(ctrl)
			tt.setupMock(queryService)

			sut := thread.NewListThreadsUsecase(queryService)
			_, err := sut.Exec(context.Background(), thread.ListThreadsCommand{})

			require.ErrorIs(t, err, tt.wantErr)
		})
	}
}
