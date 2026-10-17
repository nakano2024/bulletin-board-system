package thread

import "errors"

var (
	// ErrThreadNotFound means no thread has the requested id.
	ErrThreadNotFound = errors.New("thread not found")
	// ErrThreadDead means the thread exists but has fallen off (dat落ち).
	ErrThreadDead = errors.New("thread is dead")
)
