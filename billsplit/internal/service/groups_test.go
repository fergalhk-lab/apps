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
