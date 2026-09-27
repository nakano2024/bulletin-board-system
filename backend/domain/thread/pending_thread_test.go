package thread_test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/nakanokota/bulletin-board-system/backend/domain/thread"
)

func TestNewPendingThread_正常系(t *testing.T) {
	tests := []struct {
		name   string
		userID string
		body   string
		fileID string
	}{
		{
			name:   "userID・body・fileIDがいずれも正常な値のとき、それぞれの値を保持したPendingThreadが生成されること",
			userID: "user-1",
			body:   "hello",
			fileID: "file-1",
		},
		{
			name:   "userID・body・fileIDがいずれも1文字のとき、それぞれの値を保持したPendingThreadが生成されること",
			userID: "u",
			body:   "a",
			fileID: "f",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := thread.NewPendingThread(tt.userID, tt.body, tt.fileID)

			require.NoError(t, err)
			assert.Equal(t, tt.userID, got.UserID())
			assert.Equal(t, tt.body, got.Body())
			assert.Equal(t, tt.fileID, got.FileID())
		})
	}
}

func TestNewPendingThread_異常系(t *testing.T) {
	tests := []struct {
		name    string
		userID  string
		body    string
		fileID  string
		wantErr error
	}{
		{
			name:    "userIDが空文字のとき、ErrPendingThreadUserIDEmptyが返ること",
			userID:  "",
			body:    "hello",
			fileID:  "file-1",
			wantErr: thread.ErrPendingThreadUserIDEmpty,
		},
		{
			name:    "bodyが空文字のとき、ErrPendingThreadBodyEmptyが返ること",
			userID:  "user-1",
			body:    "",
			fileID:  "file-1",
			wantErr: thread.ErrPendingThreadBodyEmpty,
		},
		{
			name:    "fileIDが空文字のとき、ErrPendingThreadFileIDEmptyが返ること",
			userID:  "user-1",
			body:    "hello",
			fileID:  "",
			wantErr: thread.ErrPendingThreadFileIDEmpty,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := thread.NewPendingThread(tt.userID, tt.body, tt.fileID)

			require.ErrorIs(t, err, tt.wantErr)
		})
	}
}
