// internal/handler/auth_test.go
package handler_test

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/fergalhk-lab/apps/billsplit/internal/fxrates"
	"github.com/fergalhk-lab/apps/billsplit/internal/handler"
	"github.com/fergalhk-lab/apps/billsplit/internal/middleware"
	"github.com/fergalhk-lab/apps/billsplit/internal/service"
	"github.com/fergalhk-lab/apps/billsplit/internal/testutil"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap/zaptest"
)

// newTestRouter registers alice and returns a router with secureCookie=false
// (httptest doesn't use HTTPS).
func newTestRouter(t *testing.T) http.Handler {
	t.Helper()
	st := testutil.NewTestStore(t)
	auth := service.NewAuthService(st, "test-secret", zaptest.NewLogger(t))
	invites := service.NewInviteService(st, zaptest.NewLogger(t))
	code, err := invites.GenerateInvite(context.Background(), false)
	require.NoError(t, err)
	require.NoError(t, auth.Register(context.Background(), "alice", "password123", code))

	svc := handler.Services{
		Auth:        auth,
		Groups:      service.NewGroupService(st, zaptest.NewLogger(t)),
		Expenses:    service.NewExpenseService(st, zaptest.NewLogger(t)),
		Settlements: service.NewSettlementService(st, zaptest.NewLogger(t)),
		Invites:     invites,
		FXRates:     fxrates.NewCache(st, zaptest.NewLogger(t)),
	}
	return handler.NewRouter(svc, zaptest.NewLogger(t), false)
}

func sessionCookie(rr *httptest.ResponseRecorder) *http.Cookie {
	for _, c := range rr.Result().Cookies() {
		if c.Name == middleware.SessionCookieName {
			return c
		}
	}
	return nil
}

func TestLoginHandler_SetsCookieAndReturnsIdentity(t *testing.T) {
	router := newTestRouter(t)

	body := `{"username":"alice","password":"password123"}`
	req := httptest.NewRequest(http.MethodPost, "/api/auth/login", bytes.NewBufferString(body))
	req.Header.Set("Content-Type", "application/json")
	rr := httptest.NewRecorder()
	router.ServeHTTP(rr, req)

	require.Equal(t, http.StatusOK, rr.Code)

	cookie := sessionCookie(rr)
	require.NotNil(t, cookie, "expected a session cookie in response")
	assert.True(t, cookie.HttpOnly, "session cookie must be HttpOnly")
	assert.NotEmpty(t, cookie.Value, "session cookie value must not be empty")
	assert.Equal(t, 86400, cookie.MaxAge)

	var resp map[string]interface{}
	require.NoError(t, json.NewDecoder(rr.Body).Decode(&resp))
	assert.Equal(t, "alice", resp["username"])
	assert.Equal(t, false, resp["isAdmin"])
}

func TestLoginHandler_InvalidCredentials_Returns401(t *testing.T) {
	router := newTestRouter(t)

	body := `{"username":"alice","password":"wrong"}`
	req := httptest.NewRequest(http.MethodPost, "/api/auth/login", bytes.NewBufferString(body))
	req.Header.Set("Content-Type", "application/json")
	rr := httptest.NewRecorder()
	router.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusUnauthorized, rr.Code)
	assert.Nil(t, sessionCookie(rr), "no cookie should be set on failed login")
}

func TestLogoutHandler_ClearsCookieAndReturns204(t *testing.T) {
	router := newTestRouter(t)

	req := httptest.NewRequest(http.MethodPost, "/api/auth/logout", nil)
	rr := httptest.NewRecorder()
	router.ServeHTTP(rr, req)

	require.Equal(t, http.StatusNoContent, rr.Code)

	cookie := sessionCookie(rr)
	require.NotNil(t, cookie, "expected session cookie to be set in response (clearing it)")
	assert.Equal(t, -1, cookie.MaxAge, "session cookie MaxAge should be -1 to delete it")
	assert.Equal(t, "", cookie.Value, "session cookie value should be empty")
}
