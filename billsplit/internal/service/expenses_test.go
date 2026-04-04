// billsplit/internal/service/expenses_test.go
package service_test

import (
	"context"
	"errors"
	"testing"

	"github.com/fergalhk-lab/apps/billsplit/internal/service"
)

func setupExpenseTest(t *testing.T) (*service.AuthService, *service.InviteService, *service.GroupService, *service.ExpenseService) {
	t.Helper()
	st := newTestStore(t)
	auth := service.NewAuthService(st, "secret")
	invites := service.NewInviteService(st)
	groups := service.NewGroupService(st)
	expenses := service.NewExpenseService(st)
	return auth, invites, groups, expenses
}

// registerAndCreateGroup is a test helper that registers two users (alice and
// bob) and creates a group containing both.
func registerAndCreateGroup(t *testing.T, auth *service.AuthService, invites *service.InviteService, groups *service.GroupService) (groupID string) {
	t.Helper()
	ctx := context.Background()

	codeA, err := invites.GenerateInvite(ctx, false)
	if err != nil {
		t.Fatalf("generate invite alice: %v", err)
	}
	if err := auth.Register(ctx, "alice", "pw", codeA); err != nil {
		t.Fatalf("register alice: %v", err)
	}

	codeB, err := invites.GenerateInvite(ctx, false)
	if err != nil {
		t.Fatalf("generate invite bob: %v", err)
	}
	if err := auth.Register(ctx, "bob", "pw", codeB); err != nil {
		t.Fatalf("register bob: %v", err)
	}

	groupID, err = groups.CreateGroup(ctx, "alice", "Trip", "EUR", []string{"bob"})
	if err != nil {
		t.Fatalf("create group: %v", err)
	}
	return groupID
}

func TestAddExpense(t *testing.T) {
	auth, invites, groups, expenses := setupExpenseTest(t)
	ctx := context.Background()
	groupID := registerAndCreateGroup(t, auth, invites, groups)

	eventID, err := expenses.AddExpense(ctx, groupID, "alice", "Dinner", "alice", 100.0, map[string]float64{
		"alice": 50.0,
		"bob":   50.0,
	})
	if err != nil {
		t.Fatalf("add expense: %v", err)
	}
	if eventID == "" {
		t.Fatal("expected non-empty event ID")
	}
}

func TestAddExpense_InvalidSplits(t *testing.T) {
	auth, invites, groups, expenses := setupExpenseTest(t)
	ctx := context.Background()
	groupID := registerAndCreateGroup(t, auth, invites, groups)

	// splits don't sum to total
	_, err := expenses.AddExpense(ctx, groupID, "alice", "Dinner", "alice", 100.0, map[string]float64{
		"alice": 40.0,
		"bob":   40.0,
	})
	if err == nil {
		t.Fatal("expected error for invalid splits")
	}
}

func TestAddExpense_UnknownMemberInSplits(t *testing.T) {
	auth, invites, groups, expenses := setupExpenseTest(t)
	ctx := context.Background()
	groupID := registerAndCreateGroup(t, auth, invites, groups)

	_, err := expenses.AddExpense(ctx, groupID, "alice", "Dinner", "alice", 100.0, map[string]float64{
		"alice":  50.0,
		"nobody": 50.0,
	})
	if err == nil {
		t.Fatal("expected error for unknown split member")
	}
}

func TestCancelExpense(t *testing.T) {
	auth, invites, groups, expenses := setupExpenseTest(t)
	ctx := context.Background()
	groupID := registerAndCreateGroup(t, auth, invites, groups)

	eventID, err := expenses.AddExpense(ctx, groupID, "alice", "Dinner", "alice", 100.0, map[string]float64{
		"alice": 50.0,
		"bob":   50.0,
	})
	if err != nil {
		t.Fatalf("add expense: %v", err)
	}

	if err := expenses.CancelExpense(ctx, groupID, "alice", eventID); err != nil {
		t.Fatalf("cancel expense: %v", err)
	}
}

