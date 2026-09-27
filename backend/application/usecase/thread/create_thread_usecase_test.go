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

// createThreadMocks bundles every mocked leaf dependency of CreateThreadUsecase. UserService and
// ThreadCreationService themselves are real; only what they are injected with is mocked.
type createThreadMocks struct {
	userRepo    *mock_user.MockIUserRepository
	fileChecker *mock_thread.MockIThreadFileChecker
	threadRepo  *mock_thread.MockIPendingThreadRepository
	timeGetter  *mock_thread.MockITimeGetter
	logger      *mock_thread.MockILogger
}

func newCreateThreadUsecase(t *testing.T, setupMocks func(createThreadMocks)) *thread.CreateThreadUsecase {
	t.Helper()
	ctrl := gomock.NewController(t)
	m := createThreadMocks{
		userRepo:    mock_user.NewMockIUserRepository(ctrl),
		fileChecker: mock_thread.NewMockIThreadFileChecker(ctrl),
		threadRepo:  mock_thread.NewMockIPendingThreadRepository(ctrl),
		timeGetter:  mock_thread.NewMockITimeGetter(ctrl),
		logger:      mock_thread.NewMockILogger(ctrl),
	}
	setupMocks(m)

	userService := domainuser.NewUserService(m.userRepo)
	threadCreationService := domainthread.NewThreadCreationService(m.fileChecker, m.threadRepo)
	return thread.NewCreateThreadUsecase(threadCreationService, userService, m.timeGetter, m.logger)
}

func TestCreateThreadUsecase_Exec_正常系(t *testing.T) {
	ip := "203.0.113.1"
	now := time.Date(2026, 9, 13, 15, 30, 0, 0, time.UTC)
	existingUser, _ := domainuser.NewUser("user-1", ip)
	newUser, _ := domainuser.NewUser("user-2", ip)
	createdThread, _ := domainthread.NewThread("thread-1", "hello", mustFilePath(t), now)

	tests := []struct {
		name         string
		cmd          thread.CreateThreadCommand
		setupMocks   func(createThreadMocks)
		wantThreadID string
	}{
		{
			name: "同一IP・同一日のUserが既に存在し、ファイルが存在かつ未使用のとき、作成されたThreadのIDがOutputのThreadIDに入ること",
			cmd:  thread.CreateThreadCommand{IP: ip, Body: "hello", FileID: "file-1"},
			setupMocks: func(m createThreadMocks) {
				m.timeGetter.EXPECT().Now(gomock.Any()).Return(now)
				m.userRepo.EXPECT().FindByIPAndDate(gomock.Any(), gomock.Any(), gomock.Any()).Return(existingUser, nil)
				m.fileChecker.EXPECT().ExistsFile(gomock.Any(), gomock.Any()).Return(true, nil)
				m.fileChecker.EXPECT().IsFileAttachedToThread(gomock.Any(), gomock.Any()).Return(false, nil)
				m.threadRepo.EXPECT().CreateThread(gomock.Any(), gomock.Any()).Return(createdThread, nil)
			},
			wantThreadID: "thread-1",
		},
		{
			name: "同一IP・同一日のUserが存在せず、ファイルが存在かつ未使用のとき、作成されたThreadのIDがOutputのThreadIDに入ること",
			cmd:  thread.CreateThreadCommand{IP: ip, Body: "hello", FileID: "file-1"},
			setupMocks: func(m createThreadMocks) {
				m.timeGetter.EXPECT().Now(gomock.Any()).Return(now)
				m.userRepo.EXPECT().FindByIPAndDate(gomock.Any(), gomock.Any(), gomock.Any()).Return(nil, nil)
				m.userRepo.EXPECT().CreateUser(gomock.Any(), gomock.Any()).Return(newUser, nil)
				m.fileChecker.EXPECT().ExistsFile(gomock.Any(), gomock.Any()).Return(true, nil)
				m.fileChecker.EXPECT().IsFileAttachedToThread(gomock.Any(), gomock.Any()).Return(false, nil)
				m.threadRepo.EXPECT().CreateThread(gomock.Any(), gomock.Any()).Return(createdThread, nil)
			},
			wantThreadID: "thread-1",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			sut := newCreateThreadUsecase(t, tt.setupMocks)

			got, err := sut.Exec(context.Background(), tt.cmd)

			require.NoError(t, err)
			assert.Equal(t, tt.wantThreadID, got.ThreadID)
		})
	}
}

