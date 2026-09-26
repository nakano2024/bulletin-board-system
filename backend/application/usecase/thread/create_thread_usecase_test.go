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
	domainuser "github.com/nakanokota/bulletin-board-system/backend/domain/user"
	"github.com/nakanokota/bulletin-board-system/backend/domain/user/mock_user"
)

func TestCreateThreadUsecase_Exec_正常系(t *testing.T) {
	ip := "203.0.113.1"
	now := time.Date(2026, 9, 13, 15, 30, 0, 0, time.UTC)
	date, _ := domainuser.NewUserCreateDate(time.Date(2026, 9, 13, 0, 0, 0, 0, time.UTC))

	fileName, _ := domainthread.NewFileName("sample.png")
	existingUser, _ := domainuser.NewUser("user-1", ip)
	newPendingUser, _ := domainuser.NewPendingUser(ip, date)
	newUser, _ := domainuser.NewUser("user-2", ip)

	pendingThreadForExistingUser, _ := domainthread.NewPendingThread("user-1", "hello", fileName)
	threadForExistingUser, _ := domainthread.NewThread("thread-1", "hello", mustFilePath(t), now)

	pendingThreadForNewUser, _ := domainthread.NewPendingThread("user-2", "hello", fileName)
	threadForNewUser, _ := domainthread.NewThread("thread-2", "hello", mustFilePath(t), now)

	tests := []struct {
		name         string
		setupMocks   func(*mock_user.MockIUserRepository, *mock_thread.MockIPendingThreadRepository, *mock_thread.MockITimeGetter)
		wantThreadID string
	}{
		{
			name: "同一IP・同一日のUserが既に存在するとき、そのUserのIDでThreadが作成され、ThreadIDがOutputに含まれること",
			setupMocks: func(userRepo *mock_user.MockIUserRepository, threadRepo *mock_thread.MockIPendingThreadRepository, timeGetter *mock_thread.MockITimeGetter) {
				timeGetter.EXPECT().Now(gomock.Any()).Return(now)
				userRepo.EXPECT().FindByIPAndDate(gomock.Any(), ip, date).Return(existingUser, nil)
				threadRepo.EXPECT().CreateThread(gomock.Any(), pendingThreadForExistingUser).Return(threadForExistingUser, nil)
			},
			wantThreadID: "thread-1",
		},
		{
			name: "同一IP・同一日のUserが存在しないとき、新規作成されたUserのIDでThreadが作成され、ThreadIDがOutputに含まれること",
			setupMocks: func(userRepo *mock_user.MockIUserRepository, threadRepo *mock_thread.MockIPendingThreadRepository, timeGetter *mock_thread.MockITimeGetter) {
				timeGetter.EXPECT().Now(gomock.Any()).Return(now)
				userRepo.EXPECT().FindByIPAndDate(gomock.Any(), ip, date).Return(nil, nil)
				userRepo.EXPECT().CreateUser(gomock.Any(), newPendingUser).Return(newUser, nil)
				threadRepo.EXPECT().CreateThread(gomock.Any(), pendingThreadForNewUser).Return(threadForNewUser, nil)
			},
			wantThreadID: "thread-2",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			userRepo := mock_user.NewMockIUserRepository(ctrl)
			threadRepo := mock_thread.NewMockIPendingThreadRepository(ctrl)
			timeGetter := mock_thread.NewMockITimeGetter(ctrl)
			logger := mock_thread.NewMockILogger(ctrl)
			tt.setupMocks(userRepo, threadRepo, timeGetter)

			userService := domainuser.NewUserService(userRepo)
			sut := thread.NewCreateThreadUsecase(threadRepo, userService, timeGetter, logger)

			got, err := sut.Exec(context.Background(), thread.CreateThreadCommand{
				IP:       ip,
				Body:     "hello",
				FileName: "sample.png",
			})

			require.NoError(t, err)
			assert.Equal(t, tt.wantThreadID, got.ThreadID)
		})
	}
}

