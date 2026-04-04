// billsplit/internal/domain/balance_test.go
package domain_test

import (
	"math"
	"testing"

	"github.com/fergalhk-lab/apps/billsplit/internal/domain"
	"github.com/stretchr/testify/assert"
)

func TestComputeBalances_SimpleExpense(t *testing.T) {
	g := domain.Group{
		Members: []string{"alice", "bob"},
		Events: []domain.Event{
			{ID: "e1", Type: domain.EventTypeExpense, PaidBy: "alice", Amount: 100,
				Splits: map[string]float64{"alice": 50, "bob": 50}},
		},
	}
	b := domain.ComputeBalances(g)
	assert.Equal(t, 50.0, b["alice"], "alice: want 50, got %v", b["alice"])
	assert.Equal(t, -50.0, b["bob"], "bob: want -50, got %v", b["bob"])
}

func TestComputeBalances_UnequalSplits(t *testing.T) {
	g := domain.Group{
		Members: []string{"alice", "bob", "carol"},
		Events: []domain.Event{
			{ID: "e1", Type: domain.EventTypeExpense, PaidBy: "alice", Amount: 60,
				Splits: map[string]float64{"alice": 10, "bob": 25, "carol": 25}},
		},
	}
	b := domain.ComputeBalances(g)
	assert.Equal(t, 50.0, b["alice"], "alice: want 50, got %v", b["alice"])
	assert.Equal(t, -25.0, b["bob"], "bob: want -25, got %v", b["bob"])
	assert.Equal(t, -25.0, b["carol"], "carol: want -25, got %v", b["carol"])
}

func TestComputeBalances_Settlement(t *testing.T) {
	g := domain.Group{
		Members: []string{"alice", "bob"},
		Events: []domain.Event{
			{ID: "e1", Type: domain.EventTypeExpense, PaidBy: "alice", Amount: 100,
				Splits: map[string]float64{"alice": 50, "bob": 50}},
			{ID: "e2", Type: domain.EventTypeSettlement, From: "bob", To: "alice", Amount: 50},
		},
	}
	b := domain.ComputeBalances(g)
	assert.Equal(t, 0.0, b["alice"], "alice: want 0, got %v", b["alice"])
	assert.Equal(t, 0.0, b["bob"], "bob: want 0, got %v", b["bob"])
}

func TestComputeBalances_Reversal(t *testing.T) {
	g := domain.Group{
		Members: []string{"alice", "bob"},
		Events: []domain.Event{
			{ID: "e1", Type: domain.EventTypeExpense, PaidBy: "alice", Amount: 100,
				Splits: map[string]float64{"alice": 50, "bob": 50}},
			{ID: "e2", Type: domain.EventTypeReversal, ReversedEventID: "e1"},
		},
	}
	b := domain.ComputeBalances(g)
	assert.Equal(t, 0.0, b["alice"], "alice: want 0, got %v", b["alice"])
	assert.Equal(t, 0.0, b["bob"], "bob: want 0, got %v", b["bob"])
}

func TestComputeBalances_EmptyGroup(t *testing.T) {
	g := domain.Group{Members: []string{"alice", "bob"}, Events: nil}
	b := domain.ComputeBalances(g)
	assert.Equal(t, 0.0, b["alice"], "expected zero balances for empty group, got %v", b)
	assert.Equal(t, 0.0, b["bob"], "expected zero balances for empty group, got %v", b)
}

// --- ComputeSettlements tests ---

func TestComputeSettlements_Simple(t *testing.T) {
	// alice is owed 50, bob owes 50
	balances := map[string]float64{"alice": 50, "bob": -50}
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
	// alice is owed 50, bob owes 25, carol owes 25
	balances := map[string]float64{"alice": 50, "bob": -25, "carol": -25}
	settlements := domain.ComputeSettlements(balances)
	if len(settlements) != 2 {
		t.Fatalf("want 2 settlements, got %d: %v", len(settlements), settlements)
	}
	// bob < carol alphabetically, so bob appears first
	if settlements[0].From != "bob" || settlements[0].To != "alice" || settlements[0].Amount != 25 {
		t.Errorf("unexpected settlement[0]: %+v", settlements[0])
	}
	if settlements[1].From != "carol" || settlements[1].To != "alice" || settlements[1].Amount != 25 {
		t.Errorf("unexpected settlement[1]: %+v", settlements[1])
	}
}