func TestCreateThreadUsecase_Exec_依存に渡す値(t *testing.T) {
	ip := "203.0.113.1"
	now := time.Date(2026, 9, 13, 15, 30, 0, 0, time.UTC)
	date, _ := domainuser.NewUserCreateDate(time.Date(2026, 9, 13, 0, 0, 0, 0, time.UTC))
	existingUser, _ := domainuser.NewUser("user-1", ip)
	newUser, _ := domainuser.NewUser("user-2", ip)
	pendingThreadForExistingUser, _ := domainthread.NewPendingThread("user-1", "hello", "file-1")
	pendingThreadForNewUser, _ := domainthread.NewPendingThread("user-2", "hello", "file-1")
	createdThread, _ := domainthread.NewThread("thread-1", "hello", mustFilePath(t), now)

	tests := []struct {
		name       string
		cmd        thread.CreateThreadCommand
		setupMocks func(createThreadMocks)
	}{
		{
			name: "FindByIPAndDateに、CommandのIPと、ITimeGetterが返した時刻の日付が渡されること",
			cmd:  thread.CreateThreadCommand{IP: ip, Body: "hello", FileID: "file-1"},
			setupMocks: func(m createThreadMocks) {
				m.timeGetter.EXPECT().Now(gomock.Any()).Return(now)
				m.userRepo.EXPECT().FindByIPAndDate(gomock.Any(), ip, date).Return(existingUser, nil)
				m.fileChecker.EXPECT().ExistsFile(gomock.Any(), gomock.Any()).Return(true, nil)
				m.fileChecker.EXPECT().IsFileAttachedToThread(gomock.Any(), gomock.Any()).Return(false, nil)
				m.threadRepo.EXPECT().CreateThread(gomock.Any(), gomock.Any()).Return(createdThread, nil)
			},
		},
		{
			name: "同一IP・同一日のUserが既に存在するとき、CreateThreadに、そのUserのID・CommandのBody・CommandのFileIDを持つPendingThreadが渡されること",
			cmd:  thread.CreateThreadCommand{IP: ip, Body: "hello", FileID: "file-1"},
			setupMocks: func(m createThreadMocks) {
				m.timeGetter.EXPECT().Now(gomock.Any()).Return(now)
				m.userRepo.EXPECT().FindByIPAndDate(gomock.Any(), gomock.Any(), gomock.Any()).Return(existingUser, nil)
				m.fileChecker.EXPECT().ExistsFile(gomock.Any(), gomock.Any()).Return(true, nil)
				m.fileChecker.EXPECT().IsFileAttachedToThread(gomock.Any(), gomock.Any()).Return(false, nil)
				m.threadRepo.EXPECT().CreateThread(gomock.Any(), pendingThreadForExistingUser).Return(createdThread, nil)
			},
		},
		{
			name: "同一IP・同一日のUserが存在しないとき、CreateThreadに、新規作成されたUserのIDを持つPendingThreadが渡されること",
			cmd:  thread.CreateThreadCommand{IP: ip, Body: "hello", FileID: "file-1"},
			setupMocks: func(m createThreadMocks) {
				m.timeGetter.EXPECT().Now(gomock.Any()).Return(now)
				m.userRepo.EXPECT().FindByIPAndDate(gomock.Any(), gomock.Any(), gomock.Any()).Return(nil, nil)
				m.userRepo.EXPECT().CreateUser(gomock.Any(), gomock.Any()).Return(newUser, nil)
				m.fileChecker.EXPECT().ExistsFile(gomock.Any(), gomock.Any()).Return(true, nil)
				m.fileChecker.EXPECT().IsFileAttachedToThread(gomock.Any(), gomock.Any()).Return(false, nil)
				m.threadRepo.EXPECT().CreateThread(gomock.Any(), pendingThreadForNewUser).Return(createdThread, nil)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			sut := newCreateThreadUsecase(t, tt.setupMocks)

			_, err := sut.Exec(context.Background(), tt.cmd)

			require.NoError(t, err)
		})
	}
}

