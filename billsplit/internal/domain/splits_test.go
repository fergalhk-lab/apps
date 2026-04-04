// billsplit/internal/domain/splits_test.go
package domain_test

import (
	"testing"

	"github.com/fergalhk-lab/apps/billsplit/internal/domain"
	"github.com/stretchr/testify/require"
)

func TestValidateSplits_Valid(t *testing.T) {
	err := domain.ValidateSplits(100, map[string]float64{"alice": 30, "bob": 70}, []string{"alice", "bob"})
	require.NoError(t, err)
}

func TestValidateSplits_DoesNotSum(t *testing.T) {
	err := domain.ValidateSplits(100, map[string]float64{"alice": 30, "bob": 60}, []string{"alice", "bob"})
	require.Error(t, err, "expected error: splits don't sum to total")
}

func TestValidateSplits_UnknownMember(t *testing.T) {
	err := domain.ValidateSplits(100, map[string]float64{"alice": 50, "carol": 50}, []string{"alice", "bob"})
	require.Error(t, err, "expected error: unknown member in splits")
}

func TestValidateSplits_NegativeAmount(t *testing.T) {
	err := domain.ValidateSplits(100, map[string]float64{"alice": 110, "bob": -10}, []string{"alice", "bob"})
	require.Error(t, err, "expected error: negative split")
}

func TestValidateSplits_Empty(t *testing.T) {
	err := domain.ValidateSplits(100, map[string]float64{}, []string{"alice", "bob"})
	require.Error(t, err, "expected error: empty splits")
}
