// billsplit/internal/service/expenses_test.go
package service_test

import (
	"context"
	"testing"

	"github.com/fergalhk-lab/apps/billsplit/internal/domain"
	"github.com/fergalhk-lab/apps/billsplit/internal/service"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap/zaptest"
)

func setupExpenseTest(t *testing.T) (*service.AuthService, *service.InviteService, *service.GroupService, *service.ExpenseService) {
	t.Helper()
	st := newTestStore(t)
	auth := service.NewAuthService(st, "secret", zaptest.NewLogger(t))
	invites := service.NewInviteService(st, zaptest.NewLogger(t))
	groups := service.NewGroupService(st, zaptest.NewLogger(t))
	expenses := service.NewExpenseService(st, zaptest.NewLogger(t))
	return auth, invites, groups, expenses
}

func registerAndCreateGroup(t *testing.T, auth *service.AuthService, invites *service.InviteService, groups *service.GroupService) (groupID string) {
	t.Helper()
	ctx := context.Background()

	codeA, err := invites.GenerateInvite(ctx, false)
	require.NoError(t, err, "generate invite alice: %v", err)
	err = auth.Register(ctx, "alice", "pw", codeA)
	require.NoError(t, err, "register alice: %v", err)

	codeB, err := invites.GenerateInvite(ctx, false)
	require.NoError(t, err, "generate invite bob: %v", err)
	err = auth.Register(ctx, "bob", "pw", codeB)
	require.NoError(t, err, "register bob: %v", err)

	groupID, err = groups.CreateGroup(ctx, "alice", "Trip", "EUR", []string{"bob"})
	require.NoError(t, err, "create group: %v", err)
	return groupID
}

func TestAddExpense(t *testing.T) {
	auth, invites, groups, expenses := setupExpenseTest(t)
	ctx := context.Background()
	groupID := registerAndCreateGroup(t, auth, invites, groups)

	eventID, err := expenses.AddExpense(ctx, groupID, "alice", "Dinner", "alice", 10000, map[string]int64{
		"alice": 5000,
		"bob":   5000,
	}, nil)
	require.NoError(t, err, "add expense: %v", err)
	require.NotEmpty(t, eventID, "expected non-empty event ID")
}

func TestAddExpense_InvalidSplits(t *testing.T) {
	auth, invites, groups, expenses := setupExpenseTest(t)
	ctx := context.Background()
	groupID := registerAndCreateGroup(t, auth, invites, groups)

	_, err := expenses.AddExpense(ctx, groupID, "alice", "Dinner", "alice", 10000, map[string]int64{
		"alice": 4000,
		"bob":   4000,
	}, nil)
	require.Error(t, err, "expected error for invalid splits")
}

func TestAddExpense_UnknownMemberInSplits(t *testing.T) {
	auth, invites, groups, expenses := setupExpenseTest(t)
	ctx := context.Background()
	groupID := registerAndCreateGroup(t, auth, invites, groups)

	_, err := expenses.AddExpense(ctx, groupID, "alice", "Dinner", "alice", 10000, map[string]int64{
		"alice":  5000,
		"nobody": 5000,
	}, nil)
	require.Error(t, err, "expected error for unknown split member")
}

func TestCancelExpense(t *testing.T) {
	auth, invites, groups, expenses := setupExpenseTest(t)
	ctx := context.Background()
	groupID := registerAndCreateGroup(t, auth, invites, groups)

	eventID, err := expenses.AddExpense(ctx, groupID, "alice", "Dinner", "alice", 10000, map[string]int64{
		"alice": 5000,
		"bob":   5000,
	}, nil)
	require.NoError(t, err, "add expense: %v", err)

	err = expenses.CancelExpense(ctx, groupID, "alice", eventID)
	require.NoError(t, err, "cancel expense: %v", err)
}

func TestCancelExpense_AlreadyCancelled(t *testing.T) {
	auth, invites, groups, expenses := setupExpenseTest(t)
	ctx := context.Background()
	groupID := registerAndCreateGroup(t, auth, invites, groups)

	eventID, _ := expenses.AddExpense(ctx, groupID, "alice", "Dinner", "alice", 10000, map[string]int64{
		"alice": 5000,
		"bob":   5000,
	}, nil)
	_ = expenses.CancelExpense(ctx, groupID, "alice", eventID)

	err := expenses.CancelExpense(ctx, groupID, "alice", eventID)
	require.Error(t, err, "expected error cancelling already-cancelled expense")
}

