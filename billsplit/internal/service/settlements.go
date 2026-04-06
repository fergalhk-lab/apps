// billsplit/internal/service/settlements.go
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
	"go.uber.org/zap"
)

type SettlementService struct {
	store  store.Store
	logger *zap.Logger
}

func NewSettlementService(s store.Store, logger *zap.Logger) *SettlementService {
	return &SettlementService{store: s, logger: logger.Named("service.settlements")}
}

func (ss *SettlementService) AddSettlement(ctx context.Context, groupID, createdBy, from, to string, amount int64) error {
	eventID := uuid.New().String()
	return withRetry(ctx, ss.store, groupKey(groupID), ss.logger, func(data []byte) ([]byte, error) {
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
		_, fromOk := memberSet[from]
		_, toOk := memberSet[to]
		if !fromOk || !toOk {
			return nil, fmt.Errorf("from and to must be group members")
		}
		if from == to {
			return nil, errors.New("from and to must be different members")
		}
		g.Events = append(g.Events, domain.Event{
			ID:        eventID,
			Type:      domain.EventTypeSettlement,
			CreatedAt: time.Now().UTC(),
			CreatedBy: createdBy,
			From:      from,
			To:        to,
			Amount:    amount,
		})
		return json.Marshal(g)
	})
}
