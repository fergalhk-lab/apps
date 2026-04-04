// billsplit/internal/domain/balance_test.go
package domain_test

import (
	"testing"

	"github.com/fergalhk-lab/apps/billsplit/internal/domain"
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
	if b["alice"] != 50 {
		t.Errorf("alice: want 50, got %v", b["alice"])
	}
	if b["bob"] != -50 {
		t.Errorf("bob: want -50, got %v", b["bob"])
	}
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
	if b["alice"] != 50 {
		t.Errorf("alice: want 50, got %v", b["alice"])
	}
	if b["bob"] != -25 {
		t.Errorf("bob: want -25, got %v", b["bob"])
	}
	if b["carol"] != -25 {
		t.Errorf("carol: want -25, got %v", b["carol"])
	}
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
	if b["alice"] != 0 {
		t.Errorf("alice: want 0, got %v", b["alice"])
	}
	if b["bob"] != 0 {
		t.Errorf("bob: want 0, got %v", b["bob"])
	}
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
	if b["alice"] != 0 {
		t.Errorf("alice: want 0, got %v", b["alice"])
	}
	if b["bob"] != 0 {
		t.Errorf("bob: want 0, got %v", b["bob"])
	}
}

func TestComputeBalances_EmptyGroup(t *testing.T) {
	g := domain.Group{Members: []string{"alice", "bob"}, Events: nil}
	b := domain.ComputeBalances(g)
	if b["alice"] != 0 || b["bob"] != 0 {
		t.Errorf("expected zero balances for empty group, got %v", b)
	}
}
