package fxrates_test

import (
	"math"
	"strings"
	"testing"

	"github.com/fergalhk-lab/apps/billsplit/internal/fxrates"
)

func TestConvert_SameCurrency(t *testing.T) {
	r := fxrates.Rates{
		Base:  "USD",
		Rates: map[string]float64{"USD": 1.0, "EUR": 0.867387},
	}
	got, err := r.Convert(100.0, "EUR", "EUR")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got != 100.0 {
		t.Fatalf("got %v, want 100.0", got)
	}
}

func TestConvert_USDToEUR(t *testing.T) {
	r := fxrates.Rates{
		Base:  "USD",
		Rates: map[string]float64{"USD": 1.0, "EUR": 0.867387},
	}
	got, err := r.Convert(100.0, "USD", "EUR")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	// 100 USD at 0.867387 EUR/USD = 86.7387 EUR; allow for float64 rounding.
	if math.Abs(got-86.7387) > 0.0001 {
		t.Fatalf("got %v, want ~86.7387", got)
	}
}

func TestConvert_CrossRate(t *testing.T) {
	r := fxrates.Rates{
		Base:  "USD",
		Rates: map[string]float64{"USD": 1.0, "EUR": 0.867387, "GBP": 0.757571},
	}
	got, err := r.Convert(100.0, "EUR", "GBP")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	want := 100.0 / 0.867387 * 0.757571
	if got != want {
		t.Fatalf("got %v, want %v", got, want)
	}
}

func TestConvert_UnknownFromCurrency(t *testing.T) {
	r := fxrates.Rates{
		Base:  "USD",
		Rates: map[string]float64{"USD": 1.0, "EUR": 0.867387},
	}
	_, err := r.Convert(100.0, "XYZ", "EUR")
	if err == nil {
		t.Fatal("expected error for unknown currency XYZ")
	}
	if !strings.Contains(err.Error(), "XYZ") {
		t.Fatalf("error should mention the unknown currency, got: %v", err)
	}
}

func TestConvert_UnknownToCurrency(t *testing.T) {
	r := fxrates.Rates{
		Base:  "USD",
		Rates: map[string]float64{"USD": 1.0, "EUR": 0.867387},
	}
	_, err := r.Convert(100.0, "EUR", "XYZ")
	if err == nil {
		t.Fatal("expected error for unknown currency XYZ")
	}
	if !strings.Contains(err.Error(), "XYZ") {
		t.Fatalf("error should mention the unknown currency, got: %v", err)
	}
}
