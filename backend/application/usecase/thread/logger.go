package thread

import "context"

type ILogger interface {
	Error(ctx context.Context, err error)
}
