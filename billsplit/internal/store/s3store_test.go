package store_test

import (
	"context"
	"testing"

	"github.com/aws/aws-sdk-go-v2/aws"
	awsconfig "github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	localstore "github.com/fergalhk-lab/apps/billsplit/internal/store"
	"github.com/fergalhk-lab/apps/billsplit/internal/testutil"
	"github.com/stretchr/testify/require"
)

func newTestStore(t *testing.T) localstore.Store {
	t.Helper()
	minio := testutil.StartMinIO(t)
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
			func(service, region string, _ ...interface{}) (aws.Endpoint, error) {
				return aws.Endpoint{URL: minio.Endpoint, HostnameImmutable: true}, nil
			},
		)),
	)
	require.NoError(t, err, "aws config: %v", err)

	client := s3.NewFromConfig(cfg, func(o *s3.Options) { o.UsePathStyle = true })
	bucket := "test-bucket"
	_, err = client.CreateBucket(ctx, &s3.CreateBucketInput{Bucket: aws.String(bucket)})
	require.NoError(t, err, "create bucket: %v", err)

	return localstore.NewS3Store(client, bucket)
}

func TestReadObject_NotFound(t *testing.T) {
	s := newTestStore(t)
	_, _, err := s.ReadObject(context.Background(), "missing.json")
	require.ErrorIs(t, err, localstore.ErrNotFound)
}

func TestWriteAndRead(t *testing.T) {
	s := newTestStore(t)
	ctx := context.Background()
	data := []byte(`{"hello":"world"}`)

	err := s.WriteObject(ctx, "test.json", data, "")
	require.NoError(t, err, "write: %v", err)

	got, etag, err := s.ReadObject(ctx, "test.json")
	require.NoError(t, err, "read: %v", err)
	require.Equal(t, data, got)
	require.NotEmpty(t, etag, "expected non-empty etag")
}

func TestConditionalUpdate(t *testing.T) {
	s := newTestStore(t)
	ctx := context.Background()

	err := s.WriteObject(ctx, "obj.json", []byte(`{"v":1}`), "")
	require.NoError(t, err, "initial write: %v", err)
	_, etag, _ := s.ReadObject(ctx, "obj.json")

	err = s.WriteObject(ctx, "obj.json", []byte(`{"v":2}`), etag)
	require.NoError(t, err, "conditional update: %v", err)
}

func TestConflictOnStaleETag(t *testing.T) {
	s := newTestStore(t)
	ctx := context.Background()

	err := s.WriteObject(ctx, "obj.json", []byte(`{"v":1}`), "")
	require.NoError(t, err, "write: %v", err)
	_, staleETag, _ := s.ReadObject(ctx, "obj.json")

	// overwrite to advance the ETag
	_, curETag, _ := s.ReadObject(ctx, "obj.json")
	err = s.WriteObject(ctx, "obj.json", []byte(`{"v":2}`), curETag)
	require.NoError(t, err, "second write: %v", err)

	// staleETag no longer matches
	err = s.WriteObject(ctx, "obj.json", []byte(`{"v":3}`), staleETag)
	require.ErrorIs(t, err, localstore.ErrConflict)
}

func TestForceWriteObject(t *testing.T) {
	s := newTestStore(t)
	ctx := context.Background()

	// Works when object does not exist.
	err := s.ForceWriteObject(ctx, "obj.json", []byte(`{"v":1}`))
	require.NoError(t, err, "force write (create): %v", err)

	// Works when object already exists — overwrites without conflict.
	err = s.ForceWriteObject(ctx, "obj.json", []byte(`{"v":2}`))
	require.NoError(t, err, "force write (overwrite): %v", err)

	got, _, err := s.ReadObject(ctx, "obj.json")
	require.NoError(t, err, "read after force write: %v", err)
	require.Equal(t, `{"v":2}`, string(got))
}

func TestCreateIfNotExists_FailsIfExists(t *testing.T) {
	s := newTestStore(t)
	ctx := context.Background()

	err := s.WriteObject(ctx, "obj.json", []byte(`{"v":1}`), "")
	require.NoError(t, err, "first write: %v", err)

	err = s.WriteObject(ctx, "obj.json", []byte(`{"v":2}`), "")
	require.ErrorIs(t, err, localstore.ErrConflict)
}
