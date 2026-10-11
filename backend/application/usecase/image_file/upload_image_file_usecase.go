package image_file

import (
	"context"

	domainimagefile "github.com/nakanokota/bulletin-board-system/backend/domain/image_file"
)

type UploadImageFileCommand struct {
	Extension   string
	ContentType string
	SizeBytes   int64
}

type UploadImageFileOutput struct {
	File      UploadImageFileOutputFile
	SignedURL string
}

// ContentType and ContentLength are the headers the client's PUT to SignedURL must send exactly,
// since they are bound into the presigned URL's signature.
type UploadImageFileOutputFile struct {
	FileID        string
	ContentType   string
	ContentLength int64
}

type UploadImageFileUsecase struct {
	imageFileRepository domainimagefile.IImageFileRepository
	uuidGenerator       IUuidGenerator
	signedURLIssuer     ISignedURLIssuer
	logger              ILogger
}

func NewUploadImageFileUsecase(
	imageFileRepository domainimagefile.IImageFileRepository,
	uuidGenerator IUuidGenerator,
	signedURLIssuer ISignedURLIssuer,
	logger ILogger,
) *UploadImageFileUsecase {
	return &UploadImageFileUsecase{
		imageFileRepository: imageFileRepository,
		uuidGenerator:       uuidGenerator,
		signedURLIssuer:     signedURLIssuer,
		logger:              logger,
	}
}

func (u *UploadImageFileUsecase) Exec(ctx context.Context, cmd UploadImageFileCommand) (*UploadImageFileOutput, error) {
	id := u.uuidGenerator.NewUUID(ctx)

	imageFile, err := domainimagefile.NewImageFile(id, cmd.Extension, cmd.ContentType, cmd.SizeBytes)
	if err != nil {
		return nil, toUploadImageFileError(err)
	}

	if err := u.imageFileRepository.CreateImageFile(ctx, imageFile); err != nil {
		u.logger.Error(ctx, err)
		return nil, err
	}

	signedURL, err := u.signedURLIssuer.IssueUploadURL(ctx, imageFile.Name().Value(), imageFile.ContentType(), imageFile.SizeBytes())
	if err != nil {
		u.logger.Error(ctx, err)
		return nil, err
	}

	return &UploadImageFileOutput{
		File: UploadImageFileOutputFile{
			FileID:        id,
			ContentType:   imageFile.ContentType(),
			ContentLength: imageFile.SizeBytes(),
		},
		SignedURL: signedURL,
	}, nil
}
