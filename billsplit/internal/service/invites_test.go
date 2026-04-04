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

func TestHasInvites(t *testing.T) {
	st := newTestStore(t)
	invites := service.NewInviteService(st)
	ctx := context.Background()

	// No invites on a fresh store
	has, err := invites.HasInvites(ctx)
	if err != nil {
		t.Fatalf("HasInvites on empty store: %v", err)
	}
	if has {
		t.Fatal("expected false on empty store, got true")
	}

	// Generate one invite
	if _, err := invites.GenerateInvite(ctx, false); err != nil {
		t.Fatalf("GenerateInvite: %v", err)
	}

	// Now should have invites
	has, err = invites.HasInvites(ctx)
	if err != nil {
		t.Fatalf("HasInvites after generate: %v", err)
	}
	if !has {
		t.Fatal("expected true after generating an invite, got false")
	}
}
