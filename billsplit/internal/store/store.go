package store

import (
	"context"
	"errors"
)

var (
	// ErrConflict is returned when a conditional write fails because the
	// object was modified since it was last read.
	ErrConflict = errors.New("conditional write conflict")

	// ErrNotFound is returned when the requested key does not exist.
	ErrNotFound = errors.New("object not found")
)

// Store abstracts S3 object storage. All business logic depends on this
// interface; no AWS SDK types leak into callers.
type Store interface {
	// ReadObject returns the object's data and ETag. Returns ErrNotFound if
	// the key does not exist.
	ReadObject(ctx context.Context, key string) (data []byte, etag string, err error)

	// WriteObject writes data to key.
	// If ifMatchETag is "", uses create-if-not-exists semantics (IfNoneMatch: *).
	// If ifMatchETag is non-empty, only writes if the current ETag matches.
	// Returns ErrConflict if the condition fails.
	WriteObject(ctx context.Context, key string, data []byte, ifMatchETag string) error
}
