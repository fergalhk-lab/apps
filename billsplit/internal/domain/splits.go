// billsplit/internal/domain/splits.go
package domain

import "fmt"

// ValidateSplits checks that splits are valid: all keys are group members,
// no values are negative, and the sum equals total exactly.
func ValidateSplits(total int64, splits map[string]int64, members []string) error {
	if len(splits) == 0 {
		return fmt.Errorf("splits must not be empty")
	}

	memberSet := make(map[string]struct{}, len(members))
	for _, m := range members {
		memberSet[m] = struct{}{}
	}

	var sum int64
	for user, amount := range splits {
		if _, ok := memberSet[user]; !ok {
			return fmt.Errorf("unknown member %q in splits", user)
		}
		if amount < 0 {
			return fmt.Errorf("split for %q must not be negative", user)
		}
		sum += amount
	}

	if sum != total {
		return fmt.Errorf("splits sum %d does not equal total %d", sum, total)
	}
	return nil
}
