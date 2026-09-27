package thread

import (
	"context"
	"errors"
	"fmt"

	domainthread "github.com/nakanokota/bulletin-board-system/backend/domain/thread"
	domainuser "github.com/nakanokota/bulletin-board-system/backend/domain/user"
)

type CreateThreadCommand struct {
	IP     string
	Body   string
	FileID string
}

type CreateThreadOutput struct {
	ThreadID string
}

type CreateThreadUsecase struct {
	threadCreationService *domainthread.ThreadCreationService
	userService           *domainuser.UserService
	timeGetter            ITimeGetter
	logger                ILogger
}

func NewCreateThreadUsecase(
	threadCreationService *domainthread.ThreadCreationService,
	userService *domainuser.UserService,
	timeGetter ITimeGetter,
	logger ILogger,
) *CreateThreadUsecase {
	return &CreateThreadUsecase{
		threadCreationService: threadCreationService,
		userService:           userService,
		timeGetter:            timeGetter,
		logger:                logger,
	}
}

func (u *CreateThreadUsecase) Exec(ctx context.Context, cmd CreateThreadCommand) (*CreateThreadOutput, error) {
	date, err := domainuser.NewUserCreateDate(u.timeGetter.Now(ctx))
	if err != nil {
		return nil, err
	}

	poster, err := u.userService.FetchOrCreate(ctx, cmd.IP, date)
	if err != nil {
		u.logger.Error(ctx, err)
		return nil, err
	}

	pendingThread, err := domainthread.NewPendingThread(poster.ID(), cmd.Body, cmd.FileID)
	if err != nil {
		return nil, toCreateThreadError(err)
	}

	createdThread, err := u.threadCreationService.Create(ctx, pendingThread)
	if err != nil {
		u.logger.Error(ctx, err)
		return nil, toCreateThreadError(err)
	}

	return &CreateThreadOutput{ThreadID: createdThread.ID()}, nil
}

// toCreateThreadError wraps the domain's sentinel errors in this package's errors so callers can classify them
// without importing domain/; the original error stays in the chain. Other errors are returned unchanged.
func toCreateThreadError(err error) error {
	switch {
	case errors.Is(err, domainthread.ErrPendingThreadBodyEmpty):
		return fmt.Errorf("%w: %w", ErrInvalidBody, err)
	case errors.Is(err, domainthread.ErrPendingThreadFileIDEmpty), errors.Is(err, domainthread.ErrFileNotFound):
		return fmt.Errorf("%w: %w", ErrInvalidFile, err)
	case errors.Is(err, domainthread.ErrFileAlreadyUsed):
		return fmt.Errorf("%w: %w", ErrFileAlreadyUsed, err)
	default:
		return err
	}
}
