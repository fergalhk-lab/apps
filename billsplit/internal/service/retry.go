// billsplit/internal/service/retry.go
package service

import (
	"context"
	"errors"

	"github.com/fergalhk-lab/apps/billsplit/internal/store"
	"go.uber.org/zap"
)

// withRetry reads key, applies mutate to the current data, then writes back
// using a conditional ETag. Retries once on ErrConflict.
// If the key does not exist, mutate receives nil data and the write uses
// create-if-not-exists semantics (empty ETag).
func withRetry(ctx context.Context, s store.Store, key string, logger *zap.Logger, mutate func(data []byte) ([]byte, error)) error {
	for attempt := 0; attempt < 2; attempt++ {
		data, etag, err := s.ReadObject(ctx, key)
		if errors.Is(err, store.ErrNotFound) {
			data, etag = nil, ""
		} else if err != nil {
			return err
		}

		newData, err := mutate(data)
		if err != nil {
			return err
		}

		if writeErr := s.WriteObject(ctx, key, newData, etag); writeErr == nil {
			return nil
		} else if errors.Is(writeErr, store.ErrConflict) && attempt == 0 {
			logger.Warn("write conflict, retrying", zap.String("key", key))
			continue
		} else {
			return writeErr
		}
	}
	return store.ErrConflict
}