func TestCancelExpense_AlreadyCancelled(t *testing.T) {
	auth, invites, groups, expenses := setupExpenseTest(t)
	ctx := context.Background()
	groupID := registerAndCreateGroup(t, auth, invites, groups)

	eventID, _ := expenses.AddExpense(ctx, groupID, "alice", "Dinner", "alice", 100.0, map[string]float64{
		"alice": 50.0,
		"bob":   50.0,
	})
	_ = expenses.CancelExpense(ctx, groupID, "alice", eventID)

	err := expenses.CancelExpense(ctx, groupID, "alice", eventID)
	if err == nil {
		t.Fatal("expected error cancelling already-cancelled expense")
	}
}

func TestCancelExpense_NotFound(t *testing.T) {
	auth, invites, groups, expenses := setupExpenseTest(t)
	ctx := context.Background()
	groupID := registerAndCreateGroup(t, auth, invites, groups)

	err := expenses.CancelExpense(ctx, groupID, "alice", "nonexistent-id")
	if !errors.Is(err, service.ErrEventNotFound) {
		t.Fatalf("expected ErrEventNotFound, got %v", err)
	}
}

func TestListEvents_NewestFirst(t *testing.T) {
	auth, invites, groups, expenses := setupExpenseTest(t)
	ctx := context.Background()
	groupID := registerAndCreateGroup(t, auth, invites, groups)

	id1, _ := expenses.AddExpense(ctx, groupID, "alice", "First", "alice", 60.0, map[string]float64{
		"alice": 30.0,
		"bob":   30.0,
	})
	id2, _ := expenses.AddExpense(ctx, groupID, "alice", "Second", "alice", 40.0, map[string]float64{
		"alice": 20.0,
		"bob":   20.0,
	})

	events, total, err := expenses.ListEvents(ctx, groupID, 10, 0)
	if err != nil {
		t.Fatalf("list events: %v", err)
	}
	if total != 2 {
		t.Fatalf("expected total=2, got %d", total)
	}
	// newest-first: second expense should come before first
	if events[0].ID != id2 {
		t.Errorf("expected first result to be %s (second expense), got %s", id2, events[0].ID)
	}
	if events[1].ID != id1 {
		t.Errorf("expected second result to be %s (first expense), got %s", id1, events[1].ID)
	}
}

func TestListEvents_Pagination(t *testing.T) {
	auth, invites, groups, expenses := setupExpenseTest(t)
	ctx := context.Background()
	groupID := registerAndCreateGroup(t, auth, invites, groups)

	for i := 0; i < 5; i++ {
		_, err := expenses.AddExpense(ctx, groupID, "alice", "Expense", "alice", 20.0, map[string]float64{
			"alice": 10.0,
			"bob":   10.0,
		})
		if err != nil {
			t.Fatalf("add expense %d: %v", i, err)
		}
	}

	page1, total, err := expenses.ListEvents(ctx, groupID, 2, 0)
	if err != nil {
		t.Fatalf("list page 1: %v", err)
	}
	if total != 5 {
		t.Fatalf("expected total=5, got %d", total)
	}
	if len(page1) != 2 {
		t.Fatalf("expected 2 events on page 1, got %d", len(page1))
	}

	page3, _, err := expenses.ListEvents(ctx, groupID, 2, 4)
	if err != nil {
		t.Fatalf("list page 3: %v", err)
	}
	if len(page3) != 1 {
		t.Fatalf("expected 1 event on last page, got %d", len(page3))
	}

	// offset beyond end returns empty slice
	empty, total2, err := expenses.ListEvents(ctx, groupID, 2, 10)
	if err != nil {
		t.Fatalf("list beyond end: %v", err)
	}
	if total2 != 5 {
		t.Fatalf("expected total=5, got %d", total2)
	}
	if len(empty) != 0 {
		t.Fatalf("expected empty slice, got %d events", len(empty))
	}
}
