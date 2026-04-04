package provider_test

import (
	"strings"
	"testing"
	"time"

	"github.com/fergalhk-lab/apps/billsplit/internal/fxrates/provider"
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
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if rates.Base != "USD" {
		t.Errorf("got Base=%q, want %q", rates.Base, "USD")
	}
	if rates.Rates["EUR"] != 0.867387 {
		t.Errorf("got EUR=%v, want 0.867387", rates.Rates["EUR"])
	}
	if rates.Rates["GBP"] != 0.757571 {
		t.Errorf("got GBP=%v, want 0.757571", rates.Rates["GBP"])
	}
	want := time.Unix(1775198851, 0).UTC()
	if !rates.ProviderUpdatedAt.Equal(want) {
		t.Errorf("got ProviderUpdatedAt=%v, want %v", rates.ProviderUpdatedAt, want)
	}
}

func TestParse_NonSuccessResult(t *testing.T) {
	data := `{"result":"error","time_last_update_unix":0,"base_code":"USD","rates":{}}`
	_, err := provider.Parse([]byte(data))
	if err == nil {
		t.Fatal("expected error for result=error")
	}
	if !strings.Contains(err.Error(), `result=`) {
		t.Fatalf("error should mention result value, got: %v", err)
	}
}

func TestParse_EmptyRates(t *testing.T) {
	data := `{"result":"success","time_last_update_unix":0,"base_code":"USD"}`
	_, err := provider.Parse([]byte(data))
	if err == nil {
		t.Fatal("expected error for absent rates map")
	}
}

func TestParse_InvalidJSON(t *testing.T) {
	_, err := provider.Parse([]byte(`not json`))
	if err == nil {
		t.Fatal("expected error for invalid JSON")
	}
}
