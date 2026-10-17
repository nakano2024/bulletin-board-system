package thread

import (
	"context"
	"errors"
	"time"

	domainthread "github.com/nakanokota/bulletin-board-system/backend/domain/thread"
)

type GetThreadCommand struct {
	ThreadID string
}

type GetThreadOutputReply struct {
	ID     string
	UserID string
	Body   string
	// ImagePath is empty when the reply has no attached image.
	ImagePath string
	CreatedAt time.Time
}

type GetThreadOutputThread struct {
	ID        string
	UserID    string
	Body      string
	ImagePath string
	CreatedAt time.Time
	// Replies are newest first.
	Replies []GetThreadOutputReply
}

type GetThreadOutput struct {
	Thread GetThreadOutputThread
}

type GetThreadUsecase struct {
	threadDetailFetcher domainthread.IThreadDetailFetcher
	logger              ILogger
}

func NewGetThreadUsecase(threadDetailFetcher domainthread.IThreadDetailFetcher, logger ILogger) *GetThreadUsecase {
	return &GetThreadUsecase{threadDetailFetcher: threadDetailFetcher, logger: logger}
}

// Exec returns the thread with its replies. It returns ErrThreadNotFound if no thread has cmd.ThreadID,
// and ErrThreadDead if the thread has fallen off (dat落ち).
func (u *GetThreadUsecase) Exec(ctx context.Context, cmd GetThreadCommand) (*GetThreadOutput, error) {
	threadDetail, err := u.threadDetailFetcher.FetchThreadDetail(ctx, cmd.ThreadID)
	if err != nil {
		// A missing thread is caused by the requested id, not a dependency failure, so it is not logged.
		if errors.Is(err, domainthread.ErrThreadNotFound) {
			return nil, ErrThreadNotFound
		}
		u.logger.Error(ctx, err)
		return nil, err
	}

	if threadDetail.IsDead() {
		return nil, ErrThreadDead
	}

	return &GetThreadOutput{Thread: toGetThreadOutputThread(threadDetail)}, nil
}

func toGetThreadOutputThread(t *domainthread.ThreadDetail) GetThreadOutputThread {
	return GetThreadOutputThread{
		ID:        t.ID(),
		UserID:    t.UserID(),
		Body:      t.Body(),
		ImagePath: t.FilePath().Value(),
		CreatedAt: t.CreatedAt(),
		Replies:   toGetThreadOutputReplies(t.Replies()),
	}
}

func toGetThreadOutputReplies(replies []*domainthread.Reply) []GetThreadOutputReply {
	outputs := make([]GetThreadOutputReply, 0, len(replies))
	for _, r := range replies {
		outputs = append(outputs, toGetThreadOutputReply(r))
	}
	return outputs
}

func toGetThreadOutputReply(r *domainthread.Reply) GetThreadOutputReply {
	imagePath := ""
	if r.FilePath() != nil {
		imagePath = r.FilePath().Value()
	}

	return GetThreadOutputReply{
		ID:        r.ID(),
		UserID:    r.UserID(),
		Body:      r.Body(),
		ImagePath: imagePath,
		CreatedAt: r.CreatedAt(),
	}
}
