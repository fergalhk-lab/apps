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
	auth := service.NewAuthService(st, "secret", testAuthExpiry, zaptest.NewLogger(t))
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

	eventID, err := expenses.AddExpense(ctx, groupID, "alice", "Dinner", "alice", 100.0, map[string]float64{
		"alice": 50.0,
		"bob":   50.0,
	}, nil)
	require.NoError(t, err, "add expense: %v", err)
	require.NotEmpty(t, eventID, "expected non-empty event ID")
}

func TestAddExpense_InvalidSplits(t *testing.T) {
	auth, invites, groups, expenses := setupExpenseTest(t)
	ctx := context.Background()
	groupID := registerAndCreateGroup(t, auth, invites, groups)

	_, err := expenses.AddExpense(ctx, groupID, "alice", "Dinner", "alice", 100.0, map[string]float64{
		"alice": 40.0,
		"bob":   40.0,
	}, nil)
	require.Error(t, err, "expected error for invalid splits")
}

func TestAddExpense_UnknownMemberInSplits(t *testing.T) {
	auth, invites, groups, expenses := setupExpenseTest(t)
	ctx := context.Background()
	groupID := registerAndCreateGroup(t, auth, invites, groups)

	_, err := expenses.AddExpense(ctx, groupID, "alice", "Dinner", "alice", 100.0, map[string]float64{
		"alice":  50.0,
		"nobody": 50.0,
	}, nil)
	require.Error(t, err, "expected error for unknown split member")
}

func TestCancelExpense(t *testing.T) {
	auth, invites, groups, expenses := setupExpenseTest(t)
	ctx := context.Background()
	groupID := registerAndCreateGroup(t, auth, invites, groups)

	eventID, err := expenses.AddExpense(ctx, groupID, "alice", "Dinner", "alice", 100.0, map[string]float64{
		"alice": 50.0,
		"bob":   50.0,
	}, nil)
	require.NoError(t, err, "add expense: %v", err)

	err = expenses.CancelExpense(ctx, groupID, "alice", eventID)
	require.NoError(t, err, "cancel expense: %v", err)
}

