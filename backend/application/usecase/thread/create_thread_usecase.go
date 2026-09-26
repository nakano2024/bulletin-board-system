package thread

import (
	"context"

	domainthread "github.com/nakanokota/bulletin-board-system/backend/domain/thread"
	domainuser "github.com/nakanokota/bulletin-board-system/backend/domain/user"
)

type CreateThreadCommand struct {
	IP       string
	Body     string
	FileName string
}

type CreateThreadOutput struct {
	ThreadID string
}

type CreateThreadUsecase struct {
	pendingThreadRepository domainthread.IPendingThreadRepository
	userService             *domainuser.UserService
	timeGetter              ITimeGetter
	logger                  ILogger
}

func NewCreateThreadUsecase(
	pendingThreadRepository domainthread.IPendingThreadRepository,
	userService *domainuser.UserService,
	timeGetter ITimeGetter,
	logger ILogger,
) *CreateThreadUsecase {
	return &CreateThreadUsecase{
		pendingThreadRepository: pendingThreadRepository,
		userService:             userService,
		timeGetter:              timeGetter,
		logger:                  logger,
	}
}

func (u *CreateThreadUsecase) Exec(ctx context.Context, cmd CreateThreadCommand) (*CreateThreadOutput, error) {
	fileName, err := domainthread.NewFileName(cmd.FileName)
	if err != nil {
		return nil, err
	}

	date, err := domainuser.NewUserCreateDate(u.timeGetter.Now(ctx))
	if err != nil {
		return nil, err
	}

	poster, err := u.userService.FetchOrCreate(ctx, cmd.IP, date)
	if err != nil {
		u.logger.Error(ctx, err)
		return nil, err
	}

	pendingThread, err := domainthread.NewPendingThread(poster.ID(), cmd.Body, fileName)
	if err != nil {
		return nil, err
	}

	createdThread, err := u.pendingThreadRepository.CreateThread(ctx, pendingThread)
	if err != nil {
		u.logger.Error(ctx, err)
		return nil, err
	}

	return &CreateThreadOutput{ThreadID: createdThread.ID()}, nil
}
