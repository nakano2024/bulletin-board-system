package thread_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"

	"github.com/nakanokota/bulletin-board-system/backend/application/mock/thread"
	"github.com/nakanokota/bulletin-board-system/backend/application/usecase/thread"
	domainthread "github.com/nakanokota/bulletin-board-system/backend/domain/thread"
)

func TestListThreadsUsecase_Exec_正常系(t *testing.T) {
	fixedTime := time.Date(2026, 9, 3, 12, 0, 0, 0, time.UTC)
	imagePath := "images/thread1.png"

	filePath, _ := domainthread.NewFilePath("images/", "thread1.png")
	threadWithImage, _ := domainthread.NewThread("thread-1", "hello", filePath, fixedTime)
	secondThreadWithImage, _ := domainthread.NewThread("thread-2", "second", filePath, fixedTime.Add(time.Minute))

	tests := []struct {
		name      string
		setupMock func(*mock_thread.MockIThreadRepository)
		want      *thread.ListThreadsOutput
	}{
		{
			name: "スレッドが1件存在するとき、Outputの要素にID/Body/ImagePath/CreatedAtが正しく変換されて含まれること",
			setupMock: func(m *mock_thread.MockIThreadRepository) {
				m.EXPECT().FetchActiveThreadListNewestFirst(gomock.Any()).Return(domainthread.NewThreadList([]*domainthread.Thread{threadWithImage}), nil)
			},
			want: &thread.ListThreadsOutput{
				Threads: []thread.ListThreadsOutputThread{
					{ID: "thread-1", Body: "hello", ImagePath: imagePath, CreatedAt: fixedTime},
				},
			},
		},
		{
			name: "スレッドが複数件存在するとき、Outputの件数・順序が一致すること",
			setupMock: func(m *mock_thread.MockIThreadRepository) {
				m.EXPECT().FetchActiveThreadListNewestFirst(gomock.Any()).Return(domainthread.NewThreadList([]*domainthread.Thread{threadWithImage, secondThreadWithImage}), nil)
			},
			want: &thread.ListThreadsOutput{
				Threads: []thread.ListThreadsOutputThread{
					{ID: "thread-1", Body: "hello", ImagePath: imagePath, CreatedAt: fixedTime},
					{ID: "thread-2", Body: "second", ImagePath: imagePath, CreatedAt: fixedTime.Add(time.Minute)},
				},
			},
		},
		{
			name: "スレッドが0件のとき、Outputの Threads が空スライスであること",
			setupMock: func(m *mock_thread.MockIThreadRepository) {
				m.EXPECT().FetchActiveThreadListNewestFirst(gomock.Any()).Return(domainthread.NewThreadList([]*domainthread.Thread{}), nil)
			},
			want: &thread.ListThreadsOutput{
				Threads: []thread.ListThreadsOutputThread{},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			threadRepository := mock_thread.NewMockIThreadRepository(ctrl)
			tt.setupMock(threadRepository)
			logger := mock_thread.NewMockILogger(ctrl)

			sut := thread.NewListThreadsUsecase(threadRepository, logger)
			got, err := sut.Exec(context.Background(), thread.ListThreadsCommand{})

			require.NoError(t, err)
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestListThreadsUsecase_Exec_異常系(t *testing.T) {
	errRepositoryFailed := errors.New("repository failed")

	tests := []struct {
		name      string
		setupMock func(*mock_thread.MockIThreadRepository)
		wantErr   error
	}{
		{
			name: "スレッド一覧の取得に失敗するとき、Execはエラーを返すこと",
			setupMock: func(m *mock_thread.MockIThreadRepository) {
				m.EXPECT().FetchActiveThreadListNewestFirst(gomock.Any()).Return(nil, errRepositoryFailed)
			},
			wantErr: errRepositoryFailed,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			threadRepository := mock_thread.NewMockIThreadRepository(ctrl)
			tt.setupMock(threadRepository)
			logger := mock_thread.NewMockILogger(ctrl)
			logger.EXPECT().Error(gomock.Any(), gomock.Any()).AnyTimes()

			sut := thread.NewListThreadsUsecase(threadRepository, logger)
			_, err := sut.Exec(context.Background(), thread.ListThreadsCommand{})

			require.ErrorIs(t, err, tt.wantErr)
		})
	}
}

func TestListThreadsUsecase_Exec_ログ出力(t *testing.T) {
	errRepositoryFailed := errors.New("repository failed")

	tests := []struct {
		name       string
		setupMock  func(*mock_thread.MockIThreadRepository)
		wantLogErr error
	}{
		{
			name: "スレッド一覧の取得に失敗するとき、ILoggerにリポジトリのエラーがそのまま渡されること",
			setupMock: func(m *mock_thread.MockIThreadRepository) {
				m.EXPECT().FetchActiveThreadListNewestFirst(gomock.Any()).Return(nil, errRepositoryFailed)
			},
			wantLogErr: errRepositoryFailed,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			threadRepository := mock_thread.NewMockIThreadRepository(ctrl)
			tt.setupMock(threadRepository)
			logger := mock_thread.NewMockILogger(ctrl)
			logger.EXPECT().Error(gomock.Any(), tt.wantLogErr)

			sut := thread.NewListThreadsUsecase(threadRepository, logger)
			_, err := sut.Exec(context.Background(), thread.ListThreadsCommand{})

			require.Error(t, err)
		})
	}
}
