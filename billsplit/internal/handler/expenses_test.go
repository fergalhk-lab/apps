// internal/handler/expenses_test.go
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
	"github.com/fergalhk-lab/apps/billsplit/internal/service"
	"github.com/fergalhk-lab/apps/billsplit/internal/testutil"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// newTestRouterWithFXRates creates a router with an fxrates cache seeded with
// the given rates (USD-based, so USD=1.0).
func newTestRouterWithFXRates(t *testing.T, rates map[string]float64) (http.Handler, string) {
	t.Helper()
	st := testutil.NewTestStore(t)
	ctx := context.Background()

	auth := service.NewAuthService(st, "test-secret")
	invites := service.NewInviteService(st)
	groups := service.NewGroupService(st)

	// Register alice and bob, create a EUR group
	codeA, err := invites.GenerateInvite(ctx, false)
	require.NoError(t, err)
	require.NoError(t, auth.Register(ctx, "alice", "password123", codeA))

	codeB, err := invites.GenerateInvite(ctx, false)
	require.NoError(t, err)
	require.NoError(t, auth.Register(ctx, "bob", "password123", codeB))

	groupID, err := groups.CreateGroup(ctx, "alice", "Trip", "EUR", []string{"bob"})
	require.NoError(t, err)

	// Seed exchange rates into store
	ratesData := fxrates.Rates{Base: "USD", Rates: rates}
	raw, err := json.Marshal(ratesData)
	require.NoError(t, err)
	require.NoError(t, st.ForceWriteObject(ctx, fxrates.S3Key, raw))

	fxCache := fxrates.NewCache(st)
	svc := handler.Services{
		Auth:        auth,
		Groups:      groups,
		Expenses:    service.NewExpenseService(st),
		Settlements: service.NewSettlementService(st),
		Invites:     invites,
		FXRates:     fxCache,
	}
	return handler.NewRouter(svc, false), groupID
}

func loginAs(t *testing.T, router http.Handler, username string) *http.Cookie {
	t.Helper()
	body, _ := json.Marshal(map[string]string{"username": username, "password": "password123"})
	req := httptest.NewRequest(http.MethodPost, "/api/auth/login", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	rr := httptest.NewRecorder()
	router.ServeHTTP(rr, req)
	require.Equal(t, http.StatusOK, rr.Code)
	cookie := sessionCookie(rr)
	require.NotNil(t, cookie)
	return cookie
}

// TestAddExpense_CrossCurrency verifies that when submitting an expense in a
// currency different from the group's base currency, both the total and the
// splits are converted, so validation passes and the expense is created.
func TestAddExpense_CrossCurrency(t *testing.T) {
	// USD=1.0, EUR=0.9: 100 USD → ~111.11 EUR; splits 50/50 USD → ~55.56/55.56 EUR
	router, groupID := newTestRouterWithFXRates(t, map[string]float64{
		"USD": 1.0,
		"EUR": 0.9,
	})

	cookie := loginAs(t, router, "alice")

	body, _ := json.Marshal(map[string]interface{}{
		"description": "Dinner",
		"amount":      100.0,
		"currency":    "USD",
		"paidBy":      "alice",
		"splits": map[string]float64{
			"alice": 50.0,
			"bob":   50.0,
		},
	})
	req := httptest.NewRequest(http.MethodPost, "/api/groups/"+groupID+"/expenses", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	req.AddCookie(cookie)

	rr := httptest.NewRecorder()
	router.ServeHTTP(rr, req)

	require.Equal(t, http.StatusCreated, rr.Code, "expected 201, got %d: %s", rr.Code, rr.Body.String())

	var resp map[string]string
	require.NoError(t, json.NewDecoder(rr.Body).Decode(&resp))
	assert.NotEmpty(t, resp["id"])
}
