// billsplit/internal/domain/balance.go
package domain

import "sort"

// ComputeBalances replays the group event log and returns each member's
// net balance in cents. Positive = owed money; negative = owes money.
func ComputeBalances(g Group) map[string]int64 {
	balances := make(map[string]int64, len(g.Members))
	for _, m := range g.Members {
		balances[m] = 0
	}

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
			balances[e.From] += e.Amount
			balances[e.To] -= e.Amount
		}
	}

	return balances
}

// ComputeSettlements returns the minimal list of payments that would settle
// all balances. Debtors are matched greedily against creditors in alphabetical
// order, which ensures deterministic output.
func ComputeSettlements(balances map[string]int64) []Settlement {
	type entry struct {
		user   string
		amount int64
	}

	var debtors, creditors []entry
	for user, bal := range balances {
		if bal < 0 {
			debtors = append(debtors, entry{user, -bal})
		} else if bal > 0 {
			creditors = append(creditors, entry{user, bal})
		}
	}

	sort.Slice(debtors, func(i, j int) bool { return debtors[i].user < debtors[j].user })
	sort.Slice(creditors, func(i, j int) bool { return creditors[i].user < creditors[j].user })

	result := make([]Settlement, 0)
	i, j := 0, 0
	for i < len(debtors) && j < len(creditors) {
		amount := min(debtors[i].amount, creditors[j].amount)
		debtors[i].amount -= amount
		creditors[j].amount -= amount
		if amount > 0 {
			result = append(result, Settlement{
				From:   debtors[i].user,
				To:     creditors[j].user,
				Amount: amount,
			})
		}
		if debtors[i].amount == 0 {
			i++
		}
		if creditors[j].amount == 0 {
			j++
		}
	}

	return result
}
