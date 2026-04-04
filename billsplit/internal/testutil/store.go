package testutil

import (
	"context"
	"testing"

	"github.com/aws/aws-sdk-go-v2/aws"
	awsconfig "github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	localstore "github.com/fergalhk-lab/apps/billsplit/internal/store"
)

// NewTestStore starts a MinIO container and returns a Store backed by it.
// Cleanup is registered with t.
func NewTestStore(t *testing.T) localstore.Store {
	t.Helper()
	minio := StartMinIO(t)
	ctx := context.Background()

	cfg, err := awsconfig.LoadDefaultConfig(ctx,
		awsconfig.WithRegion("us-east-1"),
		awsconfig.WithCredentialsProvider(aws.CredentialsProviderFunc(
			func(ctx context.Context) (aws.Credentials, error) {
				return aws.Credentials{
					AccessKeyID:     minio.AccessKey,
					SecretAccessKey: minio.SecretKey,
				}, nil
			},
		)),
		awsconfig.WithEndpointResolverWithOptions(aws.EndpointResolverWithOptionsFunc(
			func(svc, region string, _ ...interface{}) (aws.Endpoint, error) {
				return aws.Endpoint{URL: minio.Endpoint, HostnameImmutable: true}, nil
			},
		)),
	)
	if err != nil {
		t.Fatalf("aws config: %v", err)
	}

	client := s3.NewFromConfig(cfg, func(o *s3.Options) { o.UsePathStyle = true })
	bucket := "test-bucket"
	if _, err := client.CreateBucket(ctx, &s3.CreateBucketInput{Bucket: aws.String(bucket)}); err != nil {
		t.Fatalf("create bucket: %v", err)
	}
	return localstore.NewS3Store(client, bucket)
}
