package thread

import "errors"

// Errors a caller (e.g. a handler) can check with errors.Is. Exec translates the domain's sentinel errors into these,
// so callers never need to import domain/.
var (
	// ErrInvalidBody means the thread body was rejected (e.g. empty).
	ErrInvalidBody = errors.New("thread body is invalid")
	// ErrFileRequired means no image file was given (empty file id).
	ErrFileRequired = errors.New("thread file is required")
	// ErrFileNotFound means the given image file does not exist.
	ErrFileNotFound = errors.New("thread file is not found")
	// ErrFileAlreadyUsed means the given image file is already attached to another thread.
	ErrFileAlreadyUsed = errors.New("thread file is already used")
)
