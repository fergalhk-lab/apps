// billsplit/internal/service/groups_test.go
package service_test

import (
	"context"
	"errors"
	"testing"

	"github.com/fergalhk-lab/apps/billsplit/internal/service"
)

func setupGroupTest(t *testing.T) (*service.AuthService, *service.InviteService, *service.GroupService) {
	t.Helper()
	st := newTestStore(t)
	auth := service.NewAuthService(st, "secret")
	invites := service.NewInviteService(st)
	groups := service.NewGroupService(st)
	return auth, invites, groups
}

func TestCreateGroup(t *testing.T) {
	auth, invites, groups := setupGroupTest(t)
	ctx := context.Background()

	code, _ := invites.GenerateInvite(ctx, false)
	_ = auth.Register(ctx, "alice", "pw", code)

	id, err := groups.CreateGroup(ctx, "alice", "Barcelona", "EUR", []string{})
	if err != nil {
		t.Fatalf("create group: %v", err)
	}
	if id == "" {
		t.Fatal("expected non-empty group ID")
	}
}

func TestCreateGroup_UnknownMember(t *testing.T) {
	auth, invites, groups := setupGroupTest(t)
	ctx := context.Background()

	code, _ := invites.GenerateInvite(ctx, false)
	_ = auth.Register(ctx, "alice", "pw", code)

	_, err := groups.CreateGroup(ctx, "alice", "Trip", "EUR", []string{"bob"})
	if err == nil {
		t.Fatal("expected error for unknown member")
	}
}

func TestCreateGroup_DuplicateInMembers(t *testing.T) {
	auth, invites, groups := setupGroupTest(t)
	ctx := context.Background()

	code, _ := invites.GenerateInvite(ctx, false)
	_ = auth.Register(ctx, "alice", "pw", code)

	_, err := groups.CreateGroup(ctx, "alice", "Trip", "EUR", []string{"bob", "bob"})
	if !errors.Is(err, service.ErrDuplicateMembers) {
		t.Fatalf("expected ErrDuplicateMembers, got %v", err)
	}
}

func TestCreateGroup_CreatorListedInMembers(t *testing.T) {
	auth, invites, groups := setupGroupTest(t)
	ctx := context.Background()

	code, _ := invites.GenerateInvite(ctx, false)
	_ = auth.Register(ctx, "alice", "pw", code)

	_, err := groups.CreateGroup(ctx, "alice", "Trip", "EUR", []string{"alice"})
	if !errors.Is(err, service.ErrDuplicateMembers) {
		t.Fatalf("expected ErrDuplicateMembers, got %v", err)
	}
}

func TestListGroups(t *testing.T) {
	auth, invites, groups := setupGroupTest(t)
	ctx := context.Background()

	code, _ := invites.GenerateInvite(ctx, false)
	_ = auth.Register(ctx, "alice", "pw", code)

	id, _ := groups.CreateGroup(ctx, "alice", "Barcelona", "EUR", []string{})
	list, err := groups.ListGroups(ctx, "alice")
	if err != nil {
		t.Fatalf("list: %v", err)
	}
	if len(list) != 1 || list[0].ID != id {
		t.Errorf("expected 1 group with id %s, got %v", id, list)
	}
}

func TestGetGroup(t *testing.T) {
	auth, invites, groups := setupGroupTest(t)
	ctx := context.Background()

	code, _ := invites.GenerateInvite(ctx, false)
	_ = auth.Register(ctx, "alice", "pw", code)

	id, _ := groups.CreateGroup(ctx, "alice", "Barcelona", "EUR", []string{})
	detail, err := groups.GetGroup(ctx, id)
	if err != nil {
		t.Fatalf("get: %v", err)
	}
	if detail.Name != "Barcelona" {
		t.Errorf("want Barcelona, got %s", detail.Name)
	}
	if len(detail.Members) != 1 || detail.Members[0] != "alice" {
		t.Errorf("unexpected members: %v", detail.Members)
	}
}

func TestGetGroup_IncludesSettlements(t *testing.T) {
	// Use a single shared store so all services see the same state.
	st := newTestStore(t)
	ctx := context.Background()
	auth := service.NewAuthService(st, "secret")
	invites := service.NewInviteService(st)
	groups := service.NewGroupService(st)
	expenses := service.NewExpenseService(st)

	codeA, _ := invites.GenerateInvite(ctx, false)
	_ = auth.Register(ctx, "alice", "pw", codeA)
	codeB, _ := invites.GenerateInvite(ctx, false)
	_ = auth.Register(ctx, "bob", "pw", codeB)

	gid, err := groups.CreateGroup(ctx, "alice", "Trip", "EUR", []string{"bob"})
	if err != nil {
		t.Fatalf("create group: %v", err)
	}

	// alice pays 100, split evenly → alice net +50, bob net -50
	_, err = expenses.AddExpense(ctx, gid, "alice", "Dinner", "alice", 100.0, map[string]float64{
		"alice": 50.0,
		"bob":   50.0,
	})
	if err != nil {
		t.Fatalf("add expense: %v", err)
	}

	detail, err := groups.GetGroup(ctx, gid)
	if err != nil {
		t.Fatalf("get group: %v", err)
	}
	if len(detail.Settlements) != 1 {
		t.Fatalf("want 1 settlement, got %d: %v", len(detail.Settlements), detail.Settlements)
	}
	s := detail.Settlements[0]
	if s.From != "bob" || s.To != "alice" || s.Amount != 50 {
		t.Errorf("unexpected settlement: %+v", s)
	}
}
