package user

import (
	"context"
	"time"
)

type IUserRepository interface {
	// FindByIPAndDate returns nil, nil when no User is found for ip on date.
	FindByIPAndDate(ctx context.Context, ip string, date time.Time) (*User, error)
	// CreateUser assigns a new id (an infra-internal concern) and persists a User for pendingUser's ip on date.
	CreateUser(ctx context.Context, pendingUser *PendingUser, date time.Time) (*User, error)
}