func TestComputeSettlements_Balanced(t *testing.T) {
	balances := map[string]float64{"alice": 0, "bob": 0}
	settlements := domain.ComputeSettlements(balances)
	if len(settlements) != 0 {
		t.Fatalf("want 0 settlements, got %d: %v", len(settlements), settlements)
	}
}

func TestComputeSettlements_AllZero(t *testing.T) {
	balances := map[string]float64{"alice": 0, "bob": 0, "carol": 0}
	settlements := domain.ComputeSettlements(balances)
	if len(settlements) != 0 {
		t.Fatalf("want 0 settlements, got %d: %v", len(settlements), settlements)
	}
}

func TestComputeSettlements_SinglePerson(t *testing.T) {
	balances := map[string]float64{"alice": 0}
	settlements := domain.ComputeSettlements(balances)
	if len(settlements) != 0 {
		t.Fatalf("want 0 settlements for single person, got %d", len(settlements))
	}
}

func TestComputeSettlements_Consolidates(t *testing.T) {
	// Net: alice +30, bob +30, carol -30, dave -30
	// Expect 2 settlements with alphabetical debtor matching to alphabetical creditor
	balances := map[string]float64{"alice": 30, "bob": 30, "carol": -30, "dave": -30}
	settlements := domain.ComputeSettlements(balances)
	if len(settlements) != 2 {
		t.Fatalf("want 2 settlements, got %d: %v", len(settlements), settlements)
	}
	// carol < dave (debtors sorted alpha); alice < bob (creditors sorted alpha)
	// carol(-30) → alice(+30): 30
	// dave(-30) → bob(+30): 30
	if settlements[0].From != "carol" || settlements[0].To != "alice" || settlements[0].Amount != 30 {
		t.Errorf("unexpected settlement[0]: %+v", settlements[0])
	}
	if settlements[1].From != "dave" || settlements[1].To != "bob" || settlements[1].Amount != 30 {
		t.Errorf("unexpected settlement[1]: %+v", settlements[1])
	}
}

func TestComputeSettlements_UnevenMatch(t *testing.T) {
	// alice is owed 100, bob owes 60, carol owes 40
	// alphabetical debtors: bob(-60) then carol(-40); creditors: alice(+100)
	// bob(-60) → alice(100): pay 60; alice now +40
	// carol(-40) → alice(40): pay 40
	balances := map[string]float64{"alice": 100, "bob": -60, "carol": -40}
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
	// alice owes bob, bob owes carol — net: alice -10, carol +10, bob 0
	// Should produce a single transfer: alice→carol 10 (bob nets to zero and drops out)
	balances := map[string]float64{"alice": -10, "bob": 0, "carol": 10}
	settlements := domain.ComputeSettlements(balances)
	if len(settlements) != 1 {
		t.Fatalf("want 1 settlement, got %d: %v", len(settlements), settlements)
	}
	if settlements[0].From != "alice" || settlements[0].To != "carol" || settlements[0].Amount != 10 {
		t.Errorf("unexpected settlement: %+v", settlements[0])
	}
}

