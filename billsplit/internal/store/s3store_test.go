package store_test

import (
	"bytes"
	"context"
	"testing"

	"github.com/aws/aws-sdk-go-v2/aws"
	awsconfig "github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	localstore "github.com/fergalhk-lab/apps/billsplit/internal/store"
	"github.com/fergalhk-lab/apps/billsplit/internal/testutil"
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

func TestReadObject_NotFound(t *testing.T) {
	s := newTestStore(t)
	_, _, err := s.ReadObject(context.Background(), "missing.json")
	if err != localstore.ErrNotFound {
		t.Fatalf("expected ErrNotFound, got %v", err)
	}
}

func TestWriteAndRead(t *testing.T) {
	s := newTestStore(t)
	ctx := context.Background()
	data := []byte(`{"hello":"world"}`)

	if err := s.WriteObject(ctx, "test.json", data, ""); err != nil {
		t.Fatalf("write: %v", err)
	}

	got, etag, err := s.ReadObject(ctx, "test.json")
	if err != nil {
		t.Fatalf("read: %v", err)
	}
	if !bytes.Equal(got, data) {
		t.Fatalf("got %s, want %s", got, data)
	}
	if etag == "" {
		t.Fatal("expected non-empty etag")
	}
}

func TestConditionalUpdate(t *testing.T) {
	s := newTestStore(t)
	ctx := context.Background()

	if err := s.WriteObject(ctx, "obj.json", []byte(`{"v":1}`), ""); err != nil {
		t.Fatalf("initial write: %v", err)
	}
	_, etag, _ := s.ReadObject(ctx, "obj.json")

	if err := s.WriteObject(ctx, "obj.json", []byte(`{"v":2}`), etag); err != nil {
		t.Fatalf("conditional update: %v", err)
	}
}

func TestConflictOnStaleETag(t *testing.T) {
	s := newTestStore(t)
	ctx := context.Background()

	if err := s.WriteObject(ctx, "obj.json", []byte(`{"v":1}`), ""); err != nil {
		t.Fatalf("write: %v", err)
	}
	_, staleETag, _ := s.ReadObject(ctx, "obj.json")

	// overwrite to advance the ETag
	_, curETag, _ := s.ReadObject(ctx, "obj.json")
	if err := s.WriteObject(ctx, "obj.json", []byte(`{"v":2}`), curETag); err != nil {
		t.Fatalf("second write: %v", err)
	}

	// staleETag no longer matches
	err := s.WriteObject(ctx, "obj.json", []byte(`{"v":3}`), staleETag)
	if err != localstore.ErrConflict {
		t.Fatalf("expected ErrConflict, got %v", err)
	}
}

func TestForceWriteObject(t *testing.T) {
	s := newTestStore(t)
	ctx := context.Background()

	// Works when object does not exist.
	if err := s.ForceWriteObject(ctx, "obj.json", []byte(`{"v":1}`)); err != nil {
		t.Fatalf("force write (create): %v", err)
	}

	// Works when object already exists — overwrites without conflict.
	if err := s.ForceWriteObject(ctx, "obj.json", []byte(`{"v":2}`)); err != nil {
		t.Fatalf("force write (overwrite): %v", err)
	}

	got, _, err := s.ReadObject(ctx, "obj.json")
	if err != nil {
		t.Fatalf("read after force write: %v", err)
	}
	if string(got) != `{"v":2}` {
		t.Fatalf("got %s, want {\"v\":2}", got)
	}
}

func TestCreateIfNotExists_FailsIfExists(t *testing.T) {
	s := newTestStore(t)
	ctx := context.Background()

	if err := s.WriteObject(ctx, "obj.json", []byte(`{"v":1}`), ""); err != nil {
		t.Fatalf("first write: %v", err)
	}

	err := s.WriteObject(ctx, "obj.json", []byte(`{"v":2}`), "")
	if err != localstore.ErrConflict {
		t.Fatalf("expected ErrConflict, got %v", err)
	}
}
