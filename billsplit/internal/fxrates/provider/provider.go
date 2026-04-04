package provider

import (
	"encoding/json"
	"fmt"
	"time"

	"github.com/fergalhk-lab/apps/billsplit/internal/fxrates"
)

// apiResponse mirrors the open.er-api.com /v6/latest/{base} JSON shape.
type apiResponse struct {
	Result             string             `json:"result"`
	TimeLastUpdateUnix int64              `json:"time_last_update_unix"`
	BaseCode           string             `json:"base_code"`
	Rates              map[string]float64 `json:"rates"`
}

// Parse maps a raw API JSON payload to the internal fxrates.Rates type.
// Returns an error if the JSON is malformed or result != "success".
func Parse(data []byte) (fxrates.Rates, error) {
	var resp apiResponse
	if err := json.Unmarshal(data, &resp); err != nil {
		return fxrates.Rates{}, fmt.Errorf("parse exchange rate response: %w", err)
	}
	if resp.Result != "success" {
		return fxrates.Rates{}, fmt.Errorf("exchange rate API returned result=%q", resp.Result)
	}
	if len(resp.Rates) == 0 {
		return fxrates.Rates{}, fmt.Errorf("exchange rate API returned empty rates map")
	}
	return fxrates.Rates{
		Base:              resp.BaseCode,
		Rates:             resp.Rates,
		ProviderUpdatedAt: time.Unix(resp.TimeLastUpdateUnix, 0).UTC(),
	}, nil
}
