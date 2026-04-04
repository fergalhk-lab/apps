// billsplit/internal/handler/admin.go
package handler

import (
	"net/http"

	"github.com/fergalhk-lab/apps/billsplit/internal/service"
)

func generateInviteHandler(invites *service.InviteService) http.HandlerFunc {
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
			writeError(w, http.StatusInternalServerError, "failed to generate invite")
			return
		}
		writeJSON(w, http.StatusCreated, map[string]string{"code": code})
	}
}
