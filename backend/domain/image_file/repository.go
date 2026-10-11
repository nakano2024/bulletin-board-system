package image_file

import "context"

// IImageFileRepository persists image file metadata.
type IImageFileRepository interface {
	// CreateImageFile registers imageFile. The creation time is assigned by infra.
	CreateImageFile(ctx context.Context, imageFile *ImageFile) error
}
