package thread

import (
	"context"
	"errors"
)

var (
	// ErrFileNotFound means the PendingThread's file id does not refer to a registered file.
	ErrFileNotFound = errors.New("file not found")
	// ErrFileAlreadyUsed means the PendingThread's file is already attached to another thread (a thread's image is unique to it).
	ErrFileAlreadyUsed = errors.New("file is already used by another thread")
)

// IThreadFetcher provides read access to persisted threads.
type IThreadFetcher interface {
	FetchActiveThreadListNewestFirst(ctx context.Context) (*ThreadList, error)
}

// IPendingThreadRepository persists a PendingThread, turning it into a full Thread.
type IPendingThreadRepository interface {
	// CreateThread assigns a new id and creation time (infra-internal concerns) and persists pendingThread.
	// ThreadCreationService checks the file beforehand; CreateThread still returns ErrFileNotFound / ErrFileAlreadyUsed
	// when a concurrent request slips past that check and the DB constraints reject the insert.
	CreateThread(ctx context.Context, pendingThread *PendingThread) (*Thread, error)
}

// IThreadFileChecker answers the lookups ThreadCreationService needs to decide whether a file can be attached to a new thread.
type IThreadFileChecker interface {
	ExistsFile(ctx context.Context, fileID string) (bool, error)
	IsFileAttachedToThread(ctx context.Context, fileID string) (bool, error)
}