func TestCreateThreadUsecase_Exec_異常系(t *testing.T) {
	ip := "203.0.113.1"
	now := time.Date(2026, 9, 13, 15, 30, 0, 0, time.UTC)
	date, _ := domainuser.NewUserCreateDate(time.Date(2026, 9, 13, 0, 0, 0, 0, time.UTC))
	errRepositoryFailed := errors.New("repository failed")

	tests := []struct {
		name       string
		cmd        thread.CreateThreadCommand
		setupMocks func(*mock_user.MockIUserRepository, *mock_thread.MockIPendingThreadRepository, *mock_thread.MockITimeGetter)
	}{
		{
			name: "fileNameの拡張子が許可されていないとき、Execはエラーを返すこと",
			cmd:  thread.CreateThreadCommand{IP: ip, Body: "hello", FileName: "sample.exe"},
			setupMocks: func(userRepo *mock_user.MockIUserRepository, threadRepo *mock_thread.MockIPendingThreadRepository, timeGetter *mock_thread.MockITimeGetter) {
				timeGetter.EXPECT().Now(gomock.Any()).Return(now).AnyTimes()
			},
		},
		{
			name: "bodyが空文字のとき、Execはエラーを返すこと",
			cmd:  thread.CreateThreadCommand{IP: ip, Body: "", FileName: "sample.png"},
			setupMocks: func(userRepo *mock_user.MockIUserRepository, threadRepo *mock_thread.MockIPendingThreadRepository, timeGetter *mock_thread.MockITimeGetter) {
				existingUser, _ := domainuser.NewUser("user-1", ip)
				timeGetter.EXPECT().Now(gomock.Any()).Return(now)
				userRepo.EXPECT().FindByIPAndDate(gomock.Any(), ip, date).Return(existingUser, nil)
				threadRepo.EXPECT().CreateThread(gomock.Any(), gomock.Any()).Times(0)
			},
		},
		{
			name: "Userの解決に失敗するとき、Execはエラーを返すこと",
			cmd:  thread.CreateThreadCommand{IP: ip, Body: "hello", FileName: "sample.png"},
			setupMocks: func(userRepo *mock_user.MockIUserRepository, threadRepo *mock_thread.MockIPendingThreadRepository, timeGetter *mock_thread.MockITimeGetter) {
				timeGetter.EXPECT().Now(gomock.Any()).Return(now)
				userRepo.EXPECT().FindByIPAndDate(gomock.Any(), ip, date).Return(nil, errRepositoryFailed)
				threadRepo.EXPECT().CreateThread(gomock.Any(), gomock.Any()).Times(0)
			},
		},
		{
			name: "ThreadRepository.CreateThreadがエラーを返すとき、Execはエラーを返すこと",
			cmd:  thread.CreateThreadCommand{IP: ip, Body: "hello", FileName: "sample.png"},
			setupMocks: func(userRepo *mock_user.MockIUserRepository, threadRepo *mock_thread.MockIPendingThreadRepository, timeGetter *mock_thread.MockITimeGetter) {
				existingUser, _ := domainuser.NewUser("user-1", ip)
				timeGetter.EXPECT().Now(gomock.Any()).Return(now)
				userRepo.EXPECT().FindByIPAndDate(gomock.Any(), ip, date).Return(existingUser, nil)
				threadRepo.EXPECT().CreateThread(gomock.Any(), gomock.Any()).Return(nil, errRepositoryFailed)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			userRepo := mock_user.NewMockIUserRepository(ctrl)
			threadRepo := mock_thread.NewMockIPendingThreadRepository(ctrl)
			timeGetter := mock_thread.NewMockITimeGetter(ctrl)
			logger := mock_thread.NewMockILogger(ctrl)
			logger.EXPECT().Error(gomock.Any(), gomock.Any()).AnyTimes()
			tt.setupMocks(userRepo, threadRepo, timeGetter)

			userService := domainuser.NewUserService(userRepo)
			sut := thread.NewCreateThreadUsecase(threadRepo, userService, timeGetter, logger)

			_, err := sut.Exec(context.Background(), tt.cmd)

			require.Error(t, err)
		})
	}
}

func TestCreateThreadUsecase_Exec_ログ出力(t *testing.T) {
	ip := "203.0.113.1"
	now := time.Date(2026, 9, 13, 15, 30, 0, 0, time.UTC)
	date, _ := domainuser.NewUserCreateDate(time.Date(2026, 9, 13, 0, 0, 0, 0, time.UTC))
	errRepositoryFailed := errors.New("repository failed")

	tests := []struct {
		name       string
		setupMocks func(*mock_user.MockIUserRepository, *mock_thread.MockIPendingThreadRepository, *mock_thread.MockITimeGetter)
	}{
		{
			name: "Userの解決に失敗するとき、ILoggerにそのエラーがそのまま渡されること",
			setupMocks: func(userRepo *mock_user.MockIUserRepository, threadRepo *mock_thread.MockIPendingThreadRepository, timeGetter *mock_thread.MockITimeGetter) {
				timeGetter.EXPECT().Now(gomock.Any()).Return(now)
				userRepo.EXPECT().FindByIPAndDate(gomock.Any(), ip, date).Return(nil, errRepositoryFailed)
			},
		},
		{
			name: "ThreadRepository.CreateThreadが失敗するとき、ILoggerにそのエラーがそのまま渡されること",
			setupMocks: func(userRepo *mock_user.MockIUserRepository, threadRepo *mock_thread.MockIPendingThreadRepository, timeGetter *mock_thread.MockITimeGetter) {
				existingUser, _ := domainuser.NewUser("user-1", ip)
				timeGetter.EXPECT().Now(gomock.Any()).Return(now)
				userRepo.EXPECT().FindByIPAndDate(gomock.Any(), ip, date).Return(existingUser, nil)
				threadRepo.EXPECT().CreateThread(gomock.Any(), gomock.Any()).Return(nil, errRepositoryFailed)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			userRepo := mock_user.NewMockIUserRepository(ctrl)
			threadRepo := mock_thread.NewMockIPendingThreadRepository(ctrl)
			timeGetter := mock_thread.NewMockITimeGetter(ctrl)
			logger := mock_thread.NewMockILogger(ctrl)
			logger.EXPECT().Error(gomock.Any(), errRepositoryFailed)
			tt.setupMocks(userRepo, threadRepo, timeGetter)

			userService := domainuser.NewUserService(userRepo)
			sut := thread.NewCreateThreadUsecase(threadRepo, userService, timeGetter, logger)

			_, err := sut.Exec(context.Background(), thread.CreateThreadCommand{IP: ip, Body: "hello", FileName: "sample.png"})

			require.Error(t, err)
		})
	}
}

func mustFilePath(t *testing.T) *domainthread.FilePath {
	t.Helper()
	fp, err := domainthread.NewFilePath("thread_images/", "sample.png")
	require.NoError(t, err)
	return fp
}
