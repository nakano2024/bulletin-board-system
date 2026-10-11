package objectstorage

import "context"

// DeploySignedURLIssuer implements application/usecase/image_file.ISignedURLIssuer for the deploy
// environment. The backing object storage for deploy is not decided yet; this is a skeleton only.
type DeploySignedURLIssuer struct{}

func NewDeploySignedURLIssuer() *DeploySignedURLIssuer {
	return &DeploySignedURLIssuer{}
}

func (i *DeploySignedURLIssuer) IssueUploadURL(ctx context.Context, fileName, contentType string, sizeBytes int64) (string, error) {
	panic("not implemented")
}
