// billsplit/internal/service/settlements_test.go
package service_test

import (
	"context"
	"testing"

	"github.com/fergalhk-lab/apps/billsplit/internal/service"
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
	if err != nil {
		t.Fatalf("add settlement: %v", err)
	}
}

func TestAddSettlement_GroupNotFound(t *testing.T) {
	_, _, _, settlements := setupSettlementTest(t)
	ctx := context.Background()

	err := settlements.AddSettlement(ctx, "nonexistent-group", "bob", "bob", "alice", 50.0)
	if err == nil {
		t.Fatal("expected error for nonexistent group")
	}
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

	if err := settlements.AddSettlement(ctx, groupID, "bob", "bob", "alice", 50.0); err != nil {
		t.Fatalf("add settlement: %v", err)
	}

	events, total, err := expenses.ListEvents(ctx, groupID, 10, 0)
	if err != nil {
		t.Fatalf("list events: %v", err)
	}
	if total != 1 {
		t.Fatalf("expected total=1, got %d", total)
	}
	if len(events) != 1 {
		t.Fatalf("expected 1 event, got %d", len(events))
	}
	if events[0].From != "bob" || events[0].To != "alice" {
		t.Errorf("unexpected settlement event: %+v", events[0])
	}
}