func TestCreateThreadUsecase_Exec_異常系(t *testing.T) {
	ip := "203.0.113.1"
	now := time.Date(2026, 9, 13, 15, 30, 0, 0, time.UTC)
	existingUser, _ := domainuser.NewUser("user-1", ip)
	errRepositoryFailed := errors.New("repository failed")

	tests := []struct {
		name       string
		cmd        thread.CreateThreadCommand
		setupMocks func(createThreadMocks)
		wantErr    error
	}{
		{
			name: "ファイルが存在しないとき、applicationのErrInvalidFileが返ること",
			cmd:  thread.CreateThreadCommand{IP: ip, Body: "hello", FileID: "file-1"},
			setupMocks: func(m createThreadMocks) {
				m.timeGetter.EXPECT().Now(gomock.Any()).Return(now)
				m.userRepo.EXPECT().FindByIPAndDate(gomock.Any(), gomock.Any(), gomock.Any()).Return(existingUser, nil)
				m.fileChecker.EXPECT().ExistsFile(gomock.Any(), gomock.Any()).Return(false, nil)
			},
			wantErr: thread.ErrInvalidFile,
		},
		{
			name: "ファイルが他スレッドで使用済みのとき、applicationのErrFileAlreadyUsedが返ること",
			cmd:  thread.CreateThreadCommand{IP: ip, Body: "hello", FileID: "file-1"},
			setupMocks: func(m createThreadMocks) {
				m.timeGetter.EXPECT().Now(gomock.Any()).Return(now)
				m.userRepo.EXPECT().FindByIPAndDate(gomock.Any(), gomock.Any(), gomock.Any()).Return(existingUser, nil)
				m.fileChecker.EXPECT().ExistsFile(gomock.Any(), gomock.Any()).Return(true, nil)
				m.fileChecker.EXPECT().IsFileAttachedToThread(gomock.Any(), gomock.Any()).Return(true, nil)
			},
			wantErr: thread.ErrFileAlreadyUsed,
		},
		{
			name: "Userの解決(FindByIPAndDate)がエラーを返したとき、そのエラーが返ること",
			cmd:  thread.CreateThreadCommand{IP: ip, Body: "hello", FileID: "file-1"},
			setupMocks: func(m createThreadMocks) {
				m.timeGetter.EXPECT().Now(gomock.Any()).Return(now)
				m.userRepo.EXPECT().FindByIPAndDate(gomock.Any(), gomock.Any(), gomock.Any()).Return(nil, errRepositoryFailed)
			},
			wantErr: errRepositoryFailed,
		},
		{
			name: "CreateThreadがErrFileNotFound・ErrFileAlreadyUsed以外のエラーを返したとき、そのエラーが返ること",
			cmd:  thread.CreateThreadCommand{IP: ip, Body: "hello", FileID: "file-1"},
			setupMocks: func(m createThreadMocks) {
				m.timeGetter.EXPECT().Now(gomock.Any()).Return(now)
				m.userRepo.EXPECT().FindByIPAndDate(gomock.Any(), gomock.Any(), gomock.Any()).Return(existingUser, nil)
				m.fileChecker.EXPECT().ExistsFile(gomock.Any(), gomock.Any()).Return(true, nil)
				m.fileChecker.EXPECT().IsFileAttachedToThread(gomock.Any(), gomock.Any()).Return(false, nil)
				m.threadRepo.EXPECT().CreateThread(gomock.Any(), gomock.Any()).Return(nil, errRepositoryFailed)
			},
			wantErr: errRepositoryFailed,
		},
		{
			name: "CommandのBodyが空文字のとき、applicationのErrInvalidBodyが返ること",
			cmd:  thread.CreateThreadCommand{IP: ip, Body: "", FileID: "file-1"},
			setupMocks: func(m createThreadMocks) {
				m.timeGetter.EXPECT().Now(gomock.Any()).Return(now)
				m.userRepo.EXPECT().FindByIPAndDate(gomock.Any(), gomock.Any(), gomock.Any()).Return(existingUser, nil)
			},
			wantErr: thread.ErrInvalidBody,
		},
		{
			name: "CommandのFileIDが空文字のとき、applicationのErrInvalidFileが返ること",
			cmd:  thread.CreateThreadCommand{IP: ip, Body: "hello", FileID: ""},
			setupMocks: func(m createThreadMocks) {
				m.timeGetter.EXPECT().Now(gomock.Any()).Return(now)
				m.userRepo.EXPECT().FindByIPAndDate(gomock.Any(), gomock.Any(), gomock.Any()).Return(existingUser, nil)
			},
			wantErr: thread.ErrInvalidFile,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			sut := newCreateThreadUsecase(t, func(m createThreadMocks) {
				m.logger.EXPECT().Error(gomock.Any(), gomock.Any()).AnyTimes()
				tt.setupMocks(m)
			})

			_, err := sut.Exec(context.Background(), tt.cmd)

			require.ErrorIs(t, err, tt.wantErr)
		})
	}
}

