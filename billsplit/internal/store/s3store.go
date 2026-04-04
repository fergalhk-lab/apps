package store

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"io"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/aws/aws-sdk-go-v2/service/s3/types"
	smithyhttp "github.com/aws/smithy-go/transport/http"
)

type S3Store struct {
	client *s3.Client
	bucket string
}

func NewS3Store(client *s3.Client, bucket string) *S3Store {
	return &S3Store{client: client, bucket: bucket}
}

func (s *S3Store) ReadObject(ctx context.Context, key string) ([]byte, string, error) {
	resp, err := s.client.GetObject(ctx, &s3.GetObjectInput{
		Bucket: aws.String(s.bucket),
		Key:    aws.String(key),
	})
	if err != nil {
		var noKey *types.NoSuchKey
		if errors.As(err, &noKey) {
			return nil, "", ErrNotFound
		}
		// MinIO returns a 404 generic error for missing keys
		var httpErr *smithyhttp.ResponseError
		if errors.As(err, &httpErr) && httpErr.HTTPStatusCode() == 404 {
			return nil, "", ErrNotFound
		}
		return nil, "", fmt.Errorf("s3 get %s: %w", key, err)
	}
	defer resp.Body.Close()

	data, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, "", fmt.Errorf("read body %s: %w", key, err)
	}

	etag := ""
	if resp.ETag != nil {
		etag = *resp.ETag
	}
	return data, etag, nil
}

func (s *S3Store) WriteObject(ctx context.Context, key string, data []byte, ifMatchETag string) error {
	input := &s3.PutObjectInput{
		Bucket:      aws.String(s.bucket),
		Key:         aws.String(key),
		Body:        bytes.NewReader(data),
		ContentType: aws.String("application/json"),
	}
	if ifMatchETag == "" {
		input.IfNoneMatch = aws.String("*")
	} else {
		input.IfMatch = aws.String(ifMatchETag)
	}

	_, err := s.client.PutObject(ctx, input)
	if err != nil {
		var httpErr *smithyhttp.ResponseError
		if errors.As(err, &httpErr) && httpErr.HTTPStatusCode() == 412 {
			return ErrConflict
		}
		return fmt.Errorf("s3 put %s: %w", key, err)
	}
	return nil
}

func (s *S3Store) ForceWriteObject(ctx context.Context, key string, data []byte) error {
	_, err := s.client.PutObject(ctx, &s3.PutObjectInput{
		Bucket:      aws.String(s.bucket),
		Key:         aws.String(key),
		Body:        bytes.NewReader(data),
		ContentType: aws.String("application/json"),
	})
	if err != nil {
		return fmt.Errorf("s3 force put %s: %w", key, err)
	}
	return nil
}
