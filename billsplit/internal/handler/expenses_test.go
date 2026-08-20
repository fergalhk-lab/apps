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
	"go.uber.org/zap/zaptest"
)

// newTestRouterWithFXRates creates a router with an fxrates cache seeded with
// the given rates (USD-based, so USD=1.0).
func newTestRouterWithFXRates(t *testing.T, rates map[string]float64) (http.Handler, string) {
	t.Helper()
	st := testutil.NewTestStore(t)
	ctx := context.Background()

	auth := service.NewAuthService(st, "test-secret", testAuthExpiry, zaptest.NewLogger(t))
	invites := service.NewInviteService(st, zaptest.NewLogger(t))
	groups := service.NewGroupService(st, zaptest.NewLogger(t))

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

	fxCache := fxrates.NewCache(st, zaptest.NewLogger(t))
	svc := handler.Services{
		Auth:        auth,
		Groups:      groups,
		Expenses:    service.NewExpenseService(st, zaptest.NewLogger(t)),
		Settlements: service.NewSettlementService(st, zaptest.NewLogger(t)),
		Invites:     invites,
		FXRates:     fxCache,
	}
	return handler.NewRouter(svc, zaptest.NewLogger(t), false), groupID
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

// addExpense creates an expense via the API and returns its event ID.
func addExpense(t *testing.T, router http.Handler, cookie *http.Cookie, groupID, description string) string {
	t.Helper()
	body, _ := json.Marshal(map[string]interface{}{
		"description": description,
		"amount":      100.0,
		"paidBy":      "alice",
		"splits":      map[string]float64{"alice": 50.0, "bob": 50.0},
	})
	req := httptest.NewRequest(http.MethodPost, "/api/groups/"+groupID+"/expenses", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	req.AddCookie(cookie)
	rr := httptest.NewRecorder()
	router.ServeHTTP(rr, req)
	require.Equal(t, http.StatusCreated, rr.Code, "add expense: %s", rr.Body.String())

	var resp map[string]string
	require.NoError(t, json.NewDecoder(rr.Body).Decode(&resp))
	return resp["id"]
}

// listEvents calls the list endpoint with the given raw query string and returns
// the decoded event count and reported total.
func listEvents(t *testing.T, router http.Handler, cookie *http.Cookie, groupID, query string) (int, int) {
	t.Helper()
	req := httptest.NewRequest(http.MethodGet, "/api/groups/"+groupID+"/expenses"+query, nil)
	req.AddCookie(cookie)
	rr := httptest.NewRecorder()
	router.ServeHTTP(rr, req)
	require.Equal(t, http.StatusOK, rr.Code, "list events: %s", rr.Body.String())

	var resp struct {
		Events []map[string]interface{} `json:"events"`
		Total  int                      `json:"total"`
	}
	require.NoError(t, json.NewDecoder(rr.Body).Decode(&resp))
	return len(resp.Events), resp.Total
}

// TestListEvents_ReversedQueryParam verifies the reversed flag is plumbed from
// the query string through to the service, and that anything unparseable falls
// back to the safe default of hiding cancelled expenses.
func TestListEvents_ReversedQueryParam(t *testing.T) {
	router, groupID := newTestRouterWithFXRates(t, map[string]float64{"USD": 1.0, "EUR": 0.9})
	cookie := loginAs(t, router, "alice")

	addExpense(t, router, cookie, groupID, "Kept")
	cancelled := addExpense(t, router, cookie, groupID, "Cancelled")

	req := httptest.NewRequest(http.MethodDelete, "/api/groups/"+groupID+"/expenses/"+cancelled, nil)
	req.AddCookie(cookie)
	rr := httptest.NewRecorder()
	router.ServeHTTP(rr, req)
	require.Equal(t, http.StatusNoContent, rr.Code, "cancel expense: %s", rr.Body.String())

	tests := []struct {
		name  string
		query string
		want  int
	}{
		{name: "defaults to hiding reversed", query: "", want: 1},
		{name: "explicit false hides reversed", query: "?reversed=false", want: 1},
		{name: "true returns the full log", query: "?reversed=true", want: 3},
		{name: "unparseable falls back to false", query: "?reversed=banana", want: 1},
		{name: "empty value falls back to false", query: "?reversed=", want: 1},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			count, total := listEvents(t, router, cookie, groupID, tc.query)
			assert.Equal(t, tc.want, count, "unexpected event count for query %q", tc.query)
			assert.Equal(t, tc.want, total, "total should match the events returned for query %q", tc.query)
		})
	}
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
