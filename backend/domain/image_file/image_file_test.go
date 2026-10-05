package image_file_test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/nakanokota/bulletin-board-system/backend/domain/image_file"
)

const sampleID = "3f2a1b4c-5d6e-4f70-8a9b-0c1d2e3f4a5b"

func TestNewImageFile_正常系(t *testing.T) {
	tests := []struct {
		name            string
		id              string
		extension       string
		contentType     string
		sizeBytes       int64
		wantID          string
		wantName        string
		wantContentType string
		wantSizeBytes   int64
	}{
		{
			name:            "各値が正常(sizeBytesが代表値1024)のとき、ID()がid、Name().Value()が\"{id}.png\"、ContentType()が\"image/png\"、SizeBytes()が1024であること",
			id:              sampleID,
			extension:       ".png",
			contentType:     "image/png",
			sizeBytes:       1024,
			wantID:          sampleID,
			wantName:        sampleID + ".png",
			wantContentType: "image/png",
			wantSizeBytes:   1024,
		},
		{
			name:            "sizeBytesが1(境界値)のとき、SizeBytes()が1のImageFileが生成されること",
			id:              sampleID,
			extension:       ".png",
			contentType:     "image/png",
			sizeBytes:       1,
			wantID:          sampleID,
			wantName:        sampleID + ".png",
			wantContentType: "image/png",
			wantSizeBytes:   1,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := image_file.NewImageFile(tt.id, tt.extension, tt.contentType, tt.sizeBytes)

			require.NoError(t, err)
			assert.Equal(t, tt.wantID, got.ID())
			assert.Equal(t, tt.wantName, got.Name().Value())
			assert.Equal(t, tt.wantContentType, got.ContentType())
			assert.Equal(t, tt.wantSizeBytes, got.SizeBytes())
		})
	}
}

func TestNewImageFile_異常系(t *testing.T) {
	tests := []struct {
		name        string
		id          string
		extension   string
		contentType string
		sizeBytes   int64
		wantErr     error
	}{
		{
			name:        "idが空文字のとき、ErrImageFileIDEmptyが返ること",
			id:          "",
			extension:   ".png",
			contentType: "image/png",
			sizeBytes:   1024,
			wantErr:     image_file.ErrImageFileIDEmpty,
		},
		{
			name:        "idにファイル名として使えない文字(\"@\")が含まれるとき、ErrFileNameInvalidが返ること",
			id:          "abc@1",
			extension:   ".png",
			contentType: "image/png",
			sizeBytes:   1024,
			wantErr:     image_file.ErrFileNameInvalid,
		},
		{
			name:        "extensionが許可された画像形式でない(\".txt\")とき、ErrFileNameInvalidが返ること",
			id:          sampleID,
			extension:   ".txt",
			contentType: "image/png",
			sizeBytes:   1024,
			wantErr:     image_file.ErrFileNameInvalid,
		},
		{
			name:        "extensionが空文字のとき、ErrFileNameInvalidが返ること",
			id:          sampleID,
			extension:   "",
			contentType: "image/png",
			sizeBytes:   1024,
			wantErr:     image_file.ErrFileNameInvalid,
		},
		{
			name:        "contentTypeが空文字のとき、ErrImageFileContentTypeEmptyが返ること",
			id:          sampleID,
			extension:   ".png",
			contentType: "",
			sizeBytes:   1024,
			wantErr:     image_file.ErrImageFileContentTypeEmpty,
		},
		{
			name:        "sizeBytesが0(境界値)のとき、ErrImageFileSizeBytesNotPositiveが返ること",
			id:          sampleID,
			extension:   ".png",
			contentType: "image/png",
			sizeBytes:   0,
			wantErr:     image_file.ErrImageFileSizeBytesNotPositive,
		},
		{
			name:        "sizeBytesが負数(-1、代表値)のとき、ErrImageFileSizeBytesNotPositiveが返ること",
			id:          sampleID,
			extension:   ".png",
			contentType: "image/png",
			sizeBytes:   -1,
			wantErr:     image_file.ErrImageFileSizeBytesNotPositive,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := image_file.NewImageFile(tt.id, tt.extension, tt.contentType, tt.sizeBytes)

			require.ErrorIs(t, err, tt.wantErr)
		})
	}
}
