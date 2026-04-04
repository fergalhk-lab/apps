// billsplit/internal/handler/router.go
package handler

import (
	"net/http"

	"github.com/fergalhk-lab/apps/billsplit/internal/service"
)

type Services struct {
	Auth        *service.AuthService
	Groups      *service.GroupService
	Expenses    *service.ExpenseService
	Settlements *service.SettlementService
	Invites     *service.InviteService
}

func NewRouter(svc Services) http.Handler {
	mux := http.NewServeMux()

	// Auth (no JWT required)
	mux.HandleFunc("POST /api/auth/register", authRegisterHandler(svc.Auth))
	mux.HandleFunc("POST /api/auth/login", authLoginHandler(svc.Auth))

	// Users
	mux.Handle("GET /api/users", RequireAuth(svc.Auth, http.HandlerFunc(listUsersHandler(svc.Auth))))

	// Groups (JWT required)
	mux.Handle("POST /api/groups", RequireAuth(svc.Auth, http.HandlerFunc(createGroupHandler(svc.Groups))))
	mux.Handle("GET /api/groups", RequireAuth(svc.Auth, http.HandlerFunc(listGroupsHandler(svc.Groups))))
	mux.Handle("GET /api/groups/{id}", RequireAuth(svc.Auth, http.HandlerFunc(getGroupHandler(svc.Groups))))

	// Expenses
	mux.Handle("POST /api/groups/{id}/expenses", RequireAuth(svc.Auth, http.HandlerFunc(addExpenseHandler(svc.Expenses))))
	mux.Handle("GET /api/groups/{id}/expenses", RequireAuth(svc.Auth, http.HandlerFunc(listEventsHandler(svc.Expenses))))
	mux.Handle("DELETE /api/groups/{id}/expenses/{eventId}", RequireAuth(svc.Auth, http.HandlerFunc(cancelExpenseHandler(svc.Expenses))))

	// Settlements
	mux.Handle("POST /api/groups/{id}/settlements", RequireAuth(svc.Auth, http.HandlerFunc(addSettlementHandler(svc.Settlements))))

	// Leave group
	mux.Handle("DELETE /api/groups/{id}/members/{username}", RequireAuth(svc.Auth, http.HandlerFunc(leaveGroupHandler(svc.Groups))))

	// Admin
	mux.Handle("POST /api/admin/invites", RequireAdmin(svc.Auth, http.HandlerFunc(generateInviteHandler(svc.Invites))))

	return mux
}
