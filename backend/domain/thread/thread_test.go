package thread_test

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/nakanokota/bulletin-board-system/backend/domain/thread"
)

func TestNewThread_正常系(t *testing.T) {
	fixedTime := time.Date(2026, 9, 3, 12, 0, 0, 0, time.UTC)
	filePath, err := thread.NewFilePath("images/", "sample123")
	require.NoError(t, err)

	tests := []struct {
		name              string
		id                string
		body              string
		filePath          *thread.FilePath
		createdAt         time.Time
		wantFilePathValue string
	}{
		{
			name:              "id・body・createdAt・filePathが全て正常な値のとき、Threadが生成されること",
			id:                "thread-1",
			body:              "hello",
			filePath:          filePath,
			createdAt:         fixedTime,
			wantFilePathValue: "images/sample123",
		},
		{
			name:              "bodyが1文字のとき、Threadが生成されること",
			id:                "thread-1",
			body:              "a",
			filePath:          filePath,
			createdAt:         fixedTime,
			wantFilePathValue: "images/sample123",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := thread.NewThread(tt.id, tt.body, tt.filePath, tt.createdAt)

			require.NoError(t, err)
			assert.Equal(t, tt.id, got.ID())
			assert.Equal(t, tt.body, got.Body())
			assert.Equal(t, tt.filePath, got.FilePath())
			assert.Equal(t, tt.wantFilePathValue, got.FilePathValue())
			assert.Equal(t, tt.createdAt, got.CreatedAt())
		})
	}
}

func TestNewThread_異常系(t *testing.T) {
	fixedTime := time.Date(2026, 9, 3, 12, 0, 0, 0, time.UTC)

	tests := []struct {
		name      string
		id        string
		body      string
		createdAt time.Time
		wantErr   error
	}{
		{
			name:      "idが空文字のとき、ErrThreadIDEmptyが返ること",
			id:        "",
			body:      "hello",
			createdAt: fixedTime,
			wantErr:   thread.ErrThreadIDEmpty,
		},
		{
			name:      "bodyが空文字のとき、ErrThreadBodyEmptyが返ること",
			id:        "thread-1",
			body:      "",
			createdAt: fixedTime,
			wantErr:   thread.ErrThreadBodyEmpty,
		},
		{
			name:      "createdAtがゼロ値のとき、ErrThreadCreatedAtZeroが返ること",
			id:        "thread-1",
			body:      "hello",
			createdAt: time.Time{},
			wantErr:   thread.ErrThreadCreatedAtZero,
		},
		{
			name:      "filePathがnilのとき、ErrThreadFilePathMissingが返ること",
			id:        "thread-1",
			body:      "hello",
			createdAt: fixedTime,
			wantErr:   thread.ErrThreadFilePathMissing,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := thread.NewThread(tt.id, tt.body, nil, tt.createdAt)

			require.ErrorIs(t, err, tt.wantErr)
		})
	}
}
