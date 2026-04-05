// billsplit/internal/service/invites.go
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

type InviteService struct {
	store  store.Store
	logger *zap.Logger
}

func NewInviteService(s store.Store, logger *zap.Logger) *InviteService {
	return &InviteService{store: s, logger: logger.Named("service.invites")}
}

func (is *InviteService) GenerateInvite(ctx context.Context, isAdmin bool) (string, error) {
	code := uuid.New().String()[:8]
	err := withRetry(ctx, is.store, usersKey, func(data []byte) ([]byte, error) {
		var ud domain.UsersData
		if data != nil {
			if err := json.Unmarshal(data, &ud); err != nil {
				return nil, err
			}
		}
		ud.Invites = append(ud.Invites, domain.Invite{Code: code, IsAdmin: isAdmin})
		return json.Marshal(ud)
	})
	if err != nil {
		return "", err
	}
	return code, nil
}

func (is *InviteService) HasInvites(ctx context.Context) (bool, error) {
	data, _, err := is.store.ReadObject(ctx, usersKey)
	if errors.Is(err, store.ErrNotFound) {
		return false, nil
	}
	if err != nil {
		return false, err
	}
	var ud domain.UsersData
	if err := json.Unmarshal(data, &ud); err != nil {
		return false, fmt.Errorf("corrupt users data: %w", err)
	}
	return len(ud.Invites) > 0, nil
}
