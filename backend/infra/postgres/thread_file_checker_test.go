package postgres_test

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/nakanokota/bulletin-board-system/backend/infra/postgres"
)

func TestPostgresThreadFileChecker_ExistsFile_正常系(t *testing.T) {
	pool := newTestPool(t)

	tests := []struct {
		name      string
		seedFiles []seedFile
		fileID    string
		want      bool
	}{
		{
			name:      "filesに登録済みのfile_idを渡したとき、trueが返ること",
			seedFiles: []seedFile{{id: "file-1", name: "sample.png"}},
			fileID:    "file-1",
			want:      true,
		},
		{
			name:      "filesに登録されていないfile_idを渡したとき、falseが返ること",
			seedFiles: []seedFile{{id: "file-1", name: "sample.png"}},
			fileID:    "file-missing",
			want:      false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := context.Background()
			tx := beginTx(t, pool)
			seedFilesAndThreads(t, tx, tt.seedFiles, nil)

			sut := postgres.NewPostgresThreadFileChecker(tx)
			got, err := sut.ExistsFile(ctx, tt.fileID)

			require.NoError(t, err)
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestPostgresThreadFileChecker_IsFileAttachedToThread_正常系(t *testing.T) {
	pool := newTestPool(t)
	fixedTime := time.Date(2026, 9, 3, 12, 0, 0, 0, time.UTC)

	tests := []struct {
		name        string
		seedFiles   []seedFile
		seedThreads []seedThread
		fileID      string
		want        bool
	}{
		{
			name:        "is_aliveがtrueのスレッドが参照しているfile_idを渡したとき、trueが返ること",
			seedFiles:   []seedFile{{id: "file-1", name: "sample.png"}},
			seedThreads: []seedThread{{id: "thread-1", body: "hello", fileID: "file-1", isAlive: true, createdAt: fixedTime}},
			fileID:      "file-1",
			want:        true,
		},
		{
			name:        "is_aliveがfalse(dat落ち)のスレッドだけが参照しているfile_idを渡したとき、trueが返ること",
			seedFiles:   []seedFile{{id: "file-1", name: "sample.png"}},
			seedThreads: []seedThread{{id: "thread-1", body: "hello", fileID: "file-1", isAlive: false, createdAt: fixedTime}},
			fileID:      "file-1",
			want:        true,
		},
		{
			name:        "filesには登録済みだが、どのスレッドにも参照されていないfile_idを渡したとき、falseが返ること",
			seedFiles:   []seedFile{{id: "file-1", name: "sample.png"}},
			seedThreads: nil,
			fileID:      "file-1",
			want:        false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := context.Background()
			tx := beginTx(t, pool)
			seedFilesAndThreads(t, tx, tt.seedFiles, tt.seedThreads)

			sut := postgres.NewPostgresThreadFileChecker(tx)
			got, err := sut.IsFileAttachedToThread(ctx, tt.fileID)

			require.NoError(t, err)
			assert.Equal(t, tt.want, got)
		})
	}
}
