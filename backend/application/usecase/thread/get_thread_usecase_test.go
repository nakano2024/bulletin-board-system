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
	"github.com/nakanokota/bulletin-board-system/backend/domain/image_file"
	domainthread "github.com/nakanokota/bulletin-board-system/backend/domain/thread"
)

func TestGetThreadUsecase_Exec_正常系(t *testing.T) {
	fixedTime := time.Date(2026, 10, 17, 12, 0, 0, 0, time.UTC)
	threadFilePath, _ := image_file.NewFilePath("images/", "thread.png")
	replyFilePath, _ := image_file.NewFilePath("images/", "reply.png")

	newerReply, _ := domainthread.NewReply("reply-2", "user-3", "second", nil, fixedTime.Add(2*time.Minute))
	olderReply, _ := domainthread.NewReply("reply-1", "user-2", "first", replyFilePath, fixedTime.Add(time.Minute))

	threadWithoutReplies, _ := domainthread.NewThreadDetail("thread-1", "user-1", "hello", threadFilePath, fixedTime, true, []*domainthread.Reply{})
	threadWithTwoReplies, _ := domainthread.NewThreadDetail("thread-1", "user-1", "hello", threadFilePath, fixedTime, true, []*domainthread.Reply{newerReply, olderReply})
	threadWithImageReply, _ := domainthread.NewThreadDetail("thread-1", "user-1", "hello", threadFilePath, fixedTime, true, []*domainthread.Reply{olderReply})
	threadWithNoImageReply, _ := domainthread.NewThreadDetail("thread-1", "user-1", "hello", threadFilePath, fixedTime, true, []*domainthread.Reply{newerReply})

	wantThread := func(replies []thread.GetThreadOutputReply) *thread.GetThreadOutput {
		return &thread.GetThreadOutput{
			Thread: thread.GetThreadOutputThread{
				ID:        "thread-1",
				UserID:    "user-1",
				Body:      "hello",
				ImagePath: "images/thread.png",
				CreatedAt: fixedTime,
				Replies:   replies,
			},
		}
	}
	wantNewerReply := thread.GetThreadOutputReply{ID: "reply-2", UserID: "user-3", Body: "second", ImagePath: "", CreatedAt: fixedTime.Add(2 * time.Minute)}
	wantOlderReply := thread.GetThreadOutputReply{ID: "reply-1", UserID: "user-2", Body: "first", ImagePath: "images/reply.png", CreatedAt: fixedTime.Add(time.Minute)}

	tests := []struct {
		name      string
		input     thread.GetThreadCommand
		setupMock func(*mock_thread.MockIThreadDetailFetcher)
		want      *thread.GetThreadOutput
	}{
		{
			name:  "レスが0件のスレッドのとき、Output.ThreadのID/UserID/Body/ImagePath/CreatedAtがスレッドの値に変換され、Output.Thread.Repliesが空スライスであること",
			input: thread.GetThreadCommand{ThreadID: "thread-1"},
			setupMock: func(m *mock_thread.MockIThreadDetailFetcher) {
				m.EXPECT().FetchThreadDetail(gomock.Any(), gomock.Any()).Return(threadWithoutReplies, nil)
			},
			want: wantThread([]thread.GetThreadOutputReply{}),
		},
		{
			name:  "レスが2件のスレッドのとき、Output.Thread.Repliesの件数・順序が取得したレスと一致し、各レスのID/UserID/Body/CreatedAtが変換されていること",
			input: thread.GetThreadCommand{ThreadID: "thread-1"},
			setupMock: func(m *mock_thread.MockIThreadDetailFetcher) {
				m.EXPECT().FetchThreadDetail(gomock.Any(), gomock.Any()).Return(threadWithTwoReplies, nil)
			},
			want: wantThread([]thread.GetThreadOutputReply{wantNewerReply, wantOlderReply}),
		},
		{
			name:  "画像ありのレスのとき、そのレスのImagePathがレスの画像パスの値であること",
			input: thread.GetThreadCommand{ThreadID: "thread-1"},
			setupMock: func(m *mock_thread.MockIThreadDetailFetcher) {
				m.EXPECT().FetchThreadDetail(gomock.Any(), gomock.Any()).Return(threadWithImageReply, nil)
			},
			want: wantThread([]thread.GetThreadOutputReply{wantOlderReply}),
		},
		{
			name:  "画像なしのレスのとき、そのレスのImagePathが空文字であること",
			input: thread.GetThreadCommand{ThreadID: "thread-1"},
			setupMock: func(m *mock_thread.MockIThreadDetailFetcher) {
				m.EXPECT().FetchThreadDetail(gomock.Any(), gomock.Any()).Return(threadWithNoImageReply, nil)
			},
			want: wantThread([]thread.GetThreadOutputReply{wantNewerReply}),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			threadDetailFetcher := mock_thread.NewMockIThreadDetailFetcher(ctrl)
			tt.setupMock(threadDetailFetcher)
			logger := mock_thread.NewMockILogger(ctrl)

			sut := thread.NewGetThreadUsecase(threadDetailFetcher, logger)
			got, err := sut.Exec(context.Background(), tt.input)

			require.NoError(t, err)
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestGetThreadUsecase_Exec_異常系(t *testing.T) {
	fixedTime := time.Date(2026, 10, 17, 12, 0, 0, 0, time.UTC)
	threadFilePath, _ := image_file.NewFilePath("images/", "thread.png")
	deadThread, _ := domainthread.NewThreadDetail("thread-1", "user-1", "hello", threadFilePath, fixedTime, false, []*domainthread.Reply{})
	errFetcherFailed := errors.New("fetcher failed")

	tests := []struct {
		name      string
		input     thread.GetThreadCommand
		setupMock func(*mock_thread.MockIThreadDetailFetcher)
		wantErr   error
	}{
		{
			name:  "FetchThreadDetailがドメインのErrThreadNotFoundを返すとき、applicationのErrThreadNotFoundが返ること",
			input: thread.GetThreadCommand{ThreadID: "thread-1"},
			setupMock: func(m *mock_thread.MockIThreadDetailFetcher) {
				m.EXPECT().FetchThreadDetail(gomock.Any(), gomock.Any()).Return(nil, domainthread.ErrThreadNotFound)
			},
			wantErr: thread.ErrThreadNotFound,
		},
		{
			name:  "取得したスレッドがdat落ち済みのとき、applicationのErrThreadDeadが返ること",
			input: thread.GetThreadCommand{ThreadID: "thread-1"},
			setupMock: func(m *mock_thread.MockIThreadDetailFetcher) {
				m.EXPECT().FetchThreadDetail(gomock.Any(), gomock.Any()).Return(deadThread, nil)
			},
			wantErr: thread.ErrThreadDead,
		},
		{
			name:  "FetchThreadDetailがその他のエラーを返すとき、そのエラーが返ること",
			input: thread.GetThreadCommand{ThreadID: "thread-1"},
			setupMock: func(m *mock_thread.MockIThreadDetailFetcher) {
				m.EXPECT().FetchThreadDetail(gomock.Any(), gomock.Any()).Return(nil, errFetcherFailed)
			},
			wantErr: errFetcherFailed,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			threadDetailFetcher := mock_thread.NewMockIThreadDetailFetcher(ctrl)
			tt.setupMock(threadDetailFetcher)
			logger := mock_thread.NewMockILogger(ctrl)
			logger.EXPECT().Error(gomock.Any(), gomock.Any()).AnyTimes()

			sut := thread.NewGetThreadUsecase(threadDetailFetcher, logger)
			_, err := sut.Exec(context.Background(), tt.input)

			require.ErrorIs(t, err, tt.wantErr)
		})
	}
}

func TestGetThreadUsecase_Exec_ログ出力(t *testing.T) {
	fixedTime := time.Date(2026, 10, 17, 12, 0, 0, 0, time.UTC)
	threadFilePath, _ := image_file.NewFilePath("images/", "thread.png")
	deadThread, _ := domainthread.NewThreadDetail("thread-1", "user-1", "hello", threadFilePath, fixedTime, false, []*domainthread.Reply{})
	errFetcherFailed := errors.New("fetcher failed")

	tests := []struct {
		name        string
		input       thread.GetThreadCommand
		setupMock   func(*mock_thread.MockIThreadDetailFetcher)
		setupLogger func(*mock_thread.MockILogger)
	}{
		{
			name:  "FetchThreadDetailがその他のエラーを返すとき、ILoggerにそのエラーがそのまま渡されること",
			input: thread.GetThreadCommand{ThreadID: "thread-1"},
			setupMock: func(m *mock_thread.MockIThreadDetailFetcher) {
				m.EXPECT().FetchThreadDetail(gomock.Any(), gomock.Any()).Return(nil, errFetcherFailed)
			},
			setupLogger: func(m *mock_thread.MockILogger) {
				m.EXPECT().Error(gomock.Any(), errFetcherFailed)
			},
		},
		{
			name:  "FetchThreadDetailがドメインのErrThreadNotFoundを返すとき、ILoggerが呼ばれないこと",
			input: thread.GetThreadCommand{ThreadID: "thread-1"},
			setupMock: func(m *mock_thread.MockIThreadDetailFetcher) {
				m.EXPECT().FetchThreadDetail(gomock.Any(), gomock.Any()).Return(nil, domainthread.ErrThreadNotFound)
			},
			setupLogger: func(m *mock_thread.MockILogger) {
				m.EXPECT().Error(gomock.Any(), gomock.Any()).Times(0)
			},
		},
		{
			name:  "取得したスレッドがdat落ち済みのとき、ILoggerが呼ばれないこと",
			input: thread.GetThreadCommand{ThreadID: "thread-1"},
			setupMock: func(m *mock_thread.MockIThreadDetailFetcher) {
				m.EXPECT().FetchThreadDetail(gomock.Any(), gomock.Any()).Return(deadThread, nil)
			},
			setupLogger: func(m *mock_thread.MockILogger) {
				m.EXPECT().Error(gomock.Any(), gomock.Any()).Times(0)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			threadDetailFetcher := mock_thread.NewMockIThreadDetailFetcher(ctrl)
			tt.setupMock(threadDetailFetcher)
			logger := mock_thread.NewMockILogger(ctrl)
			tt.setupLogger(logger)

			sut := thread.NewGetThreadUsecase(threadDetailFetcher, logger)
			_, err := sut.Exec(context.Background(), tt.input)

			require.Error(t, err)
		})
	}
}

func TestGetThreadUsecase_Exec_依存への入力(t *testing.T) {
	fixedTime := time.Date(2026, 10, 17, 12, 0, 0, 0, time.UTC)
	threadFilePath, _ := image_file.NewFilePath("images/", "thread.png")
	aliveThread, _ := domainthread.NewThreadDetail("thread-1", "user-1", "hello", threadFilePath, fixedTime, true, []*domainthread.Reply{})

	tests := []struct {
		name      string
		input     thread.GetThreadCommand
		setupMock func(*mock_thread.MockIThreadDetailFetcher)
	}{
		{
			name:  "CommandのThreadIDが、FetchThreadDetailのidにそのまま渡されること",
			input: thread.GetThreadCommand{ThreadID: "thread-1"},
			setupMock: func(m *mock_thread.MockIThreadDetailFetcher) {
				m.EXPECT().FetchThreadDetail(gomock.Any(), "thread-1").Return(aliveThread, nil)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			threadDetailFetcher := mock_thread.NewMockIThreadDetailFetcher(ctrl)
			tt.setupMock(threadDetailFetcher)
			logger := mock_thread.NewMockILogger(ctrl)

			sut := thread.NewGetThreadUsecase(threadDetailFetcher, logger)
			_, err := sut.Exec(context.Background(), tt.input)

			require.NoError(t, err)
		})
	}
}
