package image_file

import "errors"

var (
	ErrImageFileIDEmpty              = errors.New("image file id is empty")
	ErrImageFileContentTypeEmpty     = errors.New("image file content type is empty")
	ErrImageFileSizeBytesNotPositive = errors.New("image file size bytes is not positive")
)

// ImageFile is the metadata of an uploaded image file. Its name is always "{id}{extension}";
// the name the user originally sent is not kept.
type ImageFile struct {
	id          string
	name        *FileName
	contentType string
	sizeBytes   int64
}

// NewImageFile builds the name from id (a UUID generated outside the domain) and extension (e.g. ".png"),
// validating it via FileName's rule. sizeBytes must be positive.
func NewImageFile(id, extension, contentType string, sizeBytes int64) (*ImageFile, error) {
	if id == "" {
		return nil, ErrImageFileIDEmpty
	}
	name, err := NewFileName(id + extension)
	if err != nil {
		return nil, err
	}
	if contentType == "" {
		return nil, ErrImageFileContentTypeEmpty
	}
	if sizeBytes <= 0 {
		return nil, ErrImageFileSizeBytesNotPositive
	}

	return &ImageFile{id: id, name: name, contentType: contentType, sizeBytes: sizeBytes}, nil
}

func (f *ImageFile) ID() string {
	return f.id
}

func (f *ImageFile) Name() *FileName {
	return f.name
}

func (f *ImageFile) ContentType() string {
	return f.contentType
}

func (f *ImageFile) SizeBytes() int64 {
	return f.sizeBytes
}
