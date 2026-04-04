// billsplit/internal/service/auth_test.go
package service_test

import (
	"context"
	"testing"

	"github.com/fergalhk-lab/apps/billsplit/internal/domain"
	"github.com/fergalhk-lab/apps/billsplit/internal/service"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap/zaptest"
)

func newAuthAndInviteServices(t *testing.T) (*service.AuthService, *service.InviteService) {
	t.Helper()
	st := newTestStore(t)
	return service.NewAuthService(st, "test-secret", zaptest.NewLogger(t)), service.NewInviteService(st, zaptest.NewLogger(t))
}

// registerUser is a test helper: generates an invite then registers the user.
func registerUser(t *testing.T, auth *service.AuthService, invites *service.InviteService, username, password string) {
	t.Helper()
	code, err := invites.GenerateInvite(context.Background(), false)
	require.NoError(t, err, "generate invite: %v", err)
	err = auth.Register(context.Background(), username, password, code)
	require.NoError(t, err, "register %s: %v", username, err)
}

func TestRegister_Success(t *testing.T) {
	auth, invites := newAuthAndInviteServices(t)
	registerUser(t, auth, invites, "alice", "password123") // no error = pass
}

func TestRegister_InvalidCode(t *testing.T) {
	auth, _ := newAuthAndInviteServices(t)
	err := auth.Register(context.Background(), "alice", "password123", "bad-code")
	require.Error(t, err, "expected error for invalid invite code")
}

func TestLogin_Success(t *testing.T) {
	auth, invites := newAuthAndInviteServices(t)
	registerUser(t, auth, invites, "alice", "password123")
	token, _, err := auth.Login(context.Background(), "alice", "password123")
	require.NoError(t, err, "login: %v", err)
	require.NotEmpty(t, token, "expected non-empty token")
}

func TestLogin_WrongPassword(t *testing.T) {
	auth, invites := newAuthAndInviteServices(t)
	registerUser(t, auth, invites, "alice", "password123")
	_, _, err := auth.Login(context.Background(), "alice", "wrong")
	require.Error(t, err, "expected error for wrong password")
}

func TestVerifyToken(t *testing.T) {
	auth, invites := newAuthAndInviteServices(t)
	registerUser(t, auth, invites, "alice", "password123")
	token, _, _ := auth.Login(context.Background(), "alice", "password123")
	claims, err := auth.VerifyToken(token)
	require.NoError(t, err, "verify: %v", err)
	assert.Equal(t, "alice", claims.Username)
}

func TestListUsers_ReturnsAllUsers(t *testing.T) {
	auth, invites := newAuthAndInviteServices(t)
	registerUser(t, auth, invites, "alice", "password123")
	registerUser(t, auth, invites, "bob", "password123")

	users, err := auth.ListUsers(context.Background())
	require.NoError(t, err, "list users: %v", err)
	require.Len(t, users, 2, "want 2 users, got %d", len(users))
	byID := make(map[string]domain.UserSummary)
	for _, u := range users {
		byID[u.ID] = u
	}
	_, aliceOk := byID["alice"]
	assert.True(t, aliceOk, "alice not in list")
	_, bobOk := byID["bob"]
	assert.True(t, bobOk, "bob not in list")
}

func TestListUsers_AdminFlagPreserved(t *testing.T) {
	auth, invites := newAuthAndInviteServices(t)
	registerUser(t, auth, invites, "alice", "password123") // non-admin

	adminCode, err := invites.GenerateInvite(context.Background(), true)
	require.NoError(t, err, "generate admin invite: %v", err)
	err = auth.Register(context.Background(), "adminuser", "password123", adminCode)
	require.NoError(t, err, "register adminuser: %v", err)

	users, err := auth.ListUsers(context.Background())
	require.NoError(t, err, "list users: %v", err)
	byID := make(map[string]domain.UserSummary)
	for _, u := range users {
		byID[u.ID] = u
	}
	_, aliceOk := byID["alice"]
	require.True(t, aliceOk, "alice not in list")
	assert.False(t, byID["alice"].IsAdmin, "alice should not be admin")
	assert.True(t, byID["adminuser"].IsAdmin, "adminuser should be admin")
}

func TestListUsers_EmptyWhenNoUsers(t *testing.T) {
	auth, _ := newAuthAndInviteServices(t)
	users, err := auth.ListUsers(context.Background())
	require.NoError(t, err, "list users: %v", err)
	require.Len(t, users, 0, "want 0 users, got %d", len(users))
}
