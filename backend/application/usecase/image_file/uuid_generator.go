package image_file

import "context"

type IUuidGenerator interface {
	NewUUID(ctx context.Context) string
}
