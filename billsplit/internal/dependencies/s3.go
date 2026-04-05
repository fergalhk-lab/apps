package dependencies

import (
	"context"
	"os"

	awsconfig "github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/s3"
)

// NewS3Client creates an S3 client from the default AWS config.
// When AWS_ENDPOINT_URL_S3 is set, path-style addressing is enabled
// (required for MinIO and other S3-compatible services).
func NewS3Client(ctx context.Context) (*s3.Client, error) {
	awsCfg, err := awsconfig.LoadDefaultConfig(ctx)
	if err != nil {
		return nil, err
	}
	return s3.NewFromConfig(awsCfg, func(o *s3.Options) {
		if os.Getenv("AWS_ENDPOINT_URL_S3") != "" {
			o.UsePathStyle = true
		}
	}), nil
}
