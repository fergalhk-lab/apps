// billsplit/internal/domain/balance.go
package domain

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
