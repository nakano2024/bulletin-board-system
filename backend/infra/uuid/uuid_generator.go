package uuid

import (
	"context"

	googleuuid "github.com/google/uuid"
)

// GoogleUuidGenerator implements application/usecase/image_file.IUuidGenerator using google/uuid.
type GoogleUuidGenerator struct{}

func NewGoogleUuidGenerator() *GoogleUuidGenerator {
	return &GoogleUuidGenerator{}
}

func (g *GoogleUuidGenerator) NewUUID(ctx context.Context) string {
	return googleuuid.NewString()
}
