package thread

import (
	"context"
	"time"

	domainthread "github.com/nakanokota/bulletin-board-system/backend/domain/thread"
)

type ListThreadsCommand struct{}

type ListThreadsOutputThread struct {
	ID        string
	Body      string
	ImagePath string
	CreatedAt time.Time
}

type ListThreadsOutput struct {
	Threads []ListThreadsOutputThread
}

type ListThreadsUsecase struct {
	threadFetcher domainthread.IThreadFetcher
	logger        ILogger
}

func NewListThreadsUsecase(threadFetcher domainthread.IThreadFetcher, logger ILogger) *ListThreadsUsecase {
	return &ListThreadsUsecase{threadFetcher: threadFetcher, logger: logger}
}

func (u *ListThreadsUsecase) Exec(ctx context.Context, cmd ListThreadsCommand) (*ListThreadsOutput, error) {
	threadList, err := u.threadFetcher.FetchActiveThreadListNewestFirst(ctx)
	if err != nil {
		u.logger.Error(ctx, err)
		return nil, err
	}

	return &ListThreadsOutput{Threads: toListThreadsOutputThreads(threadList.Threads())}, nil
}

func toListThreadsOutputThreads(threads []*domainthread.Thread) []ListThreadsOutputThread {
	outputs := make([]ListThreadsOutputThread, 0, len(threads))
	for _, t := range threads {
		outputs = append(outputs, toListThreadsOutputThread(t))
	}
	return outputs
}

func toListThreadsOutputThread(t *domainthread.Thread) ListThreadsOutputThread {
	return ListThreadsOutputThread{
		ID:        t.ID(),
		Body:      t.Body(),
		ImagePath: t.FilePathValue(),
		CreatedAt: t.CreatedAt(),
	}
}
