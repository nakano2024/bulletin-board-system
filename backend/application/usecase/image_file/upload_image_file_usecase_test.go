package image_file_test

import (
	"context"
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"

	"github.com/nakanokota/bulletin-board-system/backend/application/mock/image_file"
	"github.com/nakanokota/bulletin-board-system/backend/application/usecase/image_file"
	domainimagefile "github.com/nakanokota/bulletin-board-system/backend/domain/image_file"
)

type uploadImageFileMocks struct {
	imageFileRepository *mock_image_file.MockIImageFileRepository
	uuidGenerator       *mock_image_file.MockIUuidGenerator
	signedURLIssuer     *mock_image_file.MockISignedURLIssuer
	logger              *mock_image_file.MockILogger
}

func newUploadImageFileUsecase(t *testing.T, setupMocks func(uploadImageFileMocks)) *image_file.UploadImageFileUsecase {
	t.Helper()
	ctrl := gomock.NewController(t)
	m := uploadImageFileMocks{
		imageFileRepository: mock_image_file.NewMockIImageFileRepository(ctrl),
		uuidGenerator:       mock_image_file.NewMockIUuidGenerator(ctrl),
		signedURLIssuer:     mock_image_file.NewMockISignedURLIssuer(ctrl),
		logger:              mock_image_file.NewMockILogger(ctrl),
	}
	setupMocks(m)

	return image_file.NewUploadImageFileUsecase(m.imageFileRepository, m.uuidGenerator, m.signedURLIssuer, m.logger)
}

