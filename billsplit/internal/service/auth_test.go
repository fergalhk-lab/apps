// billsplit/internal/service/auth_test.go
package service_test

import (
	"context"
	"testing"

	"github.com/fergalhk-lab/apps/billsplit/internal/domain"
	"github.com/fergalhk-lab/apps/billsplit/internal/service"
)

func newAuthAndInviteServices(t *testing.T) (*service.AuthService, *service.InviteService) {
	t.Helper()
	st := newTestStore(t)
	return service.NewAuthService(st, "test-secret"), service.NewInviteService(st)
}

// registerUser is a test helper: generates an invite then registers the user.
func registerUser(t *testing.T, auth *service.AuthService, invites *service.InviteService, username, password string) {
	t.Helper()
	code, err := invites.GenerateInvite(context.Background(), false)
	if err != nil {
		t.Fatalf("generate invite: %v", err)
	}
	if err := auth.Register(context.Background(), username, password, code); err != nil {
		t.Fatalf("register %s: %v", username, err)
	}
}

func TestRegister_Success(t *testing.T) {
	auth, invites := newAuthAndInviteServices(t)
	registerUser(t, auth, invites, "alice", "password123") // no error = pass
}

func TestRegister_InvalidCode(t *testing.T) {
	auth, _ := newAuthAndInviteServices(t)
	if err := auth.Register(context.Background(), "alice", "password123", "bad-code"); err == nil {
		t.Fatal("expected error for invalid invite code")
	}
}

func TestLogin_Success(t *testing.T) {
	auth, invites := newAuthAndInviteServices(t)
	registerUser(t, auth, invites, "alice", "password123")
	token, err := auth.Login(context.Background(), "alice", "password123")
	if err != nil {
		t.Fatalf("login: %v", err)
	}
	if token == "" {
		t.Fatal("expected non-empty token")
	}
}

func TestLogin_WrongPassword(t *testing.T) {
	auth, invites := newAuthAndInviteServices(t)
	registerUser(t, auth, invites, "alice", "password123")
	if _, err := auth.Login(context.Background(), "alice", "wrong"); err == nil {
		t.Fatal("expected error for wrong password")
	}
}

func TestVerifyToken(t *testing.T) {
	auth, invites := newAuthAndInviteServices(t)
	registerUser(t, auth, invites, "alice", "password123")
	token, _ := auth.Login(context.Background(), "alice", "password123")
	claims, err := auth.VerifyToken(token)
	if err != nil {
		t.Fatalf("verify: %v", err)
	}
	if claims.Username != "alice" {
		t.Errorf("want alice, got %s", claims.Username)
	}
}

func TestListUsers_ReturnsAllUsers(t *testing.T) {
	auth, invites := newAuthAndInviteServices(t)
	registerUser(t, auth, invites, "alice", "password123")
	registerUser(t, auth, invites, "bob", "password123")

	users, err := auth.ListUsers(context.Background())
	if err != nil {
		t.Fatalf("list users: %v", err)
	}
	if len(users) != 2 {
		t.Fatalf("want 2 users, got %d", len(users))
	}
	byID := make(map[string]domain.UserSummary)
	for _, u := range users {
		byID[u.ID] = u
	}
	if _, ok := byID["alice"]; !ok {
		t.Error("alice not in list")
	}
	if _, ok := byID["bob"]; !ok {
		t.Error("bob not in list")
	}
}

func TestListUsers_AdminFlagPreserved(t *testing.T) {
	auth, invites := newAuthAndInviteServices(t)
	registerUser(t, auth, invites, "alice", "password123") // non-admin

	adminCode, err := invites.GenerateInvite(context.Background(), true)
	if err != nil {
		t.Fatalf("generate admin invite: %v", err)
	}
	if err := auth.Register(context.Background(), "adminuser", "password123", adminCode); err != nil {
		t.Fatalf("register adminuser: %v", err)
	}

	users, err := auth.ListUsers(context.Background())
	if err != nil {
		t.Fatalf("list users: %v", err)
	}
	byID := make(map[string]domain.UserSummary)
	for _, u := range users {
		byID[u.ID] = u
	}
	if _, ok := byID["alice"]; !ok {
		t.Fatal("alice not in list")
	}
	if byID["alice"].IsAdmin {
		t.Error("alice should not be admin")
	}
	if !byID["adminuser"].IsAdmin {
		t.Error("adminuser should be admin")
	}
}

func TestListUsers_EmptyWhenNoUsers(t *testing.T) {
	auth, _ := newAuthAndInviteServices(t)
	users, err := auth.ListUsers(context.Background())
	if err != nil {
		t.Fatalf("list users: %v", err)
	}
	if len(users) != 0 {
		t.Fatalf("want 0 users, got %d", len(users))
	}
}
