package thread

import "context"

type IThreadRepository interface {
	FetchActiveThreadList(ctx context.Context) (*ThreadList, error)
}