func TestUploadImageFileUsecase_Exec_正常系(t *testing.T) {
	tests := []struct {
		name       string
		cmd        image_file.UploadImageFileCommand
		setupMocks func(uploadImageFileMocks)
		want       *image_file.UploadImageFileOutput
	}{
		{
			name: "拡張子・ContentType・SizeBytesがすべて正当な値のとき、OutputのFile.FileIDに生成されたUUIDが、File.ContentType・File.ContentLengthにCommandのContentType・SizeBytesが設定され、SignedURLにISignedURLIssuerが返した値が設定されること",
			cmd:  image_file.UploadImageFileCommand{Extension: ".png", ContentType: "image/png", SizeBytes: 1024},
			setupMocks: func(m uploadImageFileMocks) {
				m.uuidGenerator.EXPECT().NewUUID(gomock.Any()).Return("uuid-1")
				m.imageFileRepository.EXPECT().CreateImageFile(gomock.Any(), gomock.Any()).Return(nil)
				m.signedURLIssuer.EXPECT().IssueUploadURL(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return("https://example.com/signed", nil)
			},
			want: &image_file.UploadImageFileOutput{
				File: image_file.UploadImageFileOutputFile{
					FileID:        "uuid-1",
					ContentType:   "image/png",
					ContentLength: 1024,
				},
				SignedURL: "https://example.com/signed",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			sut := newUploadImageFileUsecase(t, tt.setupMocks)

			got, err := sut.Exec(context.Background(), tt.cmd)

			require.NoError(t, err)
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestUploadImageFileUsecase_Exec_依存に渡す値(t *testing.T) {
	expectedImageFile, err := domainimagefile.NewImageFile("uuid-1", ".png", "image/png", 1024)
	require.NoError(t, err)

	tests := []struct {
		name       string
		cmd        image_file.UploadImageFileCommand
		setupMocks func(uploadImageFileMocks)
	}{
		{
			name: "CreateImageFileに、IDがIUuidGeneratorの返した値、拡張子・ContentType・SizeBytesがCommandの値と一致するImageFileが渡されること",
			cmd:  image_file.UploadImageFileCommand{Extension: ".png", ContentType: "image/png", SizeBytes: 1024},
			setupMocks: func(m uploadImageFileMocks) {
				m.uuidGenerator.EXPECT().NewUUID(gomock.Any()).Return("uuid-1")
				m.imageFileRepository.EXPECT().CreateImageFile(gomock.Any(), expectedImageFile).Return(nil)
				m.signedURLIssuer.EXPECT().IssueUploadURL(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return("https://example.com/signed", nil)
			},
		},
		{
			name: "IssueUploadURLに、生成されたUUID+拡張子のファイル名が渡されること",
			cmd:  image_file.UploadImageFileCommand{Extension: ".png", ContentType: "image/png", SizeBytes: 1024},
			setupMocks: func(m uploadImageFileMocks) {
				m.uuidGenerator.EXPECT().NewUUID(gomock.Any()).Return("uuid-1")
				m.imageFileRepository.EXPECT().CreateImageFile(gomock.Any(), gomock.Any()).Return(nil)
				m.signedURLIssuer.EXPECT().IssueUploadURL(gomock.Any(), "uuid-1.png", gomock.Any(), gomock.Any()).Return("https://example.com/signed", nil)
			},
		},
		{
			name: "IssueUploadURLに、CommandのContentTypeが渡されること",
			cmd:  image_file.UploadImageFileCommand{Extension: ".png", ContentType: "image/png", SizeBytes: 1024},
			setupMocks: func(m uploadImageFileMocks) {
				m.uuidGenerator.EXPECT().NewUUID(gomock.Any()).Return("uuid-1")
				m.imageFileRepository.EXPECT().CreateImageFile(gomock.Any(), gomock.Any()).Return(nil)
				m.signedURLIssuer.EXPECT().IssueUploadURL(gomock.Any(), gomock.Any(), "image/png", gomock.Any()).Return("https://example.com/signed", nil)
			},
		},
		{
			name: "IssueUploadURLに、CommandのSizeBytesが渡されること",
			cmd:  image_file.UploadImageFileCommand{Extension: ".png", ContentType: "image/png", SizeBytes: 1024},
			setupMocks: func(m uploadImageFileMocks) {
				m.uuidGenerator.EXPECT().NewUUID(gomock.Any()).Return("uuid-1")
				m.imageFileRepository.EXPECT().CreateImageFile(gomock.Any(), gomock.Any()).Return(nil)
				m.signedURLIssuer.EXPECT().IssueUploadURL(gomock.Any(), gomock.Any(), gomock.Any(), int64(1024)).Return("https://example.com/signed", nil)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			sut := newUploadImageFileUsecase(t, tt.setupMocks)

			_, err := sut.Exec(context.Background(), tt.cmd)

			require.NoError(t, err)
		})
	}
}

func TestUploadImageFileUsecase_Exec_異常系(t *testing.T) {
	errRepositoryFailed := errors.New("repository failed")
	errIssuerFailed := errors.New("issuer failed")

	tests := []struct {
		name       string
		cmd        image_file.UploadImageFileCommand
		setupMocks func(uploadImageFileMocks)
		wantErr    error
	}{
		{
			name: "Extensionが許可されていない拡張子のとき、ErrInvalidExtensionが返ること",
			cmd:  image_file.UploadImageFileCommand{Extension: ".exe", ContentType: "image/png", SizeBytes: 1024},
			setupMocks: func(m uploadImageFileMocks) {
				m.uuidGenerator.EXPECT().NewUUID(gomock.Any()).Return("uuid-1")
			},
			wantErr: image_file.ErrInvalidExtension,
		},
		{
			name: "ContentTypeが空文字のとき、ErrContentTypeRequiredが返ること",
			cmd:  image_file.UploadImageFileCommand{Extension: ".png", ContentType: "", SizeBytes: 1024},
			setupMocks: func(m uploadImageFileMocks) {
				m.uuidGenerator.EXPECT().NewUUID(gomock.Any()).Return("uuid-1")
			},
			wantErr: image_file.ErrContentTypeRequired,
		},
		{
			name: "SizeBytesが0以下のとき、ErrSizeBytesInvalidが返ること",
			cmd:  image_file.UploadImageFileCommand{Extension: ".png", ContentType: "image/png", SizeBytes: 0},
			setupMocks: func(m uploadImageFileMocks) {
				m.uuidGenerator.EXPECT().NewUUID(gomock.Any()).Return("uuid-1")
			},
			wantErr: image_file.ErrSizeBytesInvalid,
		},
		{
			name: "imageFileRepository.CreateImageFileがエラーを返したとき、そのエラーが返ること",
			cmd:  image_file.UploadImageFileCommand{Extension: ".png", ContentType: "image/png", SizeBytes: 1024},
			setupMocks: func(m uploadImageFileMocks) {
				m.uuidGenerator.EXPECT().NewUUID(gomock.Any()).Return("uuid-1")
				m.imageFileRepository.EXPECT().CreateImageFile(gomock.Any(), gomock.Any()).Return(errRepositoryFailed)
			},
			wantErr: errRepositoryFailed,
		},
		{
			name: "signedURLIssuer.IssueUploadURLがエラーを返したとき、そのエラーが返ること",
			cmd:  image_file.UploadImageFileCommand{Extension: ".png", ContentType: "image/png", SizeBytes: 1024},
			setupMocks: func(m uploadImageFileMocks) {
				m.uuidGenerator.EXPECT().NewUUID(gomock.Any()).Return("uuid-1")
				m.imageFileRepository.EXPECT().CreateImageFile(gomock.Any(), gomock.Any()).Return(nil)
				m.signedURLIssuer.EXPECT().IssueUploadURL(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return("", errIssuerFailed)
			},
			wantErr: errIssuerFailed,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			sut := newUploadImageFileUsecase(t, func(m uploadImageFileMocks) {
				m.logger.EXPECT().Error(gomock.Any(), gomock.Any()).AnyTimes()
				tt.setupMocks(m)
			})

			_, err := sut.Exec(context.Background(), tt.cmd)

			require.ErrorIs(t, err, tt.wantErr)
		})
	}
}

func TestUploadImageFileUsecase_Exec_ログ出力(t *testing.T) {
	errRepositoryFailed := errors.New("repository failed")
	errIssuerFailed := errors.New("issuer failed")

	tests := []struct {
		name       string
		cmd        image_file.UploadImageFileCommand
		setupMocks func(uploadImageFileMocks)
	}{
		{
			name: "imageFileRepository.CreateImageFileがエラーを返したとき、ILoggerにそのエラーが渡されること",
			cmd:  image_file.UploadImageFileCommand{Extension: ".png", ContentType: "image/png", SizeBytes: 1024},
			setupMocks: func(m uploadImageFileMocks) {
				m.uuidGenerator.EXPECT().NewUUID(gomock.Any()).Return("uuid-1")
				m.imageFileRepository.EXPECT().CreateImageFile(gomock.Any(), gomock.Any()).Return(errRepositoryFailed)
				m.logger.EXPECT().Error(gomock.Any(), errRepositoryFailed).Times(1)
			},
		},
		{
			name: "signedURLIssuer.IssueUploadURLがエラーを返したとき、ILoggerにそのエラーが渡されること",
			cmd:  image_file.UploadImageFileCommand{Extension: ".png", ContentType: "image/png", SizeBytes: 1024},
			setupMocks: func(m uploadImageFileMocks) {
				m.uuidGenerator.EXPECT().NewUUID(gomock.Any()).Return("uuid-1")
				m.imageFileRepository.EXPECT().CreateImageFile(gomock.Any(), gomock.Any()).Return(nil)
				m.signedURLIssuer.EXPECT().IssueUploadURL(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return("", errIssuerFailed)
				m.logger.EXPECT().Error(gomock.Any(), errIssuerFailed).Times(1)
			},
		},
		{
			name: "Extensionが不正なとき、ILoggerが呼ばれないこと",
			cmd:  image_file.UploadImageFileCommand{Extension: ".exe", ContentType: "image/png", SizeBytes: 1024},
			setupMocks: func(m uploadImageFileMocks) {
				m.uuidGenerator.EXPECT().NewUUID(gomock.Any()).Return("uuid-1")
				m.logger.EXPECT().Error(gomock.Any(), gomock.Any()).Times(0)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			sut := newUploadImageFileUsecase(t, tt.setupMocks)

			_, err := sut.Exec(context.Background(), tt.cmd)

			require.Error(t, err)
		})
	}
}
