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
