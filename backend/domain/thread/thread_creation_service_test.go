package thread_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"

	"github.com/nakanokota/bulletin-board-system/backend/domain/thread"
	"github.com/nakanokota/bulletin-board-system/backend/domain/thread/mock_thread"
)

func TestThreadCreationService_Create_正常系(t *testing.T) {
	pendingThread, _ := thread.NewPendingThread("user-1", "hello", "file-1")
	filePath, _ := thread.NewFilePath("thread_images/", "sample.png")
	createdThread, _ := thread.NewThread("thread-1", "hello", filePath, time.Date(2026, 9, 27, 12, 0, 0, 0, time.UTC))

	tests := []struct {
		name      string
		setupMock func(*mock_thread.MockIThreadFileChecker, *mock_thread.MockIPendingThreadRepository)
		want      *thread.Thread
	}{
		{
			name: "ファイルが存在し、他スレッドに未使用のとき、CreateThreadが返したThreadが返ること",
			setupMock: func(c *mock_thread.MockIThreadFileChecker, r *mock_thread.MockIPendingThreadRepository) {
				c.EXPECT().ExistsFile(gomock.Any(), gomock.Any()).Return(true, nil)
				c.EXPECT().IsFileAttachedToThread(gomock.Any(), gomock.Any()).Return(false, nil)
				r.EXPECT().CreateThread(gomock.Any(), gomock.Any()).Return(createdThread, nil)
			},
			want: createdThread,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			fileChecker := mock_thread.NewMockIThreadFileChecker(ctrl)
			pendingThreadRepository := mock_thread.NewMockIPendingThreadRepository(ctrl)
			tt.setupMock(fileChecker, pendingThreadRepository)

			sut := thread.NewThreadCreationService(fileChecker, pendingThreadRepository)
			got, err := sut.Create(context.Background(), pendingThread)

			require.NoError(t, err)
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestThreadCreationService_Create_依存に渡す値(t *testing.T) {
	pendingThread, _ := thread.NewPendingThread("user-1", "hello", "file-1")
	filePath, _ := thread.NewFilePath("thread_images/", "sample.png")
	createdThread, _ := thread.NewThread("thread-1", "hello", filePath, time.Date(2026, 9, 27, 12, 0, 0, 0, time.UTC))

	tests := []struct {
		name      string
		setupMock func(*mock_thread.MockIThreadFileChecker, *mock_thread.MockIPendingThreadRepository)
	}{
		{
			name: "ExistsFileに、PendingThreadのfileIDが渡されること",
			setupMock: func(c *mock_thread.MockIThreadFileChecker, r *mock_thread.MockIPendingThreadRepository) {
				c.EXPECT().ExistsFile(gomock.Any(), "file-1").Return(true, nil)
				c.EXPECT().IsFileAttachedToThread(gomock.Any(), gomock.Any()).Return(false, nil)
				r.EXPECT().CreateThread(gomock.Any(), gomock.Any()).Return(createdThread, nil)
			},
		},
		{
			name: "IsFileAttachedToThreadに、PendingThreadのfileIDが渡されること",
			setupMock: func(c *mock_thread.MockIThreadFileChecker, r *mock_thread.MockIPendingThreadRepository) {
				c.EXPECT().ExistsFile(gomock.Any(), gomock.Any()).Return(true, nil)
				c.EXPECT().IsFileAttachedToThread(gomock.Any(), "file-1").Return(false, nil)
				r.EXPECT().CreateThread(gomock.Any(), gomock.Any()).Return(createdThread, nil)
			},
		},
		{
			name: "CreateThreadに、引数で受け取ったPendingThreadが渡されること",
			setupMock: func(c *mock_thread.MockIThreadFileChecker, r *mock_thread.MockIPendingThreadRepository) {
				c.EXPECT().ExistsFile(gomock.Any(), gomock.Any()).Return(true, nil)
				c.EXPECT().IsFileAttachedToThread(gomock.Any(), gomock.Any()).Return(false, nil)
				r.EXPECT().CreateThread(gomock.Any(), pendingThread).Return(createdThread, nil)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			fileChecker := mock_thread.NewMockIThreadFileChecker(ctrl)
			pendingThreadRepository := mock_thread.NewMockIPendingThreadRepository(ctrl)
			tt.setupMock(fileChecker, pendingThreadRepository)

			sut := thread.NewThreadCreationService(fileChecker, pendingThreadRepository)
			_, err := sut.Create(context.Background(), pendingThread)

			require.NoError(t, err)
		})
	}
}

func TestThreadCreationService_Create_異常系(t *testing.T) {
	pendingThread, _ := thread.NewPendingThread("user-1", "hello", "file-1")
	errDependencyFailed := errors.New("dependency failed")

	tests := []struct {
		name      string
		setupMock func(*mock_thread.MockIThreadFileChecker, *mock_thread.MockIPendingThreadRepository)
		wantErr   error
	}{
		{
			name: "ファイルが存在しないとき、ErrFileNotFoundが返ること",
			setupMock: func(c *mock_thread.MockIThreadFileChecker, r *mock_thread.MockIPendingThreadRepository) {
				c.EXPECT().ExistsFile(gomock.Any(), gomock.Any()).Return(false, nil)
			},
			wantErr: thread.ErrFileNotFound,
		},
		{
			name: "ファイルが存在し、他スレッドで使用済みのとき、ErrFileAlreadyUsedが返ること",
			setupMock: func(c *mock_thread.MockIThreadFileChecker, r *mock_thread.MockIPendingThreadRepository) {
				c.EXPECT().ExistsFile(gomock.Any(), gomock.Any()).Return(true, nil)
				c.EXPECT().IsFileAttachedToThread(gomock.Any(), gomock.Any()).Return(true, nil)
			},
			wantErr: thread.ErrFileAlreadyUsed,
		},
		{
			name: "ExistsFileがエラーを返したとき、そのエラーが返ること",
			setupMock: func(c *mock_thread.MockIThreadFileChecker, r *mock_thread.MockIPendingThreadRepository) {
				c.EXPECT().ExistsFile(gomock.Any(), gomock.Any()).Return(false, errDependencyFailed)
			},
			wantErr: errDependencyFailed,
		},
		{
			name: "IsFileAttachedToThreadがエラーを返したとき、そのエラーが返ること",
			setupMock: func(c *mock_thread.MockIThreadFileChecker, r *mock_thread.MockIPendingThreadRepository) {
				c.EXPECT().ExistsFile(gomock.Any(), gomock.Any()).Return(true, nil)
				c.EXPECT().IsFileAttachedToThread(gomock.Any(), gomock.Any()).Return(false, errDependencyFailed)
			},
			wantErr: errDependencyFailed,
		},
		{
			name: "CreateThreadがエラーを返したとき、そのエラーが返ること",
			setupMock: func(c *mock_thread.MockIThreadFileChecker, r *mock_thread.MockIPendingThreadRepository) {
				c.EXPECT().ExistsFile(gomock.Any(), gomock.Any()).Return(true, nil)
				c.EXPECT().IsFileAttachedToThread(gomock.Any(), gomock.Any()).Return(false, nil)
				r.EXPECT().CreateThread(gomock.Any(), gomock.Any()).Return(nil, thread.ErrFileAlreadyUsed)
			},
			wantErr: thread.ErrFileAlreadyUsed,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			fileChecker := mock_thread.NewMockIThreadFileChecker(ctrl)
			pendingThreadRepository := mock_thread.NewMockIPendingThreadRepository(ctrl)
			tt.setupMock(fileChecker, pendingThreadRepository)

			sut := thread.NewThreadCreationService(fileChecker, pendingThreadRepository)
			_, err := sut.Create(context.Background(), pendingThread)

			require.ErrorIs(t, err, tt.wantErr)
		})
	}
}
