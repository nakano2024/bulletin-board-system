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
	txManager             ITransactionManager
	logger                ILogger
}

func NewCreateThreadUsecase(
	threadCreationService *domainthread.ThreadCreationService,
	userService *domainuser.UserService,
	timeGetter ITimeGetter,
	txManager ITransactionManager,
	logger ILogger,
) *CreateThreadUsecase {
	return &CreateThreadUsecase{
		threadCreationService: threadCreationService,
		userService:           userService,
		timeGetter:            timeGetter,
		txManager:             txManager,
		logger:                logger,
	}
}

// Exec runs the whole creation (resolving the poster through saving the thread) in one transaction,
// so a failure leaves neither a new user nor a thread behind.
func (u *CreateThreadUsecase) Exec(ctx context.Context, cmd CreateThreadCommand) (*CreateThreadOutput, error) {
	date, err := domainuser.NewUserCreateDate(u.timeGetter.Now(ctx))
	if err != nil {
		return nil, err
	}

	var createdThread *domainthread.Thread
	err = u.txManager.RunInTx(ctx, func(txCtx context.Context) error {
		poster, err := u.userService.FetchOrCreate(txCtx, cmd.IP, date)
		if err != nil {
			return err
		}

		pendingThread, err := domainthread.NewPendingThread(poster.ID(), cmd.Body, cmd.FileID)
		if err != nil {
			return err
		}

		createdThread, err = u.threadCreationService.Create(txCtx, pendingThread)
		return err
	})
	if err != nil {
		// PendingThread's own validation errors are user-caused input rejections, not a dependency
		// failure, so they are not logged here; everything else (service/repository/commit errors) is.
		if !isPendingThreadValidationError(err) {
			u.logger.Error(ctx, err)
		}
		return nil, toCreateThreadError(err)
	}

	return &CreateThreadOutput{ThreadID: createdThread.ID()}, nil
}

func isPendingThreadValidationError(err error) bool {
	return errors.Is(err, domainthread.ErrPendingThreadBodyEmpty) || errors.Is(err, domainthread.ErrPendingThreadFileIDEmpty)
}

// toCreateThreadError wraps the domain's sentinel errors in this package's errors so callers can classify them
// without importing domain/; the original error stays in the chain. Other errors are returned unchanged.
func toCreateThreadError(err error) error {
	switch {
	case errors.Is(err, domainthread.ErrPendingThreadBodyEmpty):
		return fmt.Errorf("%w: %w", ErrInvalidBody, err)
	case errors.Is(err, domainthread.ErrPendingThreadFileIDEmpty):
		return fmt.Errorf("%w: %w", ErrFileRequired, err)
	case errors.Is(err, domainthread.ErrFileNotFound):
		return fmt.Errorf("%w: %w", ErrFileNotFound, err)
	case errors.Is(err, domainthread.ErrFileAlreadyUsed):
		return fmt.Errorf("%w: %w", ErrFileAlreadyUsed, err)
	default:
		return err
	}
}
