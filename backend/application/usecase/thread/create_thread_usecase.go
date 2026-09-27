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
	fnSucceeded := false
	err = u.txManager.RunInTx(ctx, func(txCtx context.Context) error {
		poster, err := u.userService.FetchOrCreate(txCtx, cmd.IP, date)
		if err != nil {
			u.logger.Error(ctx, err)
			return err
		}

		pendingThread, err := domainthread.NewPendingThread(poster.ID(), cmd.Body, cmd.FileID)
		if err != nil {
			return err
		}

		createdThread, err = u.threadCreationService.Create(txCtx, pendingThread)
		if err != nil {
			u.logger.Error(ctx, err)
			return err
		}

		fnSucceeded = true
		return nil
	})
	if err != nil {
		// Errors from inside fn were already logged there; only a failure of the transaction itself (e.g. commit) is new.
		if fnSucceeded {
			u.logger.Error(ctx, err)
		}
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
