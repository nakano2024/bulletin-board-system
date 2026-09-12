package thread

import "context"

type IThreadRepository interface {
	FetchActiveThreadListNewestFirst(ctx context.Context) (*ThreadList, error)
}
