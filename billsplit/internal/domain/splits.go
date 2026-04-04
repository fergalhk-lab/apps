// billsplit/internal/domain/splits.go
package domain

import (
	"fmt"
	"math"
)

const splitEpsilon = 0.01

// ValidateSplits checks that splits are valid: all keys are group members,
// no values are negative, and the sum equals total within splitEpsilon.
func ValidateSplits(total float64, splits map[string]float64, members []string) error {
	if len(splits) == 0 {
		return fmt.Errorf("splits must not be empty")
	}

	memberSet := make(map[string]struct{}, len(members))
	for _, m := range members {
		memberSet[m] = struct{}{}
	}

	var sum float64
	for user, amount := range splits {
		if _, ok := memberSet[user]; !ok {
			return fmt.Errorf("unknown member %q in splits", user)
		}
		if amount < 0 {
			return fmt.Errorf("split for %q must not be negative", user)
		}
		sum += amount
	}

	if math.Abs(sum-total) > splitEpsilon {
		return fmt.Errorf("splits sum %.2f does not equal total %.2f", sum, total)
	}
	return nil
}
