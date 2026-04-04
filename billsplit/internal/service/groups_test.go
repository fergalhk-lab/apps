// billsplit/internal/service/groups_test.go
package service_test

import (
	"context"
	"encoding/json"
	"testing"

	"github.com/fergalhk-lab/apps/billsplit/internal/domain"
	"github.com/fergalhk-lab/apps/billsplit/internal/service"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"go.uber.org/zap/zaptest"
	"go.uber.org/zap/zaptest/observer"
)

func setupGroupTest(t *testing.T) (*service.AuthService, *service.InviteService, *service.GroupService) {
	t.Helper()
	st := newTestStore(t)
	auth := service.NewAuthService(st, "secret", zaptest.NewLogger(t))
	invites := service.NewInviteService(st, zaptest.NewLogger(t))
	groups := service.NewGroupService(st, zaptest.NewLogger(t))
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

func TestGetGroup_IncludesSettlements(t *testing.T) {
	// Use a single shared store so all services see the same state.
	st := newTestStore(t)
	ctx := context.Background()
	auth := service.NewAuthService(st, "secret", zaptest.NewLogger(t))
	invites := service.NewInviteService(st, zaptest.NewLogger(t))
	groups := service.NewGroupService(st, zaptest.NewLogger(t))
	expenses := service.NewExpenseService(st, zaptest.NewLogger(t))

	codeA, _ := invites.GenerateInvite(ctx, false)
	_ = auth.Register(ctx, "alice", "pw", codeA)
	codeB, _ := invites.GenerateInvite(ctx, false)
	_ = auth.Register(ctx, "bob", "pw", codeB)

	gid, err := groups.CreateGroup(ctx, "alice", "Trip", "EUR", []string{"bob"})
	require.NoError(t, err, "create group")

	// alice pays 100, split evenly → alice net +50, bob net -50
	_, err = expenses.AddExpense(ctx, gid, "alice", "Dinner", "alice", 100.0, map[string]float64{
		"alice": 50.0,
		"bob":   50.0,
	}, nil)
	require.NoError(t, err, "add expense")

	detail, err := groups.GetGroup(ctx, gid)
	require.NoError(t, err, "get group")
	require.Len(t, detail.Settlements, 1, "want 1 settlement, got %d: %v", len(detail.Settlements), detail.Settlements)
	s := detail.Settlements[0]
	assert.Equal(t, "bob", s.From, "unexpected settlement From: %+v", s)
	assert.Equal(t, "alice", s.To, "unexpected settlement To: %+v", s)
	assert.Equal(t, float64(50), s.Amount, "unexpected settlement Amount: %+v", s)
}

// TestListGroups_WarnsOnMissingGroup verifies that ListGroups emits a Warn log
// when a group ID present in users.json has no corresponding group object in S3.
func TestListGroups_WarnsOnMissingGroup(t *testing.T) {
	core, logs := observer.New(zapcore.WarnLevel)
	logger := zap.New(core)

	st := newTestStore(t)
	ctx := context.Background()

	// Write a users.json where alice holds a group ID that has no group object in S3.
	ud := domain.UsersData{
		Users: []domain.User{
			{Username: "alice", PasswordHash: "x", GroupIDs: []string{"nonexistent-id"}, IsAdmin: false},
		},
		Invites: []domain.Invite{},
	}
	data, err := json.Marshal(ud)
	require.NoError(t, err)
	require.NoError(t, st.ForceWriteObject(ctx, "users.json", data))

	groups := service.NewGroupService(st, logger)
	list, err := groups.ListGroups(ctx, "alice")
	require.NoError(t, err, "ListGroups should not return an error when a group is missing")
	require.Empty(t, list, "missing group should be skipped, not returned")

	require.Equal(t, 1, logs.Len(), "expected exactly 1 warn log for the skipped group")
	entry := logs.All()[0]
	require.Equal(t, "group not found, skipping", entry.Message)
	require.Equal(t, "nonexistent-id", entry.ContextMap()["group_id"],
		"log should identify the missing group by ID")
}
