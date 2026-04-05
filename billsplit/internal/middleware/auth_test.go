// internal/middleware/auth_test.go
package middleware_test

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/fergalhk-lab/apps/billsplit/internal/middleware"
	"github.com/fergalhk-lab/apps/billsplit/internal/service"
	"github.com/fergalhk-lab/apps/billsplit/internal/testutil"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func setupAuth(t *testing.T) (*service.AuthService, string) {
	t.Helper()
	st := testutil.NewTestStore(t)
	auth := service.NewAuthService(st, "test-secret")
	invites := service.NewInviteService(st)
	code, err := invites.GenerateInvite(context.Background(), false)
	require.NoError(t, err)
	require.NoError(t, auth.Register(context.Background(), "alice", "password123", code))
	token, _, err := auth.Login(context.Background(), "alice", "password123")
	require.NoError(t, err)
	return auth, token
}

func okHandler() http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})
}

func TestRequireAuth_ValidCookie_Passes(t *testing.T) {
	auth, token := setupAuth(t)
	handler := middleware.RequireAuth(auth, okHandler())

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.AddCookie(&http.Cookie{Name: middleware.SessionCookieName, Value: token})
	rr := httptest.NewRecorder()
	handler.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusOK, rr.Code)
}

func TestRequireAuth_ValidCookie_SetsUsernameInContext(t *testing.T) {
	auth, token := setupAuth(t)
	var gotUsername string
	handler := middleware.RequireAuth(auth, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotUsername = middleware.UsernameFromCtx(r)
		w.WriteHeader(http.StatusOK)
	}))

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.AddCookie(&http.Cookie{Name: middleware.SessionCookieName, Value: token})
	rr := httptest.NewRecorder()
	handler.ServeHTTP(rr, req)

	require.Equal(t, http.StatusOK, rr.Code)
	assert.Equal(t, "alice", gotUsername)
}

func TestRequireAuth_NoCookie_Returns401(t *testing.T) {
	auth, _ := setupAuth(t)
	handler := middleware.RequireAuth(auth, okHandler())

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rr := httptest.NewRecorder()
	handler.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusUnauthorized, rr.Code)
}

func TestRequireAuth_InvalidTokenInCookie_Returns401(t *testing.T) {
	auth, _ := setupAuth(t)
	handler := middleware.RequireAuth(auth, okHandler())

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.AddCookie(&http.Cookie{Name: middleware.SessionCookieName, Value: "not-a-valid-jwt"})
	rr := httptest.NewRecorder()
	handler.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusUnauthorized, rr.Code)
}

func TestRequireAuth_BearerHeader_Returns401(t *testing.T) {
	// Bearer tokens must no longer be accepted — cookie only.
	auth, token := setupAuth(t)
	handler := middleware.RequireAuth(auth, okHandler())

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	rr := httptest.NewRecorder()
	handler.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusUnauthorized, rr.Code)
}
