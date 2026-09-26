package thread

import "context"

// IThreadFetcher provides read access to persisted threads.
type IThreadFetcher interface {
	FetchActiveThreadListNewestFirst(ctx context.Context) (*ThreadList, error)
}

// IPendingThreadRepository persists a PendingThread, turning it into a full Thread.
type IPendingThreadRepository interface {
	// CreateThread assigns a new id and creation time (infra-internal concerns) and persists pendingThread.
	CreateThread(ctx context.Context, pendingThread *PendingThread) (*Thread, error)
}
