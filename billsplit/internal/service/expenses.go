// billsplit/internal/service/expenses.go
package service

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"github.com/fergalhk-lab/apps/billsplit/internal/domain"
	"github.com/fergalhk-lab/apps/billsplit/internal/store"
	"github.com/google/uuid"
)

var ErrEventNotFound = errors.New("event not found")

type ExpenseService struct {
	store store.Store
}

func NewExpenseService(s store.Store) *ExpenseService {
	return &ExpenseService{store: s}
}

// GetGroupCurrency returns the base currency for the given group.
func (es *ExpenseService) GetGroupCurrency(ctx context.Context, groupID string) (string, error) {
	data, _, err := es.store.ReadObject(ctx, groupKey(groupID))
	if err != nil {
		return "", err
	}
	var g domain.Group
	if err := json.Unmarshal(data, &g); err != nil {
		return "", err
	}
	return g.Currency, nil
}

// AddExpense appends a new expense event to the group. originalExpense is
// non-nil only when the input currency differs from the group's base currency.
func (es *ExpenseService) AddExpense(ctx context.Context, groupID, createdBy, description, paidBy string, amount float64, splits map[string]float64, originalExpense *domain.OriginalExpense) (string, error) {
	eventID := uuid.New().String()
	return eventID, withRetry(ctx, es.store, groupKey(groupID), func(data []byte) ([]byte, error) {
		if data == nil {
			return nil, store.ErrNotFound
		}
		var g domain.Group
		if err := json.Unmarshal(data, &g); err != nil {
			return nil, err
		}
		memberSet := make(map[string]struct{}, len(g.Members))
		for _, m := range g.Members {
			memberSet[m] = struct{}{}
		}
		if _, ok := memberSet[createdBy]; !ok {
			return nil, fmt.Errorf("user is not a member of this group")
		}
		if err := domain.ValidateSplits(amount, splits, g.Members); err != nil {
			return nil, err
		}
		g.Events = append(g.Events, domain.Event{
			ID:              eventID,
			Type:            domain.EventTypeExpense,
			CreatedAt:       time.Now().UTC(),
			CreatedBy:       createdBy,
			Description:     description,
			Amount:          amount,
			PaidBy:          paidBy,
			Splits:          splits,
			OriginalExpense: originalExpense,
		})
		return json.Marshal(g)
	})
}

func (es *ExpenseService) CancelExpense(ctx context.Context, groupID, cancelledBy, eventID string) error {
	reversalID := uuid.New().String()
	return withRetry(ctx, es.store, groupKey(groupID), func(data []byte) ([]byte, error) {
		if data == nil {
			return nil, store.ErrNotFound
		}
		var g domain.Group
		if err := json.Unmarshal(data, &g); err != nil {
			return nil, err
		}
		found := false
		for _, e := range g.Events {
			if e.ID == eventID && e.Type == domain.EventTypeExpense {
				found = true
			}
			if e.Type == domain.EventTypeReversal && e.ReversedEventID == eventID {
				return nil, fmt.Errorf("expense already cancelled")
			}
		}
		if !found {
			return nil, ErrEventNotFound
		}
		g.Events = append(g.Events, domain.Event{
			ID:              reversalID,
			Type:            domain.EventTypeReversal,
			CreatedAt:       time.Now().UTC(),
			CreatedBy:       cancelledBy,
			ReversedEventID: eventID,
		})
		return json.Marshal(g)
	})
}

// ListEvents returns events newest-first with pagination.
func (es *ExpenseService) ListEvents(ctx context.Context, groupID string, limit, offset int) ([]domain.Event, int, error) {
	data, _, err := es.store.ReadObject(ctx, groupKey(groupID))
	if err != nil {
		return nil, 0, err
	}
	var g domain.Group
	if err := json.Unmarshal(data, &g); err != nil {
		return nil, 0, err
	}

	reversed := make([]domain.Event, len(g.Events))
	for i, e := range g.Events {
		reversed[len(g.Events)-1-i] = e
	}

	total := len(reversed)
	if offset >= total {
		return []domain.Event{}, total, nil
	}
	end := offset + limit
	if end > total {
		end = total
	}
	return reversed[offset:end], total, nil
}
