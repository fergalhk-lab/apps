// billsplit/internal/handler/groups.go
package handler

import (
	"errors"
	"net/http"

	"github.com/fergalhk-lab/apps/billsplit/internal/middleware"
	"github.com/fergalhk-lab/apps/billsplit/internal/service"
	"github.com/fergalhk-lab/apps/billsplit/internal/store"
	"go.uber.org/zap"
)

func createGroupHandler(groups *service.GroupService, logger *zap.Logger) http.HandlerFunc {
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
			logger.Error("create group failed", zap.Error(err))
			writeError(w, http.StatusInternalServerError, "failed to create group")
			return
		}
		writeJSON(w, http.StatusCreated, map[string]string{"id": id})
	}
}

func listGroupsHandler(groups *service.GroupService, logger *zap.Logger) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		username := middleware.UsernameFromCtx(r)
		list, err := groups.ListGroups(r.Context(), username)
		if err != nil {
			logger.Error("list groups failed", zap.Error(err))
			writeError(w, http.StatusInternalServerError, "failed to list groups")
			return
		}
		writeJSON(w, http.StatusOK, list)
	}
}

func getGroupHandler(groups *service.GroupService, logger *zap.Logger) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		groupID := r.PathValue("id")
		detail, err := groups.GetGroup(r.Context(), groupID)
		if err != nil {
			if errors.Is(err, store.ErrNotFound) {
				writeError(w, http.StatusNotFound, "group not found")
				return
			}
			logger.Error("get group failed", zap.Error(err))
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

func deleteGroupHandler(groups *service.GroupService, logger *zap.Logger) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		groupID := r.PathValue("id")
		if err := groups.DeleteGroup(r.Context(), groupID); err != nil {
			if errors.Is(err, store.ErrNotFound) {
				writeError(w, http.StatusNotFound, "group not found")
				return
			}
			if errors.Is(err, service.ErrOutstandingBalances) {
				writeError(w, http.StatusConflict, err.Error())
				return
			}
			logger.Error("delete group failed", zap.Error(err))
			writeError(w, http.StatusInternalServerError, "failed to delete group")
			return
		}
		w.WriteHeader(http.StatusNoContent)
	}
}
