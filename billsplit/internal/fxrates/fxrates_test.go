package fxrates_test

import (
	"math"
	"testing"

	"github.com/fergalhk-lab/apps/billsplit/internal/fxrates"
	"github.com/stretchr/testify/require"
)

func TestConvert_SameCurrency(t *testing.T) {
	r := fxrates.Rates{
		Base:  "USD",
		Rates: map[string]float64{"USD": 1.0, "EUR": 0.867387},
	}
	got, err := r.Convert(100.0, "EUR", "EUR")
	require.NoError(t, err)
	require.Equal(t, 100.0, got)
}

func TestConvert_USDToEUR(t *testing.T) {
	r := fxrates.Rates{
		Base:  "USD",
		Rates: map[string]float64{"USD": 1.0, "EUR": 0.867387},
	}
	got, err := r.Convert(100.0, "USD", "EUR")
	require.NoError(t, err)
	// 100 USD at 0.867387 EUR/USD = 86.7387 EUR; allow for float64 rounding.
	require.True(t, math.Abs(got-86.7387) <= 0.0001, "got %v, want ~86.7387", got)
}

func TestConvert_CrossRate(t *testing.T) {
	r := fxrates.Rates{
		Base:  "USD",
		Rates: map[string]float64{"USD": 1.0, "EUR": 0.867387, "GBP": 0.757571},
	}
	got, err := r.Convert(100.0, "EUR", "GBP")
	require.NoError(t, err)
	want := 100.0 / 0.867387 * 0.757571
	require.Equal(t, want, got)
}

func TestConvert_UnknownFromCurrency(t *testing.T) {
	r := fxrates.Rates{
		Base:  "USD",
		Rates: map[string]float64{"USD": 1.0, "EUR": 0.867387},
	}
	_, err := r.Convert(100.0, "XYZ", "EUR")
	require.Error(t, err, "expected error for unknown currency XYZ")
	require.ErrorContains(t, err, "XYZ")
}

func TestConvert_UnknownToCurrency(t *testing.T) {
	r := fxrates.Rates{
		Base:  "USD",
		Rates: map[string]float64{"USD": 1.0, "EUR": 0.867387},
	}
	_, err := r.Convert(100.0, "EUR", "XYZ")
	require.Error(t, err, "expected error for unknown currency XYZ")
	require.ErrorContains(t, err, "XYZ")
}
