// billsplit/internal/service/retry_test.go
package service_test

import (
	"context"
	"testing"

	"github.com/fergalhk-lab/apps/billsplit/internal/service"
	"github.com/fergalhk-lab/apps/billsplit/internal/store"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"go.uber.org/zap/zaptest/observer"
)

// conflictOnceStore wraps a store.Store and returns ErrConflict on the first
// WriteObject call, then delegates subsequent calls to the underlying store.
type conflictOnceStore struct {
	store.Store
	written int
}

func (c *conflictOnceStore) WriteObject(ctx context.Context, key string, data []byte, etag string) error {
	c.written++
	if c.written == 1 {
		return store.ErrConflict
	}
	return c.Store.WriteObject(ctx, key, data, etag)
}

// TestWithRetry_LogsWarnOnConflict verifies that withRetry emits a Warn log
// when a conditional write fails on the first attempt and succeeds on retry.
func TestWithRetry_LogsWarnOnConflict(t *testing.T) {
	core, logs := observer.New(zapcore.WarnLevel)
	logger := zap.New(core)

	st := newTestStore(t)
	cs := &conflictOnceStore{Store: st}
	invites := service.NewInviteService(cs, logger)

	code, err := invites.GenerateInvite(context.Background(), false)
	require.NoError(t, err, "GenerateInvite should succeed after retry")
	require.NotEmpty(t, code)

	require.Equal(t, 1, logs.Len(), "expected exactly 1 warn log for the conflict retry")
	entry := logs.All()[0]
	require.Equal(t, "write conflict, retrying", entry.Message)
	require.Equal(t, "users.json", entry.ContextMap()["key"],
		"log should identify the conflicting key")
}
