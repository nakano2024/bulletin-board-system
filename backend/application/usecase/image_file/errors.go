package image_file

import (
	"errors"
	"fmt"

	domainimagefile "github.com/nakanokota/bulletin-board-system/backend/domain/image_file"
)

// Errors a caller (e.g. a handler) can check with errors.Is. Exec translates the domain's sentinel errors into these,
// so callers never need to import domain/.
var (
	// ErrInvalidExtension means the given extension is not an allowed image extension.
	ErrInvalidExtension = errors.New("image file extension is invalid")
	// ErrContentTypeRequired means the content type was empty.
	ErrContentTypeRequired = errors.New("image file content type is required")
	// ErrSizeBytesInvalid means the size in bytes was not positive.
	ErrSizeBytesInvalid = errors.New("image file size bytes is invalid")
)

// toUploadImageFileError wraps the domain's sentinel errors in this package's errors so callers can classify them
// without importing domain/; the original error stays in the chain. Other errors are returned unchanged.
func toUploadImageFileError(err error) error {
	switch {
	case errors.Is(err, domainimagefile.ErrFileNameInvalid):
		return fmt.Errorf("%w: %w", ErrInvalidExtension, err)
	case errors.Is(err, domainimagefile.ErrImageFileContentTypeEmpty):
		return fmt.Errorf("%w: %w", ErrContentTypeRequired, err)
	case errors.Is(err, domainimagefile.ErrImageFileSizeBytesNotPositive):
		return fmt.Errorf("%w: %w", ErrSizeBytesInvalid, err)
	default:
		return err
	}
}
