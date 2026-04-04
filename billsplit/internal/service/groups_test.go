// billsplit/internal/service/groups_test.go
package service_test

import (
	"context"
	"testing"

	"github.com/fergalhk-lab/apps/billsplit/internal/service"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
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
	require.NoError(t, err, "create group: %v", err)
	require.NotEmpty(t, id, "expected non-empty group ID")
}

func TestCreateGroup_UnknownMember(t *testing.T) {
	auth, invites, groups := setupGroupTest(t)
	ctx := context.Background()

	code, _ := invites.GenerateInvite(ctx, false)
	_ = auth.Register(ctx, "alice", "pw", code)

	_, err := groups.CreateGroup(ctx, "alice", "Trip", "EUR", []string{"bob"})
	require.Error(t, err, "expected error for unknown member")
}

func TestCreateGroup_DuplicateInMembers(t *testing.T) {
	auth, invites, groups := setupGroupTest(t)
	ctx := context.Background()

	code, _ := invites.GenerateInvite(ctx, false)
	_ = auth.Register(ctx, "alice", "pw", code)

	_, err := groups.CreateGroup(ctx, "alice", "Trip", "EUR", []string{"bob", "bob"})
	require.ErrorIs(t, err, service.ErrDuplicateMembers)
}

func TestCreateGroup_CreatorListedInMembers(t *testing.T) {
	auth, invites, groups := setupGroupTest(t)
	ctx := context.Background()

	code, _ := invites.GenerateInvite(ctx, false)
	_ = auth.Register(ctx, "alice", "pw", code)

	_, err := groups.CreateGroup(ctx, "alice", "Trip", "EUR", []string{"alice"})
	require.ErrorIs(t, err, service.ErrDuplicateMembers)
}

func TestListGroups(t *testing.T) {
	auth, invites, groups := setupGroupTest(t)
	ctx := context.Background()

	code, _ := invites.GenerateInvite(ctx, false)
	_ = auth.Register(ctx, "alice", "pw", code)

	id, _ := groups.CreateGroup(ctx, "alice", "Barcelona", "EUR", []string{})
	list, err := groups.ListGroups(ctx, "alice")
	require.NoError(t, err, "list: %v", err)
	require.Len(t, list, 1)
	assert.Equal(t, id, list[0].ID, "expected 1 group with id %s, got %v", id, list)
}

func TestGetGroup(t *testing.T) {
	auth, invites, groups := setupGroupTest(t)
	ctx := context.Background()

	code, _ := invites.GenerateInvite(ctx, false)
	_ = auth.Register(ctx, "alice", "pw", code)

	id, _ := groups.CreateGroup(ctx, "alice", "Barcelona", "EUR", []string{})
	detail, err := groups.GetGroup(ctx, id)
	require.NoError(t, err, "get: %v", err)
	assert.Equal(t, "Barcelona", detail.Name)
	require.Len(t, detail.Members, 1)
	assert.Equal(t, "alice", detail.Members[0], "unexpected members: %v", detail.Members)
}
