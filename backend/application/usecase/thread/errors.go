package thread

import "errors"

// Errors a caller (e.g. a handler) can check with errors.Is. Exec translates the domain's sentinel errors into these,
// so callers never need to import domain/.
var (
	// ErrInvalidBody means the thread body was rejected (e.g. empty).
	ErrInvalidBody = errors.New("thread body is invalid")
	// ErrInvalidFile means no usable image file was given (empty file id, or no such file).
	ErrInvalidFile = errors.New("thread file is invalid")
	// ErrFileAlreadyUsed means the given image file is already attached to another thread.
	ErrFileAlreadyUsed = errors.New("thread file is already used")
)
