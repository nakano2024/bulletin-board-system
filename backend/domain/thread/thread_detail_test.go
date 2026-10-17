package thread_test

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/nakanokota/bulletin-board-system/backend/domain/image_file"
	"github.com/nakanokota/bulletin-board-system/backend/domain/thread"
)

func TestNewThreadDetail_正常系(t *testing.T) {
	fixedTime := time.Date(2026, 10, 17, 12, 0, 0, 0, time.UTC)
	filePath, err := image_file.NewFilePath("images/", "sample.png")
	require.NoError(t, err)

	tests := []struct {
		name      string
		id        string
		userID    string
		body      string
		filePath  *image_file.FilePath
		createdAt time.Time
	}{
		{
			name:      "id・userID・body・filePath・createdAtが全て正常な値のとき、ID()・UserID()・Body()・FilePath()・CreatedAt()が入力どおりの値のThreadDetailが返ること",
			id:        "thread-1",
			userID:    "user-1",
			body:      "hello",
			filePath:  filePath,
			createdAt: fixedTime,
		},
		{
			name:      "bodyが1文字のとき、Body()が入力どおりの値のThreadDetailが返ること",
			id:        "thread-1",
			userID:    "user-1",
			body:      "a",
			filePath:  filePath,
			createdAt: fixedTime,
		},
		{
			name:      "userIDが1文字のとき、UserID()が入力どおりの値のThreadDetailが返ること",
			id:        "thread-1",
			userID:    "u",
			body:      "hello",
			filePath:  filePath,
			createdAt: fixedTime,
		},
		{
			name:      "idが1文字のとき、ID()が入力どおりの値のThreadDetailが返ること",
			id:        "t",
			userID:    "user-1",
			body:      "hello",
			filePath:  filePath,
			createdAt: fixedTime,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := thread.NewThreadDetail(tt.id, tt.userID, tt.body, tt.filePath, tt.createdAt, true, []*thread.Reply{})

			require.NoError(t, err)
			assert.Equal(t, tt.id, got.ID())
			assert.Equal(t, tt.userID, got.UserID())
			assert.Equal(t, tt.body, got.Body())
			assert.Equal(t, tt.filePath, got.FilePath())
			assert.Equal(t, tt.createdAt, got.CreatedAt())
		})
	}
}

func TestNewThreadDetail_異常系(t *testing.T) {
	fixedTime := time.Date(2026, 10, 17, 12, 0, 0, 0, time.UTC)
	filePath, err := image_file.NewFilePath("images/", "sample.png")
	require.NoError(t, err)

	tests := []struct {
		name      string
		id        string
		userID    string
		body      string
		filePath  *image_file.FilePath
		createdAt time.Time
		replies   []*thread.Reply
		wantErr   error
	}{
		{
			name:      "idが空文字のとき、ErrThreadDetailIDEmptyが返ること",
			id:        "",
			userID:    "user-1",
			body:      "hello",
			filePath:  filePath,
			createdAt: fixedTime,
			replies:   []*thread.Reply{},
			wantErr:   thread.ErrThreadDetailIDEmpty,
		},
		{
			name:      "userIDが空文字のとき、ErrThreadDetailUserIDEmptyが返ること",
			id:        "thread-1",
			userID:    "",
			body:      "hello",
			filePath:  filePath,
			createdAt: fixedTime,
			replies:   []*thread.Reply{},
			wantErr:   thread.ErrThreadDetailUserIDEmpty,
		},
		{
			name:      "bodyが空文字のとき、ErrThreadDetailBodyEmptyが返ること",
			id:        "thread-1",
			userID:    "user-1",
			body:      "",
			filePath:  filePath,
			createdAt: fixedTime,
			replies:   []*thread.Reply{},
			wantErr:   thread.ErrThreadDetailBodyEmpty,
		},
		{
			name:      "createdAtがゼロ値のとき、ErrThreadDetailCreatedAtZeroが返ること",
			id:        "thread-1",
			userID:    "user-1",
			body:      "hello",
			filePath:  filePath,
			createdAt: time.Time{},
			replies:   []*thread.Reply{},
			wantErr:   thread.ErrThreadDetailCreatedAtZero,
		},
		{
			name:      "filePathがnilのとき、ErrThreadDetailFilePathMissingが返ること",
			id:        "thread-1",
			userID:    "user-1",
			body:      "hello",
			filePath:  nil,
			createdAt: fixedTime,
			replies:   []*thread.Reply{},
			wantErr:   thread.ErrThreadDetailFilePathMissing,
		},
		{
			name:      "repliesがnilのとき、ErrThreadDetailRepliesNilが返ること",
			id:        "thread-1",
			userID:    "user-1",
			body:      "hello",
			filePath:  filePath,
			createdAt: fixedTime,
			replies:   nil,
			wantErr:   thread.ErrThreadDetailRepliesNil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := thread.NewThreadDetail(tt.id, tt.userID, tt.body, tt.filePath, tt.createdAt, true, tt.replies)

			require.ErrorIs(t, err, tt.wantErr)
		})
	}
}

func TestThreadDetail_IsDead(t *testing.T) {
	fixedTime := time.Date(2026, 10, 17, 12, 0, 0, 0, time.UTC)
	filePath, err := image_file.NewFilePath("images/", "sample.png")
	require.NoError(t, err)

	tests := []struct {
		name    string
		isAlive bool
		want    bool
	}{
		{
			name:    "isAliveがtrue(dat落ちしていない)のとき、IsDead()がfalseを返すこと",
			isAlive: true,
			want:    false,
		},
		{
			name:    "isAliveがfalse(dat落ち済み)のとき、IsDead()がtrueを返すこと",
			isAlive: false,
			want:    true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			sut, err := thread.NewThreadDetail("thread-1", "user-1", "hello", filePath, fixedTime, tt.isAlive, []*thread.Reply{})
			require.NoError(t, err)

			assert.Equal(t, tt.want, sut.IsDead())
		})
	}
}

func TestThreadDetail_Replies(t *testing.T) {
	fixedTime := time.Date(2026, 10, 17, 12, 0, 0, 0, time.UTC)
	filePath, err := image_file.NewFilePath("images/", "sample.png")
	require.NoError(t, err)
	newerReply, err := thread.NewReply("reply-2", "user-2", "second", nil, fixedTime.Add(time.Minute))
	require.NoError(t, err)
	olderReply, err := thread.NewReply("reply-1", "user-1", "first", filePath, fixedTime)
	require.NoError(t, err)

	tests := []struct {
		name    string
		replies []*thread.Reply
		want    []*thread.Reply
	}{
		{
			name:    "repliesに2件のReplyを渡したとき、Replies()が渡した2件を渡した順のまま返すこと",
			replies: []*thread.Reply{newerReply, olderReply},
			want:    []*thread.Reply{newerReply, olderReply},
		},
		{
			name:    "repliesに空配列を渡したとき、Replies()が空配列を返すこと",
			replies: []*thread.Reply{},
			want:    []*thread.Reply{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			sut, err := thread.NewThreadDetail("thread-1", "user-1", "hello", filePath, fixedTime, true, tt.replies)
			require.NoError(t, err)

			assert.Equal(t, tt.want, sut.Replies())
		})
	}
}
