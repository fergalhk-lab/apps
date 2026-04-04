package fxrates

import (
	"fmt"
	"time"
)

// S3Key is the object key used to store and retrieve exchange rates in S3.
const S3Key = "fxrates/latest.json"

// Rates holds USD-based exchange rates fetched from the provider.
// All values in Rates are relative to USD (USD = 1.0).
type Rates struct {
	Base              string             `json:"base"`
	Rates             map[string]float64 `json:"rates"`
	ProviderUpdatedAt time.Time          `json:"providerUpdatedAt"`
}

// Convert converts amount from one currency to another using USD as the pivot.
// Returns an error if either currency code is absent from the rates map.
func (r *Rates) Convert(amount float64, from, to string) (float64, error) {
	if from == to {
		return amount, nil
	}
	fromRate, ok := r.Rates[from]
	if !ok {
		return 0, fmt.Errorf("unknown currency: %s", from)
	}
	if fromRate == 0 {
		return 0, fmt.Errorf("invalid rate for currency %s: zero value", from)
	}
	toRate, ok := r.Rates[to]
	if !ok {
		return 0, fmt.Errorf("unknown currency: %s", to)
	}
	return amount / fromRate * toRate, nil
}
