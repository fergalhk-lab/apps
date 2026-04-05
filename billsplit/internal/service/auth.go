// billsplit/internal/service/auth.go
package service

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"github.com/fergalhk-lab/apps/billsplit/internal/domain"
	"github.com/fergalhk-lab/apps/billsplit/internal/store"
	"github.com/golang-jwt/jwt/v5"
	"golang.org/x/crypto/bcrypt"
)

const usersKey = "users.json"

var (
	ErrInvalidInvite  = errors.New("invalid or already-used invite code")
	ErrUsernameTaken  = errors.New("username already taken")
	ErrBadCredentials = errors.New("invalid username or password")
)

type Claims struct {
	Username string
	IsAdmin  bool
}

type AuthService struct {
	store     store.Store
	jwtSecret string
}

func NewAuthService(s store.Store, jwtSecret string) *AuthService {
	return &AuthService{store: s, jwtSecret: jwtSecret}
}

func (a *AuthService) Register(ctx context.Context, username, password, inviteCode string) error {
	return withRetry(ctx, a.store, usersKey, func(data []byte) ([]byte, error) {
		ud, err := a.unmarshalOrEmpty(data)
		if err != nil {
			return nil, err
		}

		if !a.hasValidInvite(ud, inviteCode) {
			return nil, ErrInvalidInvite
		}
		for _, u := range ud.Users {
			if u.Username == username {
				return nil, ErrUsernameTaken
			}
		}

		hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
		if err != nil {
			return nil, fmt.Errorf("hash password: %w", err)
		}

		isAdmin := a.inviteIsAdmin(ud, inviteCode)
		for i := range ud.Invites {
			if ud.Invites[i].Code == inviteCode {
				ud.Invites[i].Used = true
			}
		}
		ud.Users = append(ud.Users, domain.User{
			Username:     username,
			PasswordHash: string(hash),
			GroupIDs:     []string{},
			IsAdmin:      isAdmin,
		})
		return json.Marshal(ud)
	})
}

func (a *AuthService) Login(ctx context.Context, username, password string) (string, Claims, error) {
	data, _, err := a.store.ReadObject(ctx, usersKey)
	if errors.Is(err, store.ErrNotFound) {
		return "", Claims{}, ErrBadCredentials
	}
	if err != nil {
		return "", Claims{}, err
	}
	var ud domain.UsersData
	if err := json.Unmarshal(data, &ud); err != nil {
		return "", Claims{}, err
	}
	for _, u := range ud.Users {
		if u.Username == username {
			if bcrypt.CompareHashAndPassword([]byte(u.PasswordHash), []byte(password)) != nil {
				return "", Claims{}, ErrBadCredentials
			}
			c := Claims{Username: u.Username, IsAdmin: u.IsAdmin}
			token, err := a.issueToken(c)
			if err != nil {
				return "", Claims{}, err
			}
			return token, c, nil
		}
	}
	return "", Claims{}, ErrBadCredentials
}

// ListUsers returns a summary of every registered user.
func (a *AuthService) ListUsers(ctx context.Context) ([]domain.UserSummary, error) {
	data, _, err := a.store.ReadObject(ctx, usersKey)
	if errors.Is(err, store.ErrNotFound) {
		return []domain.UserSummary{}, nil
	}
	if err != nil {
		return nil, err
	}
	var ud domain.UsersData
	if err := json.Unmarshal(data, &ud); err != nil {
		return nil, fmt.Errorf("corrupt users data: %w", err)
	}
	result := make([]domain.UserSummary, len(ud.Users))
	for i, u := range ud.Users {
		result[i] = domain.UserSummary{ID: u.Username, IsAdmin: u.IsAdmin}
	}
	return result, nil
}

func (a *AuthService) VerifyToken(tokenStr string) (Claims, error) {
	token, err := jwt.Parse(tokenStr, func(t *jwt.Token) (interface{}, error) {
		if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method")
		}
		return []byte(a.jwtSecret), nil
	})
	if err != nil || !token.Valid {
		return Claims{}, errors.New("invalid token")
	}
	mc, ok := token.Claims.(jwt.MapClaims)
	if !ok {
		return Claims{}, errors.New("invalid claims")
	}
	username, ok := mc["username"].(string)
	if !ok || username == "" {
		return Claims{}, errors.New("invalid token claims")
	}
	isAdmin, _ := mc["isAdmin"].(bool)
	return Claims{Username: username, IsAdmin: isAdmin}, nil
}

// Store returns the underlying store for use by other services.
func (a *AuthService) Store() store.Store { return a.store }

func (a *AuthService) issueToken(c Claims) (string, error) {
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"username": c.Username,
		"isAdmin":  c.IsAdmin,
		"exp":      time.Now().Add(24 * time.Hour).Unix(),
	})
	return token.SignedString([]byte(a.jwtSecret))
}

func (a *AuthService) unmarshalOrEmpty(data []byte) (domain.UsersData, error) {
	if data == nil {
		return domain.UsersData{}, nil
	}
	var ud domain.UsersData
	if err := json.Unmarshal(data, &ud); err != nil {
		return domain.UsersData{}, fmt.Errorf("corrupt users data: %w", err)
	}
	return ud, nil
}

func (a *AuthService) hasValidInvite(ud domain.UsersData, code string) bool {
	for _, inv := range ud.Invites {
		if inv.Code == code && !inv.Used {
			return true
		}
	}
	return false
}

func (a *AuthService) inviteIsAdmin(ud domain.UsersData, code string) bool {
	for _, inv := range ud.Invites {
		if inv.Code == code {
			return inv.IsAdmin
		}
	}
	return false
}
