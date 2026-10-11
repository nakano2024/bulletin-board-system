package objectstorage

import (
	"context"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/service/s3"
)

const rustFSPresignExpiry = 5 * time.Minute

// RustFSSignedURLIssuer implements application/usecase/image_file.ISignedURLIssuer against a RustFS
// (S3-compatible) endpoint for local development, issuing a presigned PUT URL per file name.
type RustFSSignedURLIssuer struct {
	presignClient *s3.PresignClient
	bucket        string
}

func NewRustFSSignedURLIssuer(endpoint, accessKey, secretKey, bucket string) *RustFSSignedURLIssuer {
	client := s3.New(s3.Options{
		Region:       "us-east-1",
		BaseEndpoint: aws.String(endpoint),
		Credentials:  credentials.NewStaticCredentialsProvider(accessKey, secretKey, ""),
		UsePathStyle: true,
	})

	return &RustFSSignedURLIssuer{presignClient: s3.NewPresignClient(client), bucket: bucket}
}

func (i *RustFSSignedURLIssuer) IssueUploadURL(ctx context.Context, fileName, contentType string, sizeBytes int64) (string, error) {
	presigned, err := i.presignClient.PresignPutObject(ctx, &s3.PutObjectInput{
		Bucket:        aws.String(i.bucket),
		Key:           aws.String(fileName),
		ContentType:   aws.String(contentType),
		ContentLength: aws.Int64(sizeBytes),
	}, s3.WithPresignExpires(rustFSPresignExpiry))
	if err != nil {
		return "", err
	}

	return presigned.URL, nil
}
