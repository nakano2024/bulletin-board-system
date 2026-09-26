package user

import (
	"context"
)

type IUserRepository interface {
	// FindByIPAndDate returns nil, nil when no User is found for ip on date.
	FindByIPAndDate(ctx context.Context, ip string, date *UserCreateDate) (*User, error)
	// CreateUser assigns a new id (an infra-internal concern) and persists pendingUser, which already
	// carries the ip and createDate it should be recorded with.
	CreateUser(ctx context.Context, pendingUser *PendingUser) (*User, error)
}
