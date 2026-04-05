// billsplit/internal/handler/groups.go
package handler

import (
	"errors"
	"net/http"

	"github.com/fergalhk-lab/apps/billsplit/internal/middleware"
	"github.com/fergalhk-lab/apps/billsplit/internal/service"
	"github.com/fergalhk-lab/apps/billsplit/internal/store"
)

func createGroupHandler(groups *service.GroupService) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req struct {
			Name     string   `json:"name"`
			Currency string   `json:"currency"`
			Members  []string `json:"members"`
		}
		if err := decodeJSON(r, &req); err != nil {
			writeError(w, http.StatusBadRequest, "invalid request body")
			return
		}
		username := middleware.UsernameFromCtx(r)
		id, err := groups.CreateGroup(r.Context(), username, req.Name, req.Currency, req.Members)
		if err != nil {
			if errors.Is(err, service.ErrUnknownMembers) {
				writeError(w, http.StatusBadRequest, err.Error())
				return
			}
			if errors.Is(err, service.ErrDuplicateMembers) {
				writeError(w, http.StatusBadRequest, err.Error())
				return
			}
			writeError(w, http.StatusInternalServerError, "failed to create group")
			return
		}
		writeJSON(w, http.StatusCreated, map[string]string{"id": id})
	}
}

func listGroupsHandler(groups *service.GroupService) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		username := middleware.UsernameFromCtx(r)
		list, err := groups.ListGroups(r.Context(), username)
		if err != nil {
			writeError(w, http.StatusInternalServerError, "failed to list groups")
			return
		}
		writeJSON(w, http.StatusOK, list)
	}
}

func getGroupHandler(groups *service.GroupService) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		groupID := r.PathValue("id")
		detail, err := groups.GetGroup(r.Context(), groupID)
		if err != nil {
			if errors.Is(err, store.ErrNotFound) {
				writeError(w, http.StatusNotFound, "group not found")
				return
			}
			writeError(w, http.StatusInternalServerError, "failed to get group")
			return
		}
		writeJSON(w, http.StatusOK, detail)
	}
}

func leaveGroupHandler(groups *service.GroupService) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		groupID := r.PathValue("id")
		username := r.PathValue("username")
		if username != middleware.UsernameFromCtx(r) {
			writeError(w, http.StatusForbidden, "can only remove yourself from a group")
			return
		}
		if err := groups.LeaveGroup(r.Context(), groupID, username); err != nil {
			if errors.Is(err, store.ErrNotFound) {
				writeError(w, http.StatusNotFound, "group not found")
				return
			}
			writeError(w, http.StatusConflict, err.Error())
			return
		}
		w.WriteHeader(http.StatusNoContent)
	}
}
