package thread

import (
	"context"
	"time"
)

type ITimeGetter interface {
	Now(ctx context.Context) time.Time
}
