// billsplit/internal/domain/balance.go
package domain

import (
	"math"
	"sort"
)

// ComputeBalances replays the group event log and returns each member's
// net balance. Positive = owed money; negative = owes money.
func ComputeBalances(g Group) map[string]float64 {
	balances := make(map[string]float64, len(g.Members))
	for _, m := range g.Members {
		balances[m] = 0
	}

	// index expenses for fast reversal lookup
	expenses := make(map[string]Event, len(g.Events))
	for _, e := range g.Events {
		if e.Type == EventTypeExpense {
			expenses[e.ID] = e
		}
	}

	for _, e := range g.Events {
		switch e.Type {
		case EventTypeExpense:
			balances[e.PaidBy] += e.Amount
			for user, split := range e.Splits {
				balances[user] -= split
			}
		case EventTypeReversal:
			if orig, ok := expenses[e.ReversedEventID]; ok {
				balances[orig.PaidBy] -= orig.Amount
				for user, split := range orig.Splits {
					balances[user] += split
				}
			}
		case EventTypeSettlement:
			// from paid to, so from's balance improves and to's decreases
			balances[e.From] += e.Amount
			balances[e.To] -= e.Amount
		}
	}

	return balances
}

// ComputeSettlements returns the minimal list of payments that would settle
// all balances. Debtors are matched greedily against creditors in alphabetical
// order, which ensures deterministic output.
func ComputeSettlements(balances map[string]float64) []Settlement {
	type entry struct {
		user   string
		amount float64 // absolute value
	}

	var debtors, creditors []entry
	for user, bal := range balances {
		if bal < -1e-9 {
			debtors = append(debtors, entry{user, -bal})
		} else if bal > 1e-9 {
			creditors = append(creditors, entry{user, bal})
		}
	}

	sort.Slice(debtors, func(i, j int) bool { return debtors[i].user < debtors[j].user })
	sort.Slice(creditors, func(i, j int) bool { return creditors[i].user < creditors[j].user })

	var result []Settlement
	i, j := 0, 0
	for i < len(debtors) && j < len(creditors) {
		amount := math.Min(debtors[i].amount, creditors[j].amount)
		debtors[i].amount -= amount
		creditors[j].amount -= amount
		rounded := math.Round(amount*100) / 100
		if rounded > 0 {
			result = append(result, Settlement{
				From:   debtors[i].user,
				To:     creditors[j].user,
				Amount: rounded,
			})
		}
		if debtors[i].amount < 1e-9 {
			i++
		}
		if creditors[j].amount < 1e-9 {
			j++
		}
	}

	return result
}
