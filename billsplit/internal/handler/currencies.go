// billsplit/internal/handler/currencies.go
package handler

import (
	"net/http"

	"github.com/fergalhk-lab/apps/billsplit/internal/fxrates"
	"go.uber.org/zap"
)

func currenciesHandler(cache *fxrates.Cache, logger *zap.Logger) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		rates, err := cache.Get(r.Context())
		if err != nil {
			logger.Error("exchange rates unavailable", zap.Error(err))
			writeError(w, http.StatusServiceUnavailable, "exchange rates unavailable")
			return
		}
		writeJSON(w, http.StatusOK, map[string]interface{}{
			"base":      rates.Base,
			"rates":     rates.Rates,
			"updatedAt": rates.ProviderUpdatedAt,
		})
	}
}
