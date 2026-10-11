package e2e

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/labstack/echo/v4"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	applicationimagefile "github.com/nakanokota/bulletin-board-system/backend/application/usecase/image_file"
	"github.com/nakanokota/bulletin-board-system/backend/handler"
	handlerimagefile "github.com/nakanokota/bulletin-board-system/backend/handler/image_file"
	infralog "github.com/nakanokota/bulletin-board-system/backend/infra/log"
	"github.com/nakanokota/bulletin-board-system/backend/infra/objectstorage"
	"github.com/nakanokota/bulletin-board-system/backend/infra/postgres"
	infrauuid "github.com/nakanokota/bulletin-board-system/backend/infra/uuid"
)

// TestE2E_UploadImageFile_正常系 wires the real Handler -> Usecase -> Repository -> DB stack (no mocks)
// and checks POST /files persists the file metadata. RustFSSignedURLIssuer's presign is a local
// signature computation with no network call, so it succeeds here without a running RustFS; the
// signed URL's own content is already verified in the Infra layer's test, so this only checks it
// is a non-empty string.
func TestE2E_UploadImageFile_正常系(t *testing.T) {
	ctx := context.Background()

	pool, err := pgxpool.New(ctx, os.Getenv("E2E_DATABASE_URL"))
	require.NoError(t, err)
	t.Cleanup(pool.Close)

	uploadImageFileHandler := newE2EUploadImageFileHandler(pool)

	e := echo.New()
	e.Validator = handler.NewRequestValidator()
	req := httptest.NewRequest(http.MethodPost, "/files", strings.NewReader(`{"file":{"original_file_name":"sample.png","content_type":"image/png","size_bytes":1024}}`))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	err = uploadImageFileHandler.UploadImageFile(c)
	require.NoError(t, err)

	assert.Equal(t, http.StatusCreated, rec.Code)

	var got handlerimagefile.UploadImageFileResponse
	require.NoError(t, json.Unmarshal(rec.Body.Bytes(), &got))
	require.NotEmpty(t, got.File.FileID)
	t.Cleanup(func() {
		_, _ = pool.Exec(context.Background(), "DELETE FROM files WHERE id = $1", got.File.FileID)
	})

	assert.Equal(t, "image/png", got.File.ContentType)
	assert.Equal(t, int64(1024), got.File.ContentLength)
	assert.NotEmpty(t, got.SignedURL)

	var storedName, storedContentType string
	var storedSizeBytes int64
	err = pool.QueryRow(ctx, "SELECT name, content_type, size_bytes FROM files WHERE id = $1", got.File.FileID).
		Scan(&storedName, &storedContentType, &storedSizeBytes)
	require.NoError(t, err)
	assert.Equal(t, got.File.FileID+".png", storedName)
	assert.Equal(t, "image/png", storedContentType)
	assert.Equal(t, int64(1024), storedSizeBytes)
}

// newE2EUploadImageFileHandler assembles POST /files the same way main.go does, with no mocks.
func newE2EUploadImageFileHandler(pool *pgxpool.Pool) *handlerimagefile.UploadImageFileHandler {
	imageFileRepository := postgres.NewPostgresImageFileRepository(pool)
	uuidGenerator := infrauuid.NewGoogleUuidGenerator()
	signedURLIssuer := objectstorage.NewRustFSSignedURLIssuer("http://localhost:9000", "access-key", "secret-key", "images")
	logger := infralog.NewStdLogger()
	uploadImageFileUsecase := applicationimagefile.NewUploadImageFileUsecase(imageFileRepository, uuidGenerator, signedURLIssuer, logger)
	return handlerimagefile.NewUploadImageFileHandler(uploadImageFileUsecase)
}
