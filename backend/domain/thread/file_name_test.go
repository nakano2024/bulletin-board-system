package thread_test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/nakanokota/bulletin-board-system/backend/domain/thread"
)

func TestNewFileName_正常系(t *testing.T) {
	tests := []struct {
		name      string
		fileName  string
		wantValue string
	}{
		{
			name:      "拡張子がpngのとき、正常に生成されること",
			fileName:  "sample.png",
			wantValue: "sample.png",
		},
		{
			name:      "拡張子がjpgのとき、正常に生成されること",
			fileName:  "sample.jpg",
			wantValue: "sample.jpg",
		},
		{
			name:      "拡張子がjpegのとき、正常に生成されること",
			fileName:  "sample.jpeg",
			wantValue: "sample.jpeg",
		},
		{
			name:      "拡張子がgifのとき、正常に生成されること",
			fileName:  "sample.gif",
			wantValue: "sample.gif",
		},
		{
			name:      "拡張子がwebpのとき、正常に生成されること",
			fileName:  "sample.webp",
			wantValue: "sample.webp",
		},
		{
			name:      "拡張子がsvgのとき、正常に生成されること",
			fileName:  "sample.svg",
			wantValue: "sample.svg",
		},
		{
			name:      "拡張子がbmpのとき、正常に生成されること",
			fileName:  "sample.bmp",
			wantValue: "sample.bmp",
		},
		{
			name:      "拡張子がheifのとき、正常に生成されること",
			fileName:  "sample.heif",
			wantValue: "sample.heif",
		},
		{
			name:      "拡張子がheicのとき、正常に生成されること",
			fileName:  "sample.heic",
			wantValue: "sample.heic",
		},
		{
			name:      "拡張子が大文字(.PNG)のとき、正常に生成されること",
			fileName:  "sample.PNG",
			wantValue: "sample.PNG",
		},
		{
			name:      "ファイル名部分が1文字のとき、正常に生成されること",
			fileName:  "a.png",
			wantValue: "a.png",
		},
		{
			name:      "ファイル名部分が半角英大文字・小文字・数字を含むとき、正常に生成されること",
			fileName:  "Sample123.png",
			wantValue: "Sample123.png",
		},
		{
			name:      "ファイル名部分にアンダースコア(スネークケース)が含まれるとき、正常に生成されること",
			fileName:  "sample_file.png",
			wantValue: "sample_file.png",
		},
		{
			name:      "ファイル名部分にハイフン(ケバブケース)が含まれるとき、正常に生成されること",
			fileName:  "sample-file.png",
			wantValue: "sample-file.png",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := thread.NewFileName(tt.fileName)

			require.NoError(t, err)
			assert.Equal(t, tt.wantValue, got.Value())
		})
	}
}

func TestNewFileName_異常系(t *testing.T) {
	tests := []struct {
		name     string
		fileName string
		wantErr  error
	}{
		{
			name:     "fileNameが空文字のとき、ErrFileNameInvalidが返ること",
			fileName: "",
			wantErr:  thread.ErrFileNameInvalid,
		},
		{
			name:     "fileNameに拡張子がないとき、ErrFileNameInvalidが返ること",
			fileName: "sample",
			wantErr:  thread.ErrFileNameInvalid,
		},
		{
			name:     "fileNameの拡張子が許可された画像形式でないとき、ErrFileNameInvalidが返ること",
			fileName: "sample.txt",
			wantErr:  thread.ErrFileNameInvalid,
		},
		{
			name:     "ファイル名部分に半角英数字・アンダースコア・ハイフン以外の文字が含まれるとき、ErrFileNameInvalidが返ること",
			fileName: "sample@1.png",
			wantErr:  thread.ErrFileNameInvalid,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := thread.NewFileName(tt.fileName)

			require.ErrorIs(t, err, tt.wantErr)
		})
	}
}
