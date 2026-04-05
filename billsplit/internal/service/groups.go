// billsplit/internal/service/groups.go
package service

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"math"

	"github.com/fergalhk-lab/apps/billsplit/internal/domain"
	"github.com/fergalhk-lab/apps/billsplit/internal/store"
	"github.com/google/uuid"
	"go.uber.org/zap"
)

var ErrUnknownMembers = errors.New("one or more members not found")
var ErrDuplicateMembers = errors.New("duplicate members")
var ErrOutstandingBalances = errors.New("group has outstanding balances")

type GroupSummary struct {
	ID         string  `json:"id"`
	Name       string  `json:"name"`
	Currency   string  `json:"currency"`
	NetBalance float64 `json:"netBalance"`
}

type GroupDetail struct {
	ID          string              `json:"id"`
	Name        string              `json:"name"`
	Members     []string            `json:"members"`
	Currency    string              `json:"currency"`
	Balances    map[string]float64  `json:"balances"`
	Settlements []domain.Settlement `json:"settlements"`
}

type GroupService struct {
	store  store.Store
	logger *zap.Logger
}

func NewGroupService(s store.Store, logger *zap.Logger) *GroupService {
	return &GroupService{store: s, logger: logger.Named("service.groups")}
}

func (gs *GroupService) CreateGroup(ctx context.Context, creatorUsername, name, currency string, otherMembers []string) (string, error) {
	allMembers := append([]string{creatorUsername}, otherMembers...)

	if err := validateNoDuplicates(allMembers); err != nil {
		return "", err
	}

	// Validate all members exist
	if err := gs.validateMembersExist(ctx, allMembers); err != nil {
		return "", err
	}

	groupID := uuid.New().String()
	group := domain.Group{
		Name:     name,
		Members:  allMembers,
		Currency: currency,
		Events:   []domain.Event{},
	}

	// Write group object first, then update users.json.
	// Note: these two writes are not atomic. If the users.json update fails after
	// max retries, the group object is orphaned (exists in S3 but no user has it
	// in their GroupIDs). This is safe to ignore in practice: withRetry retries
	// once on concurrent-write conflicts, and users.json contention is very low.
	data, err := json.Marshal(group)
	if err != nil {
		return "", err
	}
	if err := gs.store.WriteObject(ctx, groupKey(groupID), data, ""); err != nil {
		return "", fmt.Errorf("create group: %w", err)
	}

	// Add groupID to all members in users.json
	if err := withRetry(ctx, gs.store, usersKey, gs.logger, func(raw []byte) ([]byte, error) {
		if raw == nil {
			return nil, errors.New("users.json not found")
		}
		var ud domain.UsersData
		if err := json.Unmarshal(raw, &ud); err != nil {
			return nil, err
		}
		for i := range ud.Users {
			for _, m := range allMembers {
				if ud.Users[i].Username == m {
					ud.Users[i].GroupIDs = append(ud.Users[i].GroupIDs, groupID)
				}
			}
		}
		return json.Marshal(ud)
	}); err != nil {
		return "", fmt.Errorf("update users: %w", err)
	}

	return groupID, nil
}

func (gs *GroupService) ListGroups(ctx context.Context, username string) ([]GroupSummary, error) {
	data, _, err := gs.store.ReadObject(ctx, usersKey)
	if errors.Is(err, store.ErrNotFound) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	var ud domain.UsersData
	if err := json.Unmarshal(data, &ud); err != nil {
		return nil, err
	}

	var groupIDs []string
	for _, u := range ud.Users {
		if u.Username == username {
			groupIDs = u.GroupIDs
			break
		}
	}

	summaries := make([]GroupSummary, 0, len(groupIDs))
	for _, id := range groupIDs {
		g, err := gs.readGroup(ctx, id)
		if err != nil {
			gs.logger.Warn("group not found, skipping", zap.String("group_id", id))
			continue
		}
		balances := domain.ComputeBalances(g)
		summaries = append(summaries, GroupSummary{
			ID:         id,
			Name:       g.Name,
			Currency:   g.Currency,
			NetBalance: balances[username],
		})
	}
	return summaries, nil
}

