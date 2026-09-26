package thread_test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/nakanokota/bulletin-board-system/backend/domain/thread"
)

func TestNewFilePath_正常系(t *testing.T) {
	tests := []struct {
		name      string
		basePath  string
		fileName  string
		wantValue string
	}{
		{
			name:      "basePathとfileNameがともに正常な値のとき、Value()が結合された文字列であること",
			basePath:  "images/",
			fileName:  "sample.png",
			wantValue: "images/sample.png",
		},
		{
			name:      "basePathが1文字+\"/\"のとき、正常に生成されること",
			basePath:  "a/",
			fileName:  "sample.png",
			wantValue: "a/sample.png",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := thread.NewFilePath(tt.basePath, tt.fileName)

			require.NoError(t, err)
			assert.Equal(t, tt.wantValue, got.Value())
		})
	}
}

func TestNewFilePath_異常系(t *testing.T) {
	tests := []struct {
		name     string
		basePath string
		fileName string
		wantErr  error
	}{
		{
			name:     "basePathが空文字のとき、ErrFilePathBasePathInvalidが返ること",
			basePath: "",
			fileName: "sample.png",
			wantErr:  thread.ErrFilePathBasePathInvalid,
		},
		{
			name:     "basePathが\"/\"のみ(半角英数字0文字)のとき、ErrFilePathBasePathInvalidが返ること",
			basePath: "/",
			fileName: "sample.png",
			wantErr:  thread.ErrFilePathBasePathInvalid,
		},
		{
			name:     "basePathが\"/\"で終わっていないとき、ErrFilePathBasePathInvalidが返ること",
			basePath: "images",
			fileName: "sample.png",
			wantErr:  thread.ErrFilePathBasePathInvalid,
		},
		{
			name:     "basePathに半角英数字・\"/\"以外の文字が含まれるとき、ErrFilePathBasePathInvalidが返ること",
			basePath: "images-1/",
			fileName: "sample.png",
			wantErr:  thread.ErrFilePathBasePathInvalid,
		},
		{
			name:     "fileNameの形式が不正なとき、ErrFileNameInvalidが返ること(FileNameへ委譲されていること)",
			basePath: "images/",
			fileName: "sample@1.png",
			wantErr:  thread.ErrFileNameInvalid,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := thread.NewFilePath(tt.basePath, tt.fileName)

			require.ErrorIs(t, err, tt.wantErr)
		})
	}
}
