package thread_test

import (
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/nakanokota/bulletin-board-system/backend/domain/image_file"
	"github.com/nakanokota/bulletin-board-system/backend/domain/thread"
)

func TestNewReply_正常系(t *testing.T) {
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
			name:      "id・userID・body・filePath・createdAtが全て正常な値のとき、ID()・UserID()・Body()・FilePath()・CreatedAt()が入力どおりの値のReplyが返ること",
			id:        "reply-1",
			userID:    "user-1",
			body:      "hello",
			filePath:  filePath,
			createdAt: fixedTime,
		},
		{
			name:      "filePathがnil(画像なし)のとき、FilePath()がnilのReplyが返ること",
			id:        "reply-1",
			userID:    "user-1",
			body:      "hello",
			filePath:  nil,
			createdAt: fixedTime,
		},
		{
			name:      "bodyが1文字のとき、Body()が入力どおりの値のReplyが返ること",
			id:        "reply-1",
			userID:    "user-1",
			body:      "a",
			filePath:  filePath,
			createdAt: fixedTime,
		},
		{
			name:      "bodyが1000文字のとき、Body()が入力どおりの値のReplyが返ること",
			id:        "reply-1",
			userID:    "user-1",
			body:      strings.Repeat("a", 1000),
			filePath:  filePath,
			createdAt: fixedTime,
		},
		{
			name:      "bodyがマルチバイト文字(「あ」)1000文字のとき、Body()が入力どおりの値のReplyが返ること",
			id:        "reply-1",
			userID:    "user-1",
			body:      strings.Repeat("あ", 1000),
			filePath:  filePath,
			createdAt: fixedTime,
		},
		{
			name:      "idが1文字のとき、ID()が入力どおりの値のReplyが返ること",
			id:        "r",
			userID:    "user-1",
			body:      "hello",
			filePath:  filePath,
			createdAt: fixedTime,
		},
		{
			name:      "userIDが1文字のとき、UserID()が入力どおりの値のReplyが返ること",
			id:        "reply-1",
			userID:    "u",
			body:      "hello",
			filePath:  filePath,
			createdAt: fixedTime,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := thread.NewReply(tt.id, tt.userID, tt.body, tt.filePath, tt.createdAt)

			require.NoError(t, err)
			assert.Equal(t, tt.id, got.ID())
			assert.Equal(t, tt.userID, got.UserID())
			assert.Equal(t, tt.body, got.Body())
			assert.Equal(t, tt.filePath, got.FilePath())
			assert.Equal(t, tt.createdAt, got.CreatedAt())
		})
	}
}

func TestNewReply_異常系(t *testing.T) {
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
		wantErr   error
	}{
		{
			name:      "idが空文字のとき、ErrReplyIDEmptyが返ること",
			id:        "",
			userID:    "user-1",
			body:      "hello",
			filePath:  filePath,
			createdAt: fixedTime,
			wantErr:   thread.ErrReplyIDEmpty,
		},
		{
			name:      "userIDが空文字のとき、ErrReplyUserIDEmptyが返ること",
			id:        "reply-1",
			userID:    "",
			body:      "hello",
			filePath:  filePath,
			createdAt: fixedTime,
			wantErr:   thread.ErrReplyUserIDEmpty,
		},
		{
			name:      "bodyが空文字のとき、ErrReplyBodyEmptyが返ること",
			id:        "reply-1",
			userID:    "user-1",
			body:      "",
			filePath:  filePath,
			createdAt: fixedTime,
			wantErr:   thread.ErrReplyBodyEmpty,
		},
		{
			name:      "bodyが1001文字のとき、ErrReplyBodyTooLongが返ること",
			id:        "reply-1",
			userID:    "user-1",
			body:      strings.Repeat("a", 1001),
			filePath:  filePath,
			createdAt: fixedTime,
			wantErr:   thread.ErrReplyBodyTooLong,
		},
		{
			name:      "createdAtがゼロ値のとき、ErrReplyCreatedAtZeroが返ること",
			id:        "reply-1",
			userID:    "user-1",
			body:      "hello",
			filePath:  filePath,
			createdAt: time.Time{},
			wantErr:   thread.ErrReplyCreatedAtZero,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := thread.NewReply(tt.id, tt.userID, tt.body, tt.filePath, tt.createdAt)

			require.ErrorIs(t, err, tt.wantErr)
		})
	}
}
