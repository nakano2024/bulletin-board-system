package thread

import (
	"context"
	"time"
)

type ListThreadsCommand struct{}

type ListThreadsOutputThread struct {
	ID        string
	Body      string
	CreatedAt time.Time
}

type ListThreadsOutput struct {
	Threads []ListThreadsOutputThread
}

type ListThreadsUsecase struct {
	queryService IThreadQueryService
}

func NewListThreadsUsecase(queryService IThreadQueryService) *ListThreadsUsecase {
	return &ListThreadsUsecase{queryService: queryService}
}

func (u *ListThreadsUsecase) Exec(ctx context.Context, cmd ListThreadsCommand) (*ListThreadsOutput, error) {
	items, err := u.queryService.FetchThreadList(ctx)
	if err != nil {
		return nil, err
	}

	return &ListThreadsOutput{Threads: toListThreadsOutputThreads(items)}, nil
}

func toListThreadsOutputThreads(items []ThreadListItem) []ListThreadsOutputThread {
	threads := make([]ListThreadsOutputThread, 0, len(items))
	for _, item := range items {
		threads = append(threads, toListThreadsOutputThread(item))
	}
	return threads
}

func toListThreadsOutputThread(item ThreadListItem) ListThreadsOutputThread {
	return ListThreadsOutputThread{
		ID:        item.ID,
		Body:      item.Body,
		CreatedAt: item.CreatedAt,
	}
}