func TestComputeSettlements_FractionalAmounts(t *testing.T) {
	// Splits that produce fractions — ensure amounts are rounded to 2dp
	// bob and carol each owe alice ~16.67
	balances := map[string]float64{"alice": 33.34, "bob": -16.67, "carol": -16.67}
	settlements := domain.ComputeSettlements(balances)
	if len(settlements) != 2 {
		t.Fatalf("want 2 settlements, got %d: %v", len(settlements), settlements)
	}
	// bob < carol alphabetically
	if settlements[0].From != "bob" || settlements[0].To != "alice" {
		t.Errorf("unexpected settlement[0] direction: %+v", settlements[0])
	}
	if settlements[1].From != "carol" || settlements[1].To != "alice" {
		t.Errorf("unexpected settlement[1] direction: %+v", settlements[1])
	}
	for _, s := range settlements {
		// Amount should be rounded to at most 2 decimal places
		rounded := math.Round(s.Amount*100) / 100
		if s.Amount != rounded {
			t.Errorf("amount %v is not rounded to 2dp", s.Amount)
		}
		if s.Amount <= 0 {
			t.Errorf("amount must be positive, got %v", s.Amount)
		}
	}
}

func TestComputeSettlements_SumsToZeroInvariant(t *testing.T) {
	// Property: sum of all settlement amounts must equal sum of all positive balances
	balances := map[string]float64{
		"alice": 120,
		"bob":   -30,
		"carol": -50,
		"dave":  -40,
	}
	settlements := domain.ComputeSettlements(balances)

	var totalSettled float64
	for _, s := range settlements {
		totalSettled += s.Amount
	}
	var totalOwed float64
	for _, bal := range balances {
		if bal > 0 {
			totalOwed += bal
		}
	}
	// Allow tiny floating-point rounding errors
	diff := math.Abs(totalSettled - totalOwed)
	if diff > 0.01 {
		t.Errorf("settlement total %.2f != total owed %.2f (diff %.4f)", totalSettled, totalOwed, diff)
	}
}

func TestComputeSettlements_NoNegativeAmounts(t *testing.T) {
	balances := map[string]float64{"alice": 75, "bob": -25, "carol": -50}
	settlements := domain.ComputeSettlements(balances)
	for _, s := range settlements {
		if s.Amount <= 0 {
			t.Errorf("settlement has non-positive amount: %+v", s)
		}
	}
}

func TestComputeSettlements_FromAndToAreDifferent(t *testing.T) {
	balances := map[string]float64{"alice": 50, "bob": -50}
	settlements := domain.ComputeSettlements(balances)
	for _, s := range settlements {
		if s.From == s.To {
			t.Errorf("settlement has same From and To: %+v", s)
		}
	}
}

func TestComputeSettlements_AfterFullSettlementEventIsEmpty(t *testing.T) {
	// Simulate a group where a settlement event has already been recorded:
	// alice paid 100 (alice +50, bob -50), then bob settled with alice.
	// ComputeBalances returns all zeros. ComputeSettlements should return empty.
	g := domain.Group{
		Members: []string{"alice", "bob"},
		Events: []domain.Event{
			{ID: "e1", Type: domain.EventTypeExpense, PaidBy: "alice", Amount: 100,
				Splits: map[string]float64{"alice": 50, "bob": 50}},
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
	// Expense then reversal → balances all zero → no settlements needed
	g := domain.Group{
		Members: []string{"alice", "bob"},
		Events: []domain.Event{
			{ID: "e1", Type: domain.EventTypeExpense, PaidBy: "alice", Amount: 100,
				Splits: map[string]float64{"alice": 50, "bob": 50}},
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
	// 5 people: alice paid everything, others owe
	balances := map[string]float64{
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
	// All should pay alice
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
	// $10 split 3 ways: each person owes 3.333..., alice paid so alice is owed 6.666...
	// In practice ComputeBalances would produce: alice +6.666..., bob -3.333..., carol -3.333...
	// We test ComputeSettlements directly with these irrational values.
	third := 10.0 / 3.0
	balances := map[string]float64{"alice": 2 * third, "bob": -third, "carol": -third}
	settlements := domain.ComputeSettlements(balances)
	// Should produce exactly 2 settlements (not 3 due to rounding phantom)
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
		rounded := math.Round(s.Amount*100) / 100
		if s.Amount != rounded {
			t.Errorf("amount %v is not rounded to 2dp", s.Amount)
		}
	}
}
