// billsplit/internal/service/invites.go
package service

import (
	"context"
	"encoding/json"

	"github.com/fergalhk-lab/apps/billsplit/internal/domain"
	"github.com/fergalhk-lab/apps/billsplit/internal/store"
	"github.com/google/uuid"
)

type InviteService struct {
	store store.Store
}

func NewInviteService(s store.Store) *InviteService {
	return &InviteService{store: s}
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
