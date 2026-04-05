// billsplit/internal/handler/router.go
package handler

import (
	"net/http"

	"github.com/fergalhk-lab/apps/billsplit/internal/middleware"
	"github.com/fergalhk-lab/apps/billsplit/internal/service"
)

type Services struct {
	Auth        *service.AuthService
	Groups      *service.GroupService
	Expenses    *service.ExpenseService
	Settlements *service.SettlementService
	Invites     *service.InviteService
}

func NewRouter(svc Services, secureCookie bool) http.Handler {
	mux := http.NewServeMux()

	// Auth (no JWT required)
	mux.HandleFunc("POST /api/auth/register", authRegisterHandler(svc.Auth))
	mux.HandleFunc("POST /api/auth/login", authLoginHandler(svc.Auth, secureCookie))
	mux.HandleFunc("POST /api/auth/logout", authLogoutHandler(secureCookie))

	// Users
	mux.Handle("GET /api/users", middleware.RequireAuth(svc.Auth, http.HandlerFunc(listUsersHandler(svc.Auth))))

	// Groups
	mux.Handle("POST /api/groups", middleware.RequireAuth(svc.Auth, http.HandlerFunc(createGroupHandler(svc.Groups))))
	mux.Handle("GET /api/groups", middleware.RequireAuth(svc.Auth, http.HandlerFunc(listGroupsHandler(svc.Groups))))
	mux.Handle("GET /api/groups/{id}", middleware.RequireAuth(svc.Auth, http.HandlerFunc(getGroupHandler(svc.Groups))))

	// Expenses
	mux.Handle("POST /api/groups/{id}/expenses", middleware.RequireAuth(svc.Auth, http.HandlerFunc(addExpenseHandler(svc.Expenses))))
	mux.Handle("GET /api/groups/{id}/expenses", middleware.RequireAuth(svc.Auth, http.HandlerFunc(listEventsHandler(svc.Expenses))))
	mux.Handle("DELETE /api/groups/{id}/expenses/{eventId}", middleware.RequireAuth(svc.Auth, http.HandlerFunc(cancelExpenseHandler(svc.Expenses))))

	// Settlements
	mux.Handle("POST /api/groups/{id}/settlements", middleware.RequireAuth(svc.Auth, http.HandlerFunc(addSettlementHandler(svc.Settlements))))

	// Leave group
	mux.Handle("DELETE /api/groups/{id}/members/{username}", middleware.RequireAuth(svc.Auth, http.HandlerFunc(leaveGroupHandler(svc.Groups))))

	// Admin
	mux.Handle("POST /api/admin/invites", middleware.RequireAdmin(svc.Auth, http.HandlerFunc(generateInviteHandler(svc.Invites))))

	return mux
}
