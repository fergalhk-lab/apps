// billsplit/internal/service/invites_test.go
package service_test

import (
	"context"
	"testing"

	"github.com/fergalhk-lab/apps/billsplit/internal/service"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap/zaptest"
)

func TestGenerateInvite(t *testing.T) {
	st := newTestStore(t)
	auth := service.NewAuthService(st, "secret", zaptest.NewLogger(t))
	invites := service.NewInviteService(st, zaptest.NewLogger(t))
	ctx := context.Background()

	// Generate invite, use it to register
	code, err := invites.GenerateInvite(ctx, false)
	require.NoError(t, err, "generate invite: %v", err)
	require.NotEmpty(t, code, "expected non-empty code")

	// New user can register with generated code
	err = auth.Register(ctx, "bob", "pw", code)
	require.NoError(t, err, "register with generated invite: %v", err)
}

func TestHasInvites(t *testing.T) {
	st := newTestStore(t)
	invites := service.NewInviteService(st, zaptest.NewLogger(t))
	ctx := context.Background()

	// No invites on a fresh store
	has, err := invites.HasInvites(ctx)
	require.NoError(t, err, "HasInvites on empty store: %v", err)
	require.False(t, has, "expected false on empty store, got true")

	// Generate one invite
	_, err = invites.GenerateInvite(ctx, false)
	require.NoError(t, err, "GenerateInvite: %v", err)

	// Now should have invites
	has, err = invites.HasInvites(ctx)
	require.NoError(t, err, "HasInvites after generate: %v", err)
	require.True(t, has, "expected true after generating an invite, got false")
}
