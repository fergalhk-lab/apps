// billsplit/internal/domain/types.go
package domain

import "time"

type User struct {
	Username     string   `json:"username"`
	PasswordHash string   `json:"passwordHash"`
	GroupIDs     []string `json:"groupIds"`
	IsAdmin      bool     `json:"isAdmin"`
}

type Invite struct {
	Code    string `json:"code"`
	Used    bool   `json:"used"`
	IsAdmin bool   `json:"isAdmin"`
}

// UsersData is the root object stored at users.json.
type UsersData struct {
	Users   []User   `json:"users"`
	Invites []Invite `json:"invites"`
}

// UserSummary is the public representation of a user returned by the API.
type UserSummary struct {
	ID      string `json:"id"`
	IsAdmin bool   `json:"isAdmin"`
}

type EventType string

const (
	EventTypeExpense    EventType = "expense"
	EventTypeReversal   EventType = "reversal"
	EventTypeSettlement EventType = "settlement"
)

// OriginalExpense records the pre-conversion currency and amount in cents when
// an expense was entered in a currency different from the group's base currency.
type OriginalExpense struct {
	Currency string `json:"currency"`
	Amount   int64  `json:"amount"` // in cents
}

type Event struct {
	ID        string    `json:"id"`
	Type      EventType `json:"type"`
	CreatedAt time.Time `json:"createdAt"`
	CreatedBy string    `json:"createdBy"`

	// expense fields
	Description     string           `json:"description,omitempty"`
	Amount          int64            `json:"amount,omitempty"` // in cents
	PaidBy          string           `json:"paidBy,omitempty"`
	Splits          map[string]int64 `json:"splits,omitempty"` // in cents
	OriginalExpense *OriginalExpense  `json:"originalExpense,omitempty"`

	// reversal fields
	ReversedEventID string `json:"reversedEventId,omitempty"`

	// settlement fields
	From string `json:"from,omitempty"`
	To   string `json:"to,omitempty"`
}

// Group is stored at groups/{id}.json.
type Group struct {
	Name     string   `json:"name"`
	Members  []string `json:"members"`
	Currency string   `json:"currency"`
	Events   []Event  `json:"events"`
}

// Settlement represents a single payment that would resolve outstanding debts.
type Settlement struct {
	From   string `json:"from"`
	To     string `json:"to"`
	Amount int64  `json:"amount"` // in cents
}