func (gs *GroupService) GetGroup(ctx context.Context, groupID string) (*GroupDetail, error) {
	g, err := gs.readGroup(ctx, groupID)
	if err != nil {
		return nil, err
	}
	balances := domain.ComputeBalances(g)
	return &GroupDetail{
		ID:          groupID,
		Name:        g.Name,
		Members:     g.Members,
		Currency:    g.Currency,
		Balances:    balances,
		Settlements: domain.ComputeSettlements(balances),
	}, nil
}

func (gs *GroupService) LeaveGroup(ctx context.Context, groupID, username string) error {
	// Check balance is zero
	g, err := gs.readGroup(ctx, groupID)
	if err != nil {
		return err
	}
	balances := domain.ComputeBalances(g)
	if b := balances[username]; b != 0 {
		return fmt.Errorf("cannot leave: outstanding balance of %.2f", b)
	}

	// Remove groupID from user's record
	return withRetry(ctx, gs.store, usersKey, gs.logger, func(raw []byte) ([]byte, error) {
		var ud domain.UsersData
		if err := json.Unmarshal(raw, &ud); err != nil {
			return nil, err
		}
		for i := range ud.Users {
			if ud.Users[i].Username == username {
				ud.Users[i].GroupIDs = removeString(ud.Users[i].GroupIDs, groupID)
			}
		}
		return json.Marshal(ud)
	})
}

func (gs *GroupService) DeleteGroup(ctx context.Context, groupID string) error {
	g, err := gs.readGroup(ctx, groupID)
	if err != nil {
		return err
	}
	for _, b := range domain.ComputeBalances(g) {
		if math.Abs(b) > 1e-9 {
			return ErrOutstandingBalances
		}
	}

	// Soft delete: remove the group ID from all members in users.json but leave
	// the group object in S3. This preserves the transaction history if needed.
	return withRetry(ctx, gs.store, usersKey, gs.logger, func(raw []byte) ([]byte, error) {
		var ud domain.UsersData
		if err := json.Unmarshal(raw, &ud); err != nil {
			return nil, err
		}
		for i := range ud.Users {
			ud.Users[i].GroupIDs = removeString(ud.Users[i].GroupIDs, groupID)
		}
		return json.Marshal(ud)
	})
}

func (gs *GroupService) ReadGroup(ctx context.Context, groupID string) (domain.Group, error) {
	return gs.readGroup(ctx, groupID)
}

func (gs *GroupService) readGroup(ctx context.Context, groupID string) (domain.Group, error) {
	data, _, err := gs.store.ReadObject(ctx, groupKey(groupID))
	if err != nil {
		return domain.Group{}, err
	}
	var g domain.Group
	return g, json.Unmarshal(data, &g)
}

func (gs *GroupService) validateMembersExist(ctx context.Context, members []string) error {
	data, _, err := gs.store.ReadObject(ctx, usersKey)
	if errors.Is(err, store.ErrNotFound) {
		return ErrUnknownMembers
	}
	if err != nil {
		return err
	}
	var ud domain.UsersData
	if err := json.Unmarshal(data, &ud); err != nil {
		return err
	}
	known := make(map[string]struct{}, len(ud.Users))
	for _, u := range ud.Users {
		known[u.Username] = struct{}{}
	}
	var missing []string
	for _, m := range members {
		if _, ok := known[m]; !ok {
			missing = append(missing, m)
		}
	}
	if len(missing) > 0 {
		return fmt.Errorf("%w: %v", ErrUnknownMembers, missing)
	}
	return nil
}

func validateNoDuplicates(members []string) error {
	seen := make(map[string]struct{}, len(members))
	for _, m := range members {
		if _, ok := seen[m]; ok {
			return fmt.Errorf("%w: %s", ErrDuplicateMembers, m)
		}
		seen[m] = struct{}{}
	}
	return nil
}
