// billsplit/internal/domain/balance_test.go
package domain_test

import (
	"testing"

	"github.com/fergalhk-lab/apps/billsplit/internal/domain"
	"github.com/stretchr/testify/assert"
)

func TestComputeBalances_SimpleExpense(t *testing.T) {
	g := domain.Group{
		Members: []string{"alice", "bob"},
		Events: []domain.Event{
			{ID: "e1", Type: domain.EventTypeExpense, PaidBy: "alice", Amount: 100,
				Splits: map[string]int64{"alice": 50, "bob": 50}},
		},
	}
	b := domain.ComputeBalances(g)
	assert.Equal(t, int64(50), b["alice"], "alice: want 50, got %v", b["alice"])
	assert.Equal(t, int64(-50), b["bob"], "bob: want -50, got %v", b["bob"])
}

func TestComputeBalances_UnequalSplits(t *testing.T) {
	g := domain.Group{
		Members: []string{"alice", "bob", "carol"},
		Events: []domain.Event{
			{ID: "e1", Type: domain.EventTypeExpense, PaidBy: "alice", Amount: 60,
				Splits: map[string]int64{"alice": 10, "bob": 25, "carol": 25}},
		},
	}
	b := domain.ComputeBalances(g)
	assert.Equal(t, int64(50), b["alice"], "alice: want 50, got %v", b["alice"])
	assert.Equal(t, int64(-25), b["bob"], "bob: want -25, got %v", b["bob"])
	assert.Equal(t, int64(-25), b["carol"], "carol: want -25, got %v", b["carol"])
}

func TestComputeBalances_Settlement(t *testing.T) {
	g := domain.Group{
		Members: []string{"alice", "bob"},
		Events: []domain.Event{
			{ID: "e1", Type: domain.EventTypeExpense, PaidBy: "alice", Amount: 100,
				Splits: map[string]int64{"alice": 50, "bob": 50}},
			{ID: "e2", Type: domain.EventTypeSettlement, From: "bob", To: "alice", Amount: 50},
		},
	}
	b := domain.ComputeBalances(g)
	assert.Equal(t, int64(0), b["alice"], "alice: want 0, got %v", b["alice"])
	assert.Equal(t, int64(0), b["bob"], "bob: want 0, got %v", b["bob"])
}

func TestComputeBalances_Reversal(t *testing.T) {
	g := domain.Group{
		Members: []string{"alice", "bob"},
		Events: []domain.Event{
			{ID: "e1", Type: domain.EventTypeExpense, PaidBy: "alice", Amount: 100,
				Splits: map[string]int64{"alice": 50, "bob": 50}},
			{ID: "e2", Type: domain.EventTypeReversal, ReversedEventID: "e1"},
		},
	}
	b := domain.ComputeBalances(g)
	assert.Equal(t, int64(0), b["alice"], "alice: want 0, got %v", b["alice"])
	assert.Equal(t, int64(0), b["bob"], "bob: want 0, got %v", b["bob"])
}

func TestComputeBalances_EmptyGroup(t *testing.T) {
	g := domain.Group{Members: []string{"alice", "bob"}, Events: nil}
	b := domain.ComputeBalances(g)
	assert.Equal(t, int64(0), b["alice"], "expected zero balances for empty group, got %v", b)
	assert.Equal(t, int64(0), b["bob"], "expected zero balances for empty group, got %v", b)
}

func TestComputeSettlements_Simple(t *testing.T) {
	balances := map[string]int64{"alice": 50, "bob": -50}
	settlements := domain.ComputeSettlements(balances)
	if len(settlements) != 1 {
		t.Fatalf("want 1 settlement, got %d: %v", len(settlements), settlements)
	}
	s := settlements[0]
	if s.From != "bob" || s.To != "alice" || s.Amount != 50 {
		t.Errorf("unexpected settlement: %+v", s)
	}
}

func TestComputeSettlements_ThreePeopleOneCreditor(t *testing.T) {
	balances := map[string]int64{"alice": 50, "bob": -25, "carol": -25}
	settlements := domain.ComputeSettlements(balances)
	if len(settlements) != 2 {
		t.Fatalf("want 2 settlements, got %d: %v", len(settlements), settlements)
	}
	if settlements[0].From != "bob" || settlements[0].To != "alice" || settlements[0].Amount != 25 {
		t.Errorf("unexpected settlement[0]: %+v", settlements[0])
	}
	if settlements[1].From != "carol" || settlements[1].To != "alice" || settlements[1].Amount != 25 {
		t.Errorf("unexpected settlement[1]: %+v", settlements[1])
	}
}

