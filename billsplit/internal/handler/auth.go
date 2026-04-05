// billsplit/internal/handler/auth.go
package handler

import (
	"errors"
	"net/http"

	"github.com/fergalhk-lab/apps/billsplit/internal/service"
)

func authRegisterHandler(auth *service.AuthService) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req struct {
			Username   string `json:"username"`
			Password   string `json:"password"`
			InviteCode string `json:"inviteCode"`
		}
		if err := decodeJSON(r, &req); err != nil {
			writeError(w, http.StatusBadRequest, "invalid request body")
			return
		}
		if err := auth.Register(r.Context(), req.Username, req.Password, req.InviteCode); err != nil {
			switch {
			case errors.Is(err, service.ErrInvalidInvite):
				writeError(w, http.StatusBadRequest, err.Error())
			case errors.Is(err, service.ErrUsernameTaken):
				writeError(w, http.StatusConflict, err.Error())
			default:
				writeError(w, http.StatusInternalServerError, "registration failed")
			}
			return
		}
		w.WriteHeader(http.StatusNoContent)
	}
}

func authLoginHandler(auth *service.AuthService, secureCookie bool) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req struct {
			Username string `json:"username"`
			Password string `json:"password"`
		}
		if err := decodeJSON(r, &req); err != nil {
			writeError(w, http.StatusBadRequest, "invalid request body")
			return
		}
		token, err := auth.Login(r.Context(), req.Username, req.Password)
		if err != nil {
			writeError(w, http.StatusUnauthorized, "invalid credentials")
			return
		}
		claims, err := auth.VerifyToken(token)
		if err != nil {
			writeError(w, http.StatusInternalServerError, "token error")
			return
		}
		http.SetCookie(w, &http.Cookie{
			Name:     "session",
			Value:    token,
			HttpOnly: true,
			SameSite: http.SameSiteStrictMode,
			Secure:   secureCookie,
			Path:     "/",
			MaxAge:   86400,
		})
		writeJSON(w, http.StatusOK, map[string]interface{}{
			"username": claims.Username,
			"isAdmin":  claims.IsAdmin,
		})
	}
}

func listUsersHandler(auth *service.AuthService) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		users, err := auth.ListUsers(r.Context())
		if err != nil {
			writeError(w, http.StatusInternalServerError, "failed to list users")
			return
		}
		writeJSON(w, http.StatusOK, map[string]interface{}{"users": users})
	}
}
