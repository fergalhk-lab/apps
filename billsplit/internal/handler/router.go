// billsplit/internal/handler/router.go
package handler

import (
	"net/http"

	"github.com/fergalhk-lab/apps/billsplit/internal/fxrates"
	"github.com/fergalhk-lab/apps/billsplit/internal/middleware"
	"github.com/fergalhk-lab/apps/billsplit/internal/service"
	"go.uber.org/zap"
)

type Services struct {
	Auth        *service.AuthService
	Groups      *service.GroupService
	Expenses    *service.ExpenseService
	Settlements *service.SettlementService
	Invites     *service.InviteService
	FXRates     *fxrates.Cache
}

func NewRouter(svc Services, logger *zap.Logger, secureCookie bool) http.Handler {
	logger = logger.Named("handler")
	mux := http.NewServeMux()

	// Probe endpoints
	mux.HandleFunc("GET /readyz", readyzHandler())

	// Auth (no JWT required)
	mux.HandleFunc("POST /api/auth/register", authRegisterHandler(svc.Auth, logger))
	mux.HandleFunc("POST /api/auth/login", authLoginHandler(svc.Auth, secureCookie))
	mux.HandleFunc("POST /api/auth/logout", authLogoutHandler(secureCookie))

	// Users
	mux.Handle("GET /api/users", middleware.RequireAuth(svc.Auth, http.HandlerFunc(listUsersHandler(svc.Auth, logger))))

	// Groups
	mux.Handle("POST /api/groups", middleware.RequireAuth(svc.Auth, http.HandlerFunc(createGroupHandler(svc.Groups, logger))))
	mux.Handle("GET /api/groups", middleware.RequireAuth(svc.Auth, http.HandlerFunc(listGroupsHandler(svc.Groups, logger))))
	mux.Handle("GET /api/groups/{id}", middleware.RequireAuth(svc.Auth, http.HandlerFunc(getGroupHandler(svc.Groups, logger))))

	// Currencies
	mux.Handle("GET /api/currencies", middleware.RequireAuth(svc.Auth, http.HandlerFunc(currenciesHandler(svc.FXRates, logger))))

	// Expenses
	mux.Handle("POST /api/groups/{id}/expenses", middleware.RequireAuth(svc.Auth, http.HandlerFunc(addExpenseHandler(svc.Expenses, svc.FXRates, logger))))
	mux.Handle("GET /api/groups/{id}/expenses", middleware.RequireAuth(svc.Auth, http.HandlerFunc(listEventsHandler(svc.Expenses, logger))))
	mux.Handle("DELETE /api/groups/{id}/expenses/{eventId}", middleware.RequireAuth(svc.Auth, http.HandlerFunc(cancelExpenseHandler(svc.Expenses))))

	// Settlements
	mux.Handle("POST /api/groups/{id}/settlements", middleware.RequireAuth(svc.Auth, http.HandlerFunc(addSettlementHandler(svc.Settlements))))

	// Leave group
	mux.Handle("DELETE /api/groups/{id}/members/{username}", middleware.RequireAuth(svc.Auth, http.HandlerFunc(leaveGroupHandler(svc.Groups))))

	// Delete group
	mux.Handle("DELETE /api/groups/{id}", middleware.RequireAuth(svc.Auth, http.HandlerFunc(deleteGroupHandler(svc.Groups, logger))))

	// Admin
	mux.Handle("POST /api/admin/invites", middleware.RequireAdmin(svc.Auth, http.HandlerFunc(generateInviteHandler(svc.Invites, logger))))

	return mux
}
