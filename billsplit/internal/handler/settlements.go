// billsplit/internal/handler/settlements.go
package handler

import (
	"net/http"

	"github.com/fergalhk-lab/apps/billsplit/internal/middleware"
	"github.com/fergalhk-lab/apps/billsplit/internal/service"
)

func addSettlementHandler(settlements *service.SettlementService) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		groupID := r.PathValue("id")
		var req struct {
			From   string  `json:"from"`
			To     string  `json:"to"`
			Amount float64 `json:"amount"`
		}
		if err := decodeJSON(r, &req); err != nil {
			writeError(w, http.StatusBadRequest, "invalid request body")
			return
		}
		username := middleware.UsernameFromCtx(r)
		if err := settlements.AddSettlement(r.Context(), groupID, username, req.From, req.To, req.Amount); err != nil {
			writeError(w, http.StatusBadRequest, err.Error())
			return
		}
		w.WriteHeader(http.StatusNoContent)
	}
}
