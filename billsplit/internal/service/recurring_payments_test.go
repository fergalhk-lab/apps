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

func setupRPTest(t *testing.T) (*service.AuthService, *service.InviteService, *service.GroupService, *service.RecurringPaymentService) {
	t.Helper()
	st := newTestStore(t)
	auth := service.NewAuthService(st, "secret", zaptest.NewLogger(t))
	invites := service.NewInviteService(st, zaptest.NewLogger(t))
	groups := service.NewGroupService(st, zaptest.NewLogger(t))
	rps := service.NewRecurringPaymentService(st, zaptest.NewLogger(t))
	return auth, invites, groups, rps
}

var testRP = domain.RecurringPayment{
	Description: "Monthly rent",
	Amount:      100.0,
	MemberID:    "alice",
	Splits:      map[string]float64{"alice": 50.0, "bob": 50.0},
	Cadence:     domain.CadenceMonthly,
}

func TestCreateRecurringPayment(t *testing.T) {
	auth, invites, groups, rps := setupRPTest(t)
	ctx := context.Background()
	groupID := registerAndCreateGroup(t, auth, invites, groups)

	rpID, err := rps.CreateRecurringPayment(ctx, groupID, "alice", testRP)
	require.NoError(t, err)
	require.NotEmpty(t, rpID)
}

func TestListRecurringPayments(t *testing.T) {
	auth, invites, groups, rps := setupRPTest(t)
	ctx := context.Background()
	groupID := registerAndCreateGroup(t, auth, invites, groups)

	rpID, err := rps.CreateRecurringPayment(ctx, groupID, "alice", testRP)
	require.NoError(t, err)

	list, err := rps.ListRecurringPayments(ctx, groupID)
	require.NoError(t, err)
	require.Len(t, list, 1)
	assert.Equal(t, rpID, list[0].ID)
	assert.Equal(t, "Monthly rent", list[0].Description)
	assert.Equal(t, domain.CadenceMonthly, list[0].Cadence)
}

func TestListRecurringPayments_Empty(t *testing.T) {
	auth, invites, groups, rps := setupRPTest(t)
	ctx := context.Background()
	groupID := registerAndCreateGroup(t, auth, invites, groups)

	list, err := rps.ListRecurringPayments(ctx, groupID)
	require.NoError(t, err)
	require.Empty(t, list)
}

func TestCreateRecurringPayment_InvalidSplits(t *testing.T) {
	auth, invites, groups, rps := setupRPTest(t)
	ctx := context.Background()
	groupID := registerAndCreateGroup(t, auth, invites, groups)

	_, err := rps.CreateRecurringPayment(ctx, groupID, "alice", domain.RecurringPayment{
		Description: "Rent",
		Amount:      100.0,
		MemberID:    "alice",
		Splits:      map[string]float64{"alice": 40.0, "bob": 40.0},
		Cadence:     domain.CadenceMonthly,
	})
	require.Error(t, err)
}

func TestCreateRecurringPayment_NonMember(t *testing.T) {
	auth, invites, groups, rps := setupRPTest(t)
	ctx := context.Background()
	groupID := registerAndCreateGroup(t, auth, invites, groups)

	_, err := rps.CreateRecurringPayment(ctx, groupID, "charlie", testRP)
	require.ErrorIs(t, err, service.ErrNotGroupMember)
}
