package image_file

import "context"

// ISignedURLIssuer issues a signed URL that lets the client upload the raw image bytes
// directly to object storage (a presigned PUT URL for fileName), without the bytes
// passing through this backend. contentType and sizeBytes are bound into the signature
// so the upload is rejected unless the client's PUT sends the same Content-Type and body size.
type ISignedURLIssuer interface {
	IssueUploadURL(ctx context.Context, fileName, contentType string, sizeBytes int64) (string, error)
}
