package provider_test

import (
	"testing"
	"time"

	"github.com/fergalhk-lab/apps/billsplit/internal/fxrates/provider"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const sampleResponse = `{
	"result": "success",
	"time_last_update_unix": 1775198851,
	"base_code": "USD",
	"rates": {
		"USD": 1,
		"EUR": 0.867387,
		"GBP": 0.757571
	}
}`

func TestParse_Success(t *testing.T) {
	rates, err := provider.Parse([]byte(sampleResponse))
	require.NoError(t, err)
	assert.Equal(t, "USD", rates.Base)
	assert.Equal(t, 0.867387, rates.Rates["EUR"])
	assert.Equal(t, 0.757571, rates.Rates["GBP"])
	want := time.Unix(1775198851, 0).UTC()
	assert.True(t, rates.ProviderUpdatedAt.Equal(want), "got ProviderUpdatedAt=%v, want %v", rates.ProviderUpdatedAt, want)
}

func TestParse_NonSuccessResult(t *testing.T) {
	data := `{"result":"error","time_last_update_unix":0,"base_code":"USD","rates":{}}`
	_, err := provider.Parse([]byte(data))
	require.Error(t, err, "expected error for result=error")
	require.ErrorContains(t, err, "result=")
}

func TestParse_EmptyRates(t *testing.T) {
	data := `{"result":"success","time_last_update_unix":0,"base_code":"USD"}`
	_, err := provider.Parse([]byte(data))
	require.Error(t, err, "expected error for absent rates map")
}

func TestParse_InvalidJSON(t *testing.T) {
	_, err := provider.Parse([]byte(`not json`))
	require.Error(t, err, "expected error for invalid JSON")
}
