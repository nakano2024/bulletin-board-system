package thread

import (
	"context"
	"time"
)

type ThreadListItem struct {
	ID        string
	Body      string
	ImagePath *string
	CreatedAt time.Time
}

type IThreadQueryService interface {
	FetchThreadList(ctx context.Context) ([]ThreadListItem, error)
}
