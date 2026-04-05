// billsplit/internal/middleware/auth.go
package middleware

import (
	"context"
	"encoding/json"
	"net/http"

	"github.com/fergalhk-lab/apps/billsplit/internal/service"
)

type contextKey string

const (
	ctxUsername       contextKey = "username"
	ctxIsAdmin        contextKey = "isAdmin"
	SessionCookieName            = "session"
)

func UsernameFromCtx(r *http.Request) string {
	v, _ := r.Context().Value(ctxUsername).(string)
	return v
}

func IsAdminFromCtx(r *http.Request) bool {
	v, _ := r.Context().Value(ctxIsAdmin).(bool)
	return v
}

func RequireAuth(auth *service.AuthService, next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		cookie, err := r.Cookie(SessionCookieName)
		if err != nil {
			writeError(w, http.StatusUnauthorized, "missing session cookie")
			return
		}
		claims, err := auth.VerifyToken(cookie.Value)
		if err != nil {
			writeError(w, http.StatusUnauthorized, "invalid token")
			return
		}
		ctx := context.WithValue(r.Context(), ctxUsername, claims.Username)
		ctx = context.WithValue(ctx, ctxIsAdmin, claims.IsAdmin)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

func RequireAdmin(auth *service.AuthService, next http.Handler) http.Handler {
	return RequireAuth(auth, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if !IsAdminFromCtx(r) {
			writeError(w, http.StatusForbidden, "admin access required")
			return
		}
		next.ServeHTTP(w, r)
	}))
}

func writeError(w http.ResponseWriter, status int, msg string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(map[string]string{"error": msg})
}
