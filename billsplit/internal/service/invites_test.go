// billsplit/internal/service/invites_test.go
package service_test

import (
	"context"
	"testing"

	"github.com/fergalhk-lab/apps/billsplit/internal/service"
)

func TestGenerateInvite(t *testing.T) {
	st := newTestStore(t)
	auth := service.NewAuthService(st, "secret")
	invites := service.NewInviteService(st)
	ctx := context.Background()

	// Generate invite, use it to register
	code, err := invites.GenerateInvite(ctx, false)
	if err != nil {
		t.Fatalf("generate invite: %v", err)
	}
	if code == "" {
		t.Fatal("expected non-empty code")
	}

	// New user can register with generated code
	if err := auth.Register(ctx, "bob", "pw", code); err != nil {
		t.Fatalf("register with generated invite: %v", err)
	}
}