func TestCancelExpense_NotFound(t *testing.T) {
	auth, invites, groups, expenses := setupExpenseTest(t)
	ctx := context.Background()
	groupID := registerAndCreateGroup(t, auth, invites, groups)

	err := expenses.CancelExpense(ctx, groupID, "alice", "nonexistent-id")
	require.ErrorIs(t, err, service.ErrEventNotFound)
}

func TestListEvents_NewestFirst(t *testing.T) {
	auth, invites, groups, expenses := setupExpenseTest(t)
	ctx := context.Background()
	groupID := registerAndCreateGroup(t, auth, invites, groups)

	id1, _ := expenses.AddExpense(ctx, groupID, "alice", "First", "alice", 6000, map[string]int64{
		"alice": 3000,
		"bob":   3000,
	}, nil)
	id2, _ := expenses.AddExpense(ctx, groupID, "alice", "Second", "alice", 4000, map[string]int64{
		"alice": 2000,
		"bob":   2000,
	}, nil)

	events, total, err := expenses.ListEvents(ctx, groupID, 10, 0)
	require.NoError(t, err, "list events: %v", err)
	require.Equal(t, 2, total, "expected total=2, got %d", total)
	// newest-first: second expense should come before first
	assert.Equal(t, id2, events[0].ID, "expected first result to be %s (second expense), got %s", id2, events[0].ID)
	assert.Equal(t, id1, events[1].ID, "expected second result to be %s (first expense), got %s", id1, events[1].ID)
}

func TestListEvents_Pagination(t *testing.T) {
	auth, invites, groups, expenses := setupExpenseTest(t)
	ctx := context.Background()
	groupID := registerAndCreateGroup(t, auth, invites, groups)

	for i := 0; i < 5; i++ {
		_, err := expenses.AddExpense(ctx, groupID, "alice", "Expense", "alice", 2000, map[string]int64{
			"alice": 1000,
			"bob":   1000,
		}, nil)
		require.NoError(t, err, "add expense %d: %v", i, err)
	}

	page1, total, err := expenses.ListEvents(ctx, groupID, 2, 0)
	require.NoError(t, err, "list page 1: %v", err)
	require.Equal(t, 5, total, "expected total=5, got %d", total)
	require.Len(t, page1, 2, "expected 2 events on page 1, got %d", len(page1))

	page3, _, err := expenses.ListEvents(ctx, groupID, 2, 4)
	require.NoError(t, err, "list page 3: %v", err)
	require.Len(t, page3, 1, "expected 1 event on last page, got %d", len(page3))

	empty, total2, err := expenses.ListEvents(ctx, groupID, 2, 10)
	require.NoError(t, err, "list beyond end: %v", err)
	require.Equal(t, 5, total2, "expected total=5, got %d", total2)
	require.Empty(t, empty, "expected empty slice, got %d events", len(empty))
}

func TestGetGroupCurrency(t *testing.T) {
	auth, invites, groups, expenses := setupExpenseTest(t)
	ctx := context.Background()
	groupID := registerAndCreateGroup(t, auth, invites, groups)

	currency, err := expenses.GetGroupCurrency(ctx, groupID)
	require.NoError(t, err)
	require.Equal(t, "EUR", currency)
}

func TestAddExpense_WithOriginalExpense(t *testing.T) {
	auth, invites, groups, expenses := setupExpenseTest(t)
	ctx := context.Background()
	groupID := registerAndCreateGroup(t, auth, invites, groups)

	orig := &domain.OriginalExpense{Currency: "GBP", Amount: 4500}
	eventID, err := expenses.AddExpense(ctx, groupID, "alice", "Dinner", "alice", 5000, map[string]int64{
		"alice": 2500,
		"bob":   2500,
	}, orig)
	require.NoError(t, err)
	require.NotEmpty(t, eventID)

	events, _, err := expenses.ListEvents(ctx, groupID, 10, 0)
	require.NoError(t, err)
	require.Len(t, events, 1)
	require.NotNil(t, events[0].OriginalExpense)
	require.Equal(t, "GBP", events[0].OriginalExpense.Currency)
	require.Equal(t, int64(4500), events[0].OriginalExpense.Amount)
}