func TestComputeSettlements_Balanced(t *testing.T) {
	balances := map[string]int64{"alice": 0, "bob": 0}
	settlements := domain.ComputeSettlements(balances)
	if len(settlements) != 0 {
		t.Fatalf("want 0 settlements, got %d: %v", len(settlements), settlements)
	}
}

func TestComputeSettlements_AllZero(t *testing.T) {
	balances := map[string]int64{"alice": 0, "bob": 0, "carol": 0}
	settlements := domain.ComputeSettlements(balances)
	if len(settlements) != 0 {
		t.Fatalf("want 0 settlements, got %d: %v", len(settlements), settlements)
	}
}

func TestComputeSettlements_SinglePerson(t *testing.T) {
	balances := map[string]int64{"alice": 0}
	settlements := domain.ComputeSettlements(balances)
	if len(settlements) != 0 {
		t.Fatalf("want 0 settlements for single person, got %d", len(settlements))
	}
}

func TestComputeSettlements_Consolidates(t *testing.T) {
	balances := map[string]int64{"alice": 30, "bob": 30, "carol": -30, "dave": -30}
	settlements := domain.ComputeSettlements(balances)
	if len(settlements) != 2 {
		t.Fatalf("want 2 settlements, got %d: %v", len(settlements), settlements)
	}
	if settlements[0].From != "carol" || settlements[0].To != "alice" || settlements[0].Amount != 30 {
		t.Errorf("unexpected settlement[0]: %+v", settlements[0])
	}
	if settlements[1].From != "dave" || settlements[1].To != "bob" || settlements[1].Amount != 30 {
		t.Errorf("unexpected settlement[1]: %+v", settlements[1])
	}
}

func TestComputeSettlements_UnevenMatch(t *testing.T) {
	balances := map[string]int64{"alice": 100, "bob": -60, "carol": -40}
	settlements := domain.ComputeSettlements(balances)
	if len(settlements) != 2 {
		t.Fatalf("want 2 settlements, got %d: %v", len(settlements), settlements)
	}
	if settlements[0].From != "bob" || settlements[0].To != "alice" || settlements[0].Amount != 60 {
		t.Errorf("unexpected settlement[0]: %+v", settlements[0])
	}
	if settlements[1].From != "carol" || settlements[1].To != "alice" || settlements[1].Amount != 40 {
		t.Errorf("unexpected settlement[1]: %+v", settlements[1])
	}
}

func TestComputeSettlements_ChainedDebts(t *testing.T) {
	balances := map[string]int64{"alice": -10, "bob": 0, "carol": 10}
	settlements := domain.ComputeSettlements(balances)
	if len(settlements) != 1 {
		t.Fatalf("want 1 settlement, got %d: %v", len(settlements), settlements)
	}
	if settlements[0].From != "alice" || settlements[0].To != "carol" || settlements[0].Amount != 10 {
		t.Errorf("unexpected settlement: %+v", settlements[0])
	}
}

func TestComputeSettlements_FractionalAmounts(t *testing.T) {
	// 3334 cents = $33.34, -1667 = -$16.67 each — integer cents, no rounding needed
	balances := map[string]int64{"alice": 3334, "bob": -1667, "carol": -1667}
	settlements := domain.ComputeSettlements(balances)
	if len(settlements) != 2 {
		t.Fatalf("want 2 settlements, got %d: %v", len(settlements), settlements)
	}
	if settlements[0].From != "bob" || settlements[0].To != "alice" {
		t.Errorf("unexpected settlement[0] direction: %+v", settlements[0])
	}
	if settlements[1].From != "carol" || settlements[1].To != "alice" {
		t.Errorf("unexpected settlement[1] direction: %+v", settlements[1])
	}
	for _, s := range settlements {
		if s.Amount <= 0 {
			t.Errorf("amount must be positive, got %v", s.Amount)
		}
	}
}

