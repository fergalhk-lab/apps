// billsplit/internal/service/settlements_test.go
package service_test

import (
	"context"
	"testing"

	"github.com/fergalhk-lab/apps/billsplit/internal/service"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func setupSettlementTest(t *testing.T) (*service.AuthService, *service.InviteService, *service.GroupService, *service.SettlementService) {
	t.Helper()
	st := newTestStore(t)
	auth := service.NewAuthService(st, "secret")
	invites := service.NewInviteService(st)
	groups := service.NewGroupService(st)
	settlements := service.NewSettlementService(st)
	return auth, invites, groups, settlements
}

func TestAddSettlement(t *testing.T) {
	auth, invites, groups, settlements := setupSettlementTest(t)
	ctx := context.Background()
	groupID := registerAndCreateGroup(t, auth, invites, groups)

	err := settlements.AddSettlement(ctx, groupID, "bob", "bob", "alice", 50.0)
	require.NoError(t, err, "add settlement: %v", err)
}

func TestAddSettlement_GroupNotFound(t *testing.T) {
	_, _, _, settlements := setupSettlementTest(t)
	ctx := context.Background()

	err := settlements.AddSettlement(ctx, "nonexistent-group", "bob", "bob", "alice", 50.0)
	require.Error(t, err, "expected error for nonexistent group")
}

func TestAddSettlement_AppearsInListEvents(t *testing.T) {
	st := newTestStore(t)
	ctx := context.Background()
	auth := service.NewAuthService(st, "secret")
	invites := service.NewInviteService(st)
	groups := service.NewGroupService(st)
	settlements := service.NewSettlementService(st)
	expenses := service.NewExpenseService(st)

	groupID := registerAndCreateGroup(t, auth, invites, groups)

	err := settlements.AddSettlement(ctx, groupID, "bob", "bob", "alice", 50.0)
	require.NoError(t, err, "add settlement: %v", err)

	events, total, err := expenses.ListEvents(ctx, groupID, 10, 0)
	require.NoError(t, err, "list events: %v", err)
	require.Equal(t, 1, total, "expected total=1, got %d", total)
	require.Len(t, events, 1, "expected 1 event, got %d", len(events))
	assert.Equal(t, "bob", events[0].From, "unexpected settlement event: %+v", events[0])
	assert.Equal(t, "alice", events[0].To, "unexpected settlement event: %+v", events[0])
}
