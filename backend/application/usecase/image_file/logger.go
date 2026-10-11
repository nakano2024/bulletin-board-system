package image_file

import "context"

type ILogger interface {
	Error(ctx context.Context, err error)
}
