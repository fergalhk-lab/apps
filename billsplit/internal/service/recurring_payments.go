package service

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"

	"github.com/fergalhk-lab/apps/billsplit/internal/domain"
	"github.com/fergalhk-lab/apps/billsplit/internal/store"
	"github.com/google/uuid"
	"go.uber.org/zap"
)

var ErrRecurringPaymentNotFound = errors.New("recurring payment not found")
var ErrNotGroupMember = errors.New("user is not a member of this group")

type RecurringPaymentService struct {
	store  store.Store
	logger *zap.Logger
}

func NewRecurringPaymentService(s store.Store, logger *zap.Logger) *RecurringPaymentService {
	return &RecurringPaymentService{store: s, logger: logger.Named("service.recurring_payments")}
}

func (rs *RecurringPaymentService) ListRecurringPayments(ctx context.Context, groupID string) ([]domain.RecurringPayment, error) {
	data, _, err := rs.store.ReadObject(ctx, groupKey(groupID))
	if err != nil {
		return nil, err
	}
	var g domain.Group
	if err := json.Unmarshal(data, &g); err != nil {
		return nil, err
	}
	if g.RecurringPayments == nil {
		return []domain.RecurringPayment{}, nil
	}
	return g.RecurringPayments, nil
}

func (rs *RecurringPaymentService) CreateRecurringPayment(ctx context.Context, groupID, callerUsername string, rp domain.RecurringPayment) (string, error) {
	rp.ID = uuid.New().String()
	return rp.ID, withRetry(ctx, rs.store, groupKey(groupID), rs.logger, func(data []byte) ([]byte, error) {
		if data == nil {
			return nil, store.ErrNotFound
		}
		var g domain.Group
		if err := json.Unmarshal(data, &g); err != nil {
			return nil, err
		}
		if err := requireMember(g, callerUsername); err != nil {
			return nil, err
		}
		if err := domain.ValidateSplits(rp.Amount, rp.Splits, g.Members); err != nil {
			return nil, err
		}
		g.RecurringPayments = append(g.RecurringPayments, rp)
		return json.Marshal(g)
	})
}

func requireMember(g domain.Group, username string) error {
	for _, m := range g.Members {
		if m == username {
			return nil
		}
	}
	return fmt.Errorf("%w: %s", ErrNotGroupMember, username)
}
