// billsplit/internal/handler/admin.go
package handler

import (
	"net/http"

	"github.com/fergalhk-lab/apps/billsplit/internal/service"
	"go.uber.org/zap"
)

func generateInviteHandler(invites *service.InviteService, logger *zap.Logger) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req struct {
			IsAdmin bool `json:"isAdmin"`
		}
		if err := decodeJSON(r, &req); err != nil {
			writeError(w, http.StatusBadRequest, "invalid request body")
			return
		}
		code, err := invites.GenerateInvite(r.Context(), req.IsAdmin)
		if err != nil {
			logger.Error("generate invite failed", zap.Error(err))
			writeError(w, http.StatusInternalServerError, "failed to generate invite")
			return
		}
		writeJSON(w, http.StatusCreated, map[string]string{"code": code})
	}
}