func TestCancelExpense_AlreadyCancelled(t *testing.T) {
	auth, invites, groups, expenses := setupExpenseTest(t)
	ctx := context.Background()
	groupID := registerAndCreateGroup(t, auth, invites, groups)

	eventID, _ := expenses.AddExpense(ctx, groupID, "alice", "Dinner", "alice", 100.0, map[string]float64{
		"alice": 50.0,
		"bob":   50.0,
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

	id1, _ := expenses.AddExpense(ctx, groupID, "alice", "First", "alice", 60.0, map[string]float64{
		"alice": 30.0,
		"bob":   30.0,
	}, nil)
	id2, _ := expenses.AddExpense(ctx, groupID, "alice", "Second", "alice", 40.0, map[string]float64{
		"alice": 20.0,
		"bob":   20.0,
	}, nil)

	events, total, err := expenses.ListEvents(ctx, groupID, 10, 0, false)
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
		_, err := expenses.AddExpense(ctx, groupID, "alice", "Expense", "alice", 20.0, map[string]float64{
			"alice": 10.0,
			"bob":   10.0,
		}, nil)
		require.NoError(t, err, "add expense %d: %v", i, err)
	}

	page1, total, err := expenses.ListEvents(ctx, groupID, 2, 0, false)
	require.NoError(t, err, "list page 1: %v", err)
	require.Equal(t, 5, total, "expected total=5, got %d", total)
	require.Len(t, page1, 2, "expected 2 events on page 1, got %d", len(page1))

	page3, _, err := expenses.ListEvents(ctx, groupID, 2, 4, false)
	require.NoError(t, err, "list page 3: %v", err)
	require.Len(t, page3, 1, "expected 1 event on last page, got %d", len(page3))

	empty, total2, err := expenses.ListEvents(ctx, groupID, 2, 10, false)
	require.NoError(t, err, "list beyond end: %v", err)
	require.Equal(t, 5, total2, "expected total=5, got %d", total2)
	require.Empty(t, empty, "expected empty slice, got %d events", len(empty))
}

func TestListEvents_ExcludesReversalAndCancelledExpense(t *testing.T) {
	auth, invites, groups, expenses := setupExpenseTest(t)
	ctx := context.Background()
	groupID := registerAndCreateGroup(t, auth, invites, groups)

	kept, _ := expenses.AddExpense(ctx, groupID, "alice", "Kept", "alice", 60.0, map[string]float64{
		"alice": 30.0,
		"bob":   30.0,
	}, nil)
	cancelled, _ := expenses.AddExpense(ctx, groupID, "alice", "Cancelled", "alice", 40.0, map[string]float64{
		"alice": 20.0,
		"bob":   20.0,
	}, nil)
	require.NoError(t, expenses.CancelExpense(ctx, groupID, "alice", cancelled))

	events, total, err := expenses.ListEvents(ctx, groupID, 10, 0, false)
	require.NoError(t, err, "list events: %v", err)
	require.Equal(t, 1, total, "total must count only the events actually returned")
	require.Len(t, events, 1)
	assert.Equal(t, kept, events[0].ID, "expected only the non-cancelled expense to survive")

	for _, e := range events {
		assert.NotEqual(t, domain.EventTypeReversal, e.Type, "reversal event leaked into filtered list")
		assert.NotEqual(t, cancelled, e.ID, "cancelled expense leaked into filtered list")
	}
}

func TestListEvents_IncludeReversedReturnsFullLog(t *testing.T) {
	auth, invites, groups, expenses := setupExpenseTest(t)
	ctx := context.Background()
	groupID := registerAndCreateGroup(t, auth, invites, groups)

	kept, _ := expenses.AddExpense(ctx, groupID, "alice", "Kept", "alice", 60.0, map[string]float64{
		"alice": 30.0,
		"bob":   30.0,
	}, nil)
	cancelled, _ := expenses.AddExpense(ctx, groupID, "alice", "Cancelled", "alice", 40.0, map[string]float64{
		"alice": 20.0,
		"bob":   20.0,
	}, nil)
	require.NoError(t, expenses.CancelExpense(ctx, groupID, "alice", cancelled))

	events, total, err := expenses.ListEvents(ctx, groupID, 10, 0, true)
	require.NoError(t, err, "list events: %v", err)
	require.Equal(t, 3, total, "expected the unfiltered log")
	require.Len(t, events, 3)

	// newest-first: reversal, cancelled expense, kept expense
	assert.Equal(t, domain.EventTypeReversal, events[0].Type)
	assert.Equal(t, cancelled, events[0].ReversedEventID, "reversal should point at the cancelled expense")
	assert.Equal(t, cancelled, events[1].ID)
	assert.Equal(t, kept, events[2].ID)
}

func TestListEvents_ExcludesReversedAcrossPagination(t *testing.T) {
	auth, invites, groups, expenses := setupExpenseTest(t)
	ctx := context.Background()
	groupID := registerAndCreateGroup(t, auth, invites, groups)

	ids := make([]string, 0, 5)
	for i := 0; i < 5; i++ {
		id, err := expenses.AddExpense(ctx, groupID, "alice", "Expense", "alice", 20.0, map[string]float64{
			"alice": 10.0,
			"bob":   10.0,
		}, nil)
		require.NoError(t, err, "add expense %d: %v", i, err)
		ids = append(ids, id)
	}
	// cancel the newest expense: if filtering ran after pagination, page 1 would
	// be emptied by the dropped reversal + cancelled pair rather than backfilled.
	require.NoError(t, expenses.CancelExpense(ctx, groupID, "alice", ids[4]))

	page1, total, err := expenses.ListEvents(ctx, groupID, 2, 0, false)
	require.NoError(t, err, "list page 1: %v", err)
	require.Equal(t, 4, total, "total should exclude the reversal and the expense it cancelled")
	require.Len(t, page1, 2, "page 1 must be a full page of surviving events")
	assert.Equal(t, ids[3], page1[0].ID)
	assert.Equal(t, ids[2], page1[1].ID)

	page2, _, err := expenses.ListEvents(ctx, groupID, 2, 2, false)
	require.NoError(t, err, "list page 2: %v", err)
	require.Len(t, page2, 2)
	assert.Equal(t, ids[1], page2[0].ID)
	assert.Equal(t, ids[0], page2[1].ID)
}

func TestListEvents_FilteringKeepsSettlements(t *testing.T) {
	st := newTestStore(t)
	ctx := context.Background()
	auth := service.NewAuthService(st, "secret", testAuthExpiry, zaptest.NewLogger(t))
	invites := service.NewInviteService(st, zaptest.NewLogger(t))
	groups := service.NewGroupService(st, zaptest.NewLogger(t))
	settlements := service.NewSettlementService(st, zaptest.NewLogger(t))
	expenses := service.NewExpenseService(st, zaptest.NewLogger(t))

	groupID := registerAndCreateGroup(t, auth, invites, groups)

	cancelled, _ := expenses.AddExpense(ctx, groupID, "alice", "Cancelled", "alice", 40.0, map[string]float64{
		"alice": 20.0,
		"bob":   20.0,
	}, nil)
	require.NoError(t, settlements.AddSettlement(ctx, groupID, "bob", "bob", "alice", 50.0))
	require.NoError(t, expenses.CancelExpense(ctx, groupID, "alice", cancelled))

	events, total, err := expenses.ListEvents(ctx, groupID, 10, 0, false)
	require.NoError(t, err, "list events: %v", err)
	require.Equal(t, 1, total, "settlement should survive reversal filtering")
	require.Len(t, events, 1)
	assert.Equal(t, domain.EventTypeSettlement, events[0].Type)
	assert.Equal(t, "bob", events[0].From)
	assert.Equal(t, "alice", events[0].To)
}

func TestAddExpense_WithOriginalExpense(t *testing.T) {
	auth, invites, groups, expenses := setupExpenseTest(t)
	ctx := context.Background()
	groupID := registerAndCreateGroup(t, auth, invites, groups)

	orig := &domain.OriginalExpense{Currency: "GBP", Amount: 45.0}
	eventID, err := expenses.AddExpense(ctx, groupID, "alice", "Dinner", "alice", 50.0, map[string]float64{
		"alice": 25.0,
		"bob":   25.0,
	}, orig)
	require.NoError(t, err)
	require.NotEmpty(t, eventID)

	events, _, err := expenses.ListEvents(ctx, groupID, 10, 0, false)
	require.NoError(t, err)
	require.Len(t, events, 1)
	require.NotNil(t, events[0].OriginalExpense)
	require.Equal(t, "GBP", events[0].OriginalExpense.Currency)
	require.Equal(t, 45.0, events[0].OriginalExpense.Amount)
}
