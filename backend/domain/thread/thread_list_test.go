package thread_test

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/nakanokota/bulletin-board-system/backend/domain/thread"
)

func TestNewThreadList(t *testing.T) {
	fixedTime := time.Date(2026, 9, 3, 12, 0, 0, 0, time.UTC)
	filePath, err := thread.NewFilePath("images/", "sample.png")
	require.NoError(t, err)
	th, err := thread.NewThread("thread-1", "hello", filePath, fixedTime)
	require.NoError(t, err)

	tests := []struct {
		name    string
		threads []*thread.Thread
	}{
		{
			name:    "Threadを1件含むスライスを渡したとき、Threads()で渡した要素がそのまま返ること",
			threads: []*thread.Thread{th},
		},
		{
			name:    "空スライスを渡したとき、Threads()が空スライスを返すこと",
			threads: []*thread.Thread{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := thread.NewThreadList(tt.threads)

			assert.Equal(t, tt.threads, got.Threads())
		})
	}
}
