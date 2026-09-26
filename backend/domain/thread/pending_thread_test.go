package thread_test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/nakanokota/bulletin-board-system/backend/domain/thread"
)

func TestNewPendingThread_正常系(t *testing.T) {
	fileName, _ := thread.NewFileName("sample.png")

	tests := []struct {
		name     string
		userID   string
		body     string
		fileName *thread.FileName
	}{
		{
			name:     "userID・body・fileName(拡張子.png)が正常な値のとき、PendingThreadが生成されること",
			userID:   "user-1",
			body:     "hello",
			fileName: fileName,
		},
		{
			name:     "userIDとbodyがともに1文字のとき、PendingThreadが生成されること",
			userID:   "u",
			body:     "a",
			fileName: fileName,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := thread.NewPendingThread(tt.userID, tt.body, tt.fileName)

			require.NoError(t, err)
			assert.Equal(t, tt.userID, got.UserID())
			assert.Equal(t, tt.body, got.Body())
			assert.Equal(t, tt.fileName, got.FileName())
		})
	}
}

func TestNewPendingThread_異常系(t *testing.T) {
	pngFileName, _ := thread.NewFileName("sample.png")

	tests := []struct {
		name     string
		userID   string
		body     string
		fileName *thread.FileName
		wantErr  error
	}{
		{
			name:     "userIDが空文字のとき、ErrPendingThreadUserIDEmptyが返ること",
			userID:   "",
			body:     "hello",
			fileName: pngFileName,
			wantErr:  thread.ErrPendingThreadUserIDEmpty,
		},
		{
			name:     "bodyが空文字のとき、ErrPendingThreadBodyEmptyが返ること",
			userID:   "user-1",
			body:     "",
			fileName: pngFileName,
			wantErr:  thread.ErrPendingThreadBodyEmpty,
		},
		{
			name:     "fileNameがnilのとき、ErrPendingThreadFileNameMissingが返ること",
			userID:   "user-1",
			body:     "hello",
			fileName: nil,
			wantErr:  thread.ErrPendingThreadFileNameMissing,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := thread.NewPendingThread(tt.userID, tt.body, tt.fileName)

			require.ErrorIs(t, err, tt.wantErr)
		})
	}
}
