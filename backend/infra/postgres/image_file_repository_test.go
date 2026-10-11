package postgres_test

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/nakanokota/bulletin-board-system/backend/domain/image_file"
	"github.com/nakanokota/bulletin-board-system/backend/infra/postgres"
)

func TestPostgresImageFileRepository_CreateImageFile_正常系(t *testing.T) {
	pool := newTestPool(t)

	tests := []struct {
		name      string
		imageFile *image_file.ImageFile
	}{
		{
			name:      "有効なImageFileを渡したとき、エラーなく保存され、filesテーブルにid・name・content_type・size_bytesが書き込まれること",
			imageFile: mustNewImageFile(t, "file-1", ".png", "image/png", 1024),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := context.Background()

			tx := beginTx(t, pool)

			sut := postgres.NewPostgresImageFileRepository(tx)
			err := sut.CreateImageFile(ctx, tt.imageFile)

			require.NoError(t, err)

			var storedName, storedContentType string
			var storedSizeBytes int64
			err = tx.QueryRow(ctx, "SELECT name, content_type, size_bytes FROM files WHERE id = $1", tt.imageFile.ID()).
				Scan(&storedName, &storedContentType, &storedSizeBytes)
			require.NoError(t, err)
			assert.Equal(t, tt.imageFile.Name().Value(), storedName)
			assert.Equal(t, tt.imageFile.ContentType(), storedContentType)
			assert.Equal(t, tt.imageFile.SizeBytes(), storedSizeBytes)
		})
	}
}

func mustNewImageFile(t *testing.T, id, extension, contentType string, sizeBytes int64) *image_file.ImageFile {
	t.Helper()
	f, err := image_file.NewImageFile(id, extension, contentType, sizeBytes)
	require.NoError(t, err)
	return f
}
