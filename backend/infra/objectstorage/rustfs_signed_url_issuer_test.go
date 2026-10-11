package objectstorage_test

import (
	"context"
	"net/url"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/nakanokota/bulletin-board-system/backend/infra/objectstorage"
)

func TestRustFSSignedURLIssuer_IssueUploadURL_正常系(t *testing.T) {
	tests := []struct {
		name        string
		endpoint    string
		bucket      string
		fileName    string
		contentType string
		sizeBytes   int64
	}{
		{
			name:        "有効なfileName・contentType・sizeBytesを渡したとき、endpoint・bucket・fileNameを反映し、有効期限300秒(5分)・contentType・content-lengthが署名対象ヘッダーに含まれるURLが発行されること",
			endpoint:    "http://localhost:9000",
			bucket:      "images",
			fileName:    "uuid-1.png",
			contentType: "image/png",
			sizeBytes:   1024,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			sut := objectstorage.NewRustFSSignedURLIssuer(tt.endpoint, "access-key", "secret-key", tt.bucket)

			got, err := sut.IssueUploadURL(context.Background(), tt.fileName, tt.contentType, tt.sizeBytes)

			require.NoError(t, err)
			parsed, err := url.Parse(got)
			require.NoError(t, err)
			assert.Equal(t, "localhost:9000", parsed.Host)
			assert.True(t, strings.Contains(parsed.Path, tt.bucket))
			assert.True(t, strings.Contains(parsed.Path, tt.fileName))
			assert.Equal(t, "300", parsed.Query().Get("X-Amz-Expires"))
			assert.True(t, strings.Contains(parsed.Query().Get("X-Amz-SignedHeaders"), "content-type"))
			assert.True(t, strings.Contains(parsed.Query().Get("X-Amz-SignedHeaders"), "content-length"))
		})
	}
}