func TestComputeSettlements_SumsToZeroInvariant(t *testing.T) {
	balances := map[string]int64{
		"alice": 120,
		"bob":   -30,
		"carol": -50,
		"dave":  -40,
	}
	settlements := domain.ComputeSettlements(balances)

	var totalSettled int64
	for _, s := range settlements {
		totalSettled += s.Amount
	}
	var totalOwed int64
	for _, bal := range balances {
		if bal > 0 {
			totalOwed += bal
		}
	}
	if totalSettled != totalOwed {
		t.Errorf("settlement total %d != total owed %d", totalSettled, totalOwed)
	}
}

func TestComputeSettlements_NoNegativeAmounts(t *testing.T) {
	balances := map[string]int64{"alice": 75, "bob": -25, "carol": -50}
	settlements := domain.ComputeSettlements(balances)
	for _, s := range settlements {
		if s.Amount <= 0 {
			t.Errorf("settlement has non-positive amount: %+v", s)
		}
	}
}

func TestComputeSettlements_FromAndToAreDifferent(t *testing.T) {
	balances := map[string]int64{"alice": 50, "bob": -50}
	settlements := domain.ComputeSettlements(balances)
	for _, s := range settlements {
		if s.From == s.To {
			t.Errorf("settlement has same From and To: %+v", s)
		}
	}
}

func TestComputeSettlements_AfterFullSettlementEventIsEmpty(t *testing.T) {
	g := domain.Group{
		Members: []string{"alice", "bob"},
		Events: []domain.Event{
			{ID: "e1", Type: domain.EventTypeExpense, PaidBy: "alice", Amount: 100,
				Splits: map[string]int64{"alice": 50, "bob": 50}},
			{ID: "e2", Type: domain.EventTypeSettlement, From: "bob", To: "alice", Amount: 50},
		},
	}
	balances := domain.ComputeBalances(g)
	settlements := domain.ComputeSettlements(balances)
	if len(settlements) != 0 {
		t.Fatalf("want 0 settlements after full settlement event, got %d: %v", len(settlements), settlements)
	}
}

func TestComputeSettlements_AfterReversalEventIsEmpty(t *testing.T) {
	g := domain.Group{
		Members: []string{"alice", "bob"},
		Events: []domain.Event{
			{ID: "e1", Type: domain.EventTypeExpense, PaidBy: "alice", Amount: 100,
				Splits: map[string]int64{"alice": 50, "bob": 50}},
			{ID: "e2", Type: domain.EventTypeReversal, ReversedEventID: "e1"},
		},
	}
	balances := domain.ComputeBalances(g)
	settlements := domain.ComputeSettlements(balances)
	if len(settlements) != 0 {
		t.Fatalf("want 0 settlements after reversal, got %d: %v", len(settlements), settlements)
	}
}

func TestComputeSettlements_LargeGroup(t *testing.T) {
	balances := map[string]int64{
		"alice": 80,
		"bob":   -20,
		"carol": -20,
		"dave":  -20,
		"eve":   -20,
	}
	settlements := domain.ComputeSettlements(balances)
	if len(settlements) != 4 {
		t.Fatalf("want 4 settlements, got %d: %v", len(settlements), settlements)
	}
	for _, s := range settlements {
		if s.To != "alice" {
			t.Errorf("expected all settlements to go to alice, got: %+v", s)
		}
		if s.Amount != 20 {
			t.Errorf("expected amount 20, got: %+v", s)
		}
	}
}

func TestComputeSettlements_IrrationalSplit(t *testing.T) {
	// 666 cents owed to alice; bob owes 333, carol owes 333 — integer, no phantom rounding
	balances := map[string]int64{"alice": 666, "bob": -333, "carol": -333}
	settlements := domain.ComputeSettlements(balances)
	if len(settlements) != 2 {
		t.Fatalf("want 2 settlements, got %d: %v", len(settlements), settlements)
	}
	if settlements[0].From != "bob" || settlements[0].To != "alice" {
		t.Errorf("unexpected settlement[0]: %+v", settlements[0])
	}
	if settlements[1].From != "carol" || settlements[1].To != "alice" {
		t.Errorf("unexpected settlement[1]: %+v", settlements[1])
	}
	for _, s := range settlements {
		if s.Amount <= 0 {
			t.Errorf("amount must be positive, got %v", s.Amount)
		}
	}
}
