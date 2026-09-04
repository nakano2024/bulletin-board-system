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
			fileName:  "sample123",
			wantValue: "images/sample123",
		},
		{
			name:      "basePathが1文字+\"/\"のとき、正常に生成されること",
			basePath:  "a/",
			fileName:  "sample123",
			wantValue: "a/sample123",
		},
		{
			name:      "fileNameが1文字のとき、正常に生成されること",
			basePath:  "images/",
			fileName:  "a",
			wantValue: "images/a",
		},
		{
			name:      "fileNameが拡張子(ドット+半角英数字)を含むとき、正常に生成されること",
			basePath:  "images/",
			fileName:  "test.png",
			wantValue: "images/test.png",
		},
		{
			name:      "fileNameが半角英大文字・小文字・数字を含むとき、正常に生成されること",
			basePath:  "images/",
			fileName:  "Sample123",
			wantValue: "images/Sample123",
		},
		{
			name:      "fileNameにアンダースコア(スネークケース)が含まれるとき、正常に生成されること",
			basePath:  "images/",
			fileName:  "sample_file",
			wantValue: "images/sample_file",
		},
		{
			name:      "fileNameにハイフン(ケバブケース)が含まれるとき、正常に生成されること",
			basePath:  "images/",
			fileName:  "sample-file",
			wantValue: "images/sample-file",
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
			fileName: "sample123",
			wantErr:  thread.ErrFilePathBasePathInvalid,
		},
		{
			name:     "basePathが\"/\"のみ(半角英数字0文字)のとき、ErrFilePathBasePathInvalidが返ること",
			basePath: "/",
			fileName: "sample123",
			wantErr:  thread.ErrFilePathBasePathInvalid,
		},
		{
			name:     "basePathが\"/\"で終わっていないとき、ErrFilePathBasePathInvalidが返ること",
			basePath: "images",
			fileName: "sample123",
			wantErr:  thread.ErrFilePathBasePathInvalid,
		},
		{
			name:     "basePathに半角英数字・\"/\"以外の文字が含まれるとき、ErrFilePathBasePathInvalidが返ること",
			basePath: "images-1/",
			fileName: "sample123",
			wantErr:  thread.ErrFilePathBasePathInvalid,
		},
		{
			name:     "fileNameが空文字のとき、ErrFilePathFileNameInvalidが返ること",
			basePath: "images/",
			fileName: "",
			wantErr:  thread.ErrFilePathFileNameInvalid,
		},
		{
			name:     "fileNameに半角英数字・アンダースコア・ハイフン・拡張子(ドット)以外の文字が含まれるとき、ErrFilePathFileNameInvalidが返ること",
			basePath: "images/",
			fileName: "sample@1",
			wantErr:  thread.ErrFilePathFileNameInvalid,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := thread.NewFilePath(tt.basePath, tt.fileName)

			require.ErrorIs(t, err, tt.wantErr)
		})
	}
}