func TestCreateThreadUsecase_Exec_ログ出力(t *testing.T) {
	ip := "203.0.113.1"
	now := time.Date(2026, 9, 13, 15, 30, 0, 0, time.UTC)
	existingUser, _ := domainuser.NewUser("user-1", ip)
	errRepositoryFailed := errors.New("repository failed")

	tests := []struct {
		name       string
		cmd        thread.CreateThreadCommand
		setupMocks func(createThreadMocks)
	}{
		{
			name: "Userの解決(FindByIPAndDate)がエラーを返したとき、ILoggerにそのエラーが渡されること",
			cmd:  thread.CreateThreadCommand{IP: ip, Body: "hello", FileID: "file-1"},
			setupMocks: func(m createThreadMocks) {
				m.timeGetter.EXPECT().Now(gomock.Any()).Return(now)
				m.userRepo.EXPECT().FindByIPAndDate(gomock.Any(), gomock.Any(), gomock.Any()).Return(nil, errRepositoryFailed)
				m.logger.EXPECT().Error(gomock.Any(), errRepositoryFailed).Times(1)
			},
		},
		{
			name: "CreateThreadがErrFileNotFound・ErrFileAlreadyUsed以外のエラーを返したとき、ILoggerにそのエラーが渡されること",
			cmd:  thread.CreateThreadCommand{IP: ip, Body: "hello", FileID: "file-1"},
			setupMocks: func(m createThreadMocks) {
				m.timeGetter.EXPECT().Now(gomock.Any()).Return(now)
				m.userRepo.EXPECT().FindByIPAndDate(gomock.Any(), gomock.Any(), gomock.Any()).Return(existingUser, nil)
				m.fileChecker.EXPECT().ExistsFile(gomock.Any(), gomock.Any()).Return(true, nil)
				m.fileChecker.EXPECT().IsFileAttachedToThread(gomock.Any(), gomock.Any()).Return(false, nil)
				m.threadRepo.EXPECT().CreateThread(gomock.Any(), gomock.Any()).Return(nil, errRepositoryFailed)
				m.logger.EXPECT().Error(gomock.Any(), errRepositoryFailed).Times(1)
			},
		},
		{
			name: "ファイルが存在しないとき、ILoggerにdomainのErrFileNotFound(ラップ前の元のエラー)が渡されること",
			cmd:  thread.CreateThreadCommand{IP: ip, Body: "hello", FileID: "file-1"},
			setupMocks: func(m createThreadMocks) {
				m.timeGetter.EXPECT().Now(gomock.Any()).Return(now)
				m.userRepo.EXPECT().FindByIPAndDate(gomock.Any(), gomock.Any(), gomock.Any()).Return(existingUser, nil)
				m.fileChecker.EXPECT().ExistsFile(gomock.Any(), gomock.Any()).Return(false, nil)
				m.logger.EXPECT().Error(gomock.Any(), domainthread.ErrFileNotFound).Times(1)
			},
		},
		{
			name: "ファイルが他スレッドで使用済みのとき、ILoggerにdomainのErrFileAlreadyUsed(ラップ前の元のエラー)が渡されること",
			cmd:  thread.CreateThreadCommand{IP: ip, Body: "hello", FileID: "file-1"},
			setupMocks: func(m createThreadMocks) {
				m.timeGetter.EXPECT().Now(gomock.Any()).Return(now)
				m.userRepo.EXPECT().FindByIPAndDate(gomock.Any(), gomock.Any(), gomock.Any()).Return(existingUser, nil)
				m.fileChecker.EXPECT().ExistsFile(gomock.Any(), gomock.Any()).Return(true, nil)
				m.fileChecker.EXPECT().IsFileAttachedToThread(gomock.Any(), gomock.Any()).Return(true, nil)
				m.logger.EXPECT().Error(gomock.Any(), domainthread.ErrFileAlreadyUsed).Times(1)
			},
		},
		{
			name: "CommandのBodyが空文字のとき、ILoggerが呼ばれないこと",
			cmd:  thread.CreateThreadCommand{IP: ip, Body: "", FileID: "file-1"},
			setupMocks: func(m createThreadMocks) {
				m.timeGetter.EXPECT().Now(gomock.Any()).Return(now)
				m.userRepo.EXPECT().FindByIPAndDate(gomock.Any(), gomock.Any(), gomock.Any()).Return(existingUser, nil)
				m.logger.EXPECT().Error(gomock.Any(), gomock.Any()).Times(0)
			},
		},
		{
			name: "CommandのFileIDが空文字のとき、ILoggerが呼ばれないこと",
			cmd:  thread.CreateThreadCommand{IP: ip, Body: "hello", FileID: ""},
			setupMocks: func(m createThreadMocks) {
				m.timeGetter.EXPECT().Now(gomock.Any()).Return(now)
				m.userRepo.EXPECT().FindByIPAndDate(gomock.Any(), gomock.Any(), gomock.Any()).Return(existingUser, nil)
				m.logger.EXPECT().Error(gomock.Any(), gomock.Any()).Times(0)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			sut := newCreateThreadUsecase(t, tt.setupMocks)

			_, err := sut.Exec(context.Background(), tt.cmd)

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
