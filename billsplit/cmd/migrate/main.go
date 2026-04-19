// billsplit/cmd/migrate/main.go
package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"math"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/fergalhk-lab/apps/billsplit/internal/config"
	"github.com/fergalhk-lab/apps/billsplit/internal/dependencies"
	"github.com/fergalhk-lab/apps/billsplit/internal/domain"
	localstore "github.com/fergalhk-lab/apps/billsplit/internal/store"
)

// oldOriginalExpense is the pre-migration schema with float64 amount.
type oldOriginalExpense struct {
	Currency string  `json:"currency"`
	Amount   float64 `json:"amount"`
}

// oldEvent is the pre-migration schema with float64 monetary fields.
type oldEvent struct {
	ID              string             `json:"id"`
	Type            domain.EventType   `json:"type"`
	CreatedAt       time.Time          `json:"createdAt"`
	CreatedBy       string             `json:"createdBy"`
	Description     string             `json:"description,omitempty"`
	Amount          float64            `json:"amount,omitempty"`
	PaidBy          string             `json:"paidBy,omitempty"`
	Splits          map[string]float64 `json:"splits,omitempty"`
	OriginalExpense *oldOriginalExpense `json:"originalExpense,omitempty"`
	ReversedEventID string             `json:"reversedEventId,omitempty"`
	From            string             `json:"from,omitempty"`
	To              string             `json:"to,omitempty"`
}

// oldGroup is the pre-migration schema.
type oldGroup struct {
	Name     string     `json:"name"`
	Members  []string   `json:"members"`
	Currency string     `json:"currency"`
	Events   []oldEvent `json:"events"`
}

func toCents(f float64) int64 {
	return int64(math.Round(f * 100))
}

func convertGroup(old oldGroup) domain.Group {
	events := make([]domain.Event, len(old.Events))
	for i, e := range old.Events {
		var splits map[string]int64
		if len(e.Splits) > 0 {
			splits = make(map[string]int64, len(e.Splits))
			for k, v := range e.Splits {
				splits[k] = toCents(v)
			}
		}
		var origExp *domain.OriginalExpense
		if e.OriginalExpense != nil {
			origExp = &domain.OriginalExpense{
				Currency: e.OriginalExpense.Currency,
				Amount:   toCents(e.OriginalExpense.Amount),
			}
		}
		events[i] = domain.Event{
			ID:              e.ID,
			Type:            e.Type,
			CreatedAt:       e.CreatedAt,
			CreatedBy:       e.CreatedBy,
			Description:     e.Description,
			Amount:          toCents(e.Amount),
			PaidBy:          e.PaidBy,
			Splits:          splits,
			OriginalExpense: origExp,
			ReversedEventID: e.ReversedEventID,
			From:            e.From,
			To:              e.To,
		}
	}
	return domain.Group{
		Name:     old.Name,
		Members:  old.Members,
		Currency: old.Currency,
		Events:   events,
	}
}

func migrateGroup(ctx context.Context, st localstore.Store, key string) error {
	data, etag, err := st.ReadObject(ctx, key)
	if err != nil {
		return fmt.Errorf("read: %w", err)
	}

	var old oldGroup
	if err := json.Unmarshal(data, &old); err != nil {
		return fmt.Errorf("unmarshal: %w", err)
	}

	// Back up original before modifying
	backupKey := "migration-backup/" + key
	if err := st.ForceWriteObject(ctx, backupKey, data); err != nil {
		return fmt.Errorf("backup: %w", err)
	}

	newGroup := convertGroup(old)
	newData, err := json.Marshal(newGroup)
	if err != nil {
		return fmt.Errorf("marshal: %w", err)
	}

	// Conditional write: fails if the object was modified since we read it
	if err := st.WriteObject(ctx, key, newData, etag); err != nil {
		return fmt.Errorf("write: %w", err)
	}
	return nil
}

func main() {
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("config: %v", err)
	}

	s3Client, err := dependencies.NewS3Client(context.Background())
	if err != nil {
		log.Fatalf("s3 client: %v", err)
	}

	st := localstore.NewS3Store(s3Client, cfg.S3Bucket)
	ctx := context.Background()

	// List all group objects
	var keys []string
	paginator := s3.NewListObjectsV2Paginator(s3Client, &s3.ListObjectsV2Input{
		Bucket: aws.String(cfg.S3Bucket),
		Prefix: aws.String("groups/"),
	})
	for paginator.HasMorePages() {
		page, err := paginator.NextPage(ctx)
		if err != nil {
			log.Fatalf("list objects: %v", err)
		}
		for _, obj := range page.Contents {
			if obj.Key != nil {
				keys = append(keys, *obj.Key)
			}
		}
	}

	log.Printf("found %d group objects", len(keys))

	ok, failed := 0, 0
	for _, key := range keys {
		if err := migrateGroup(ctx, st, key); err != nil {
			log.Printf("ERROR %s: %v", key, err)
			failed++
		} else {
			log.Printf("migrated %s", key)
			ok++
		}
	}

	log.Printf("done: %d migrated, %d failed", ok, failed)
	if failed > 0 {
		log.Fatal("migration completed with errors — check logs and restore from migration-backup/ if needed")
	}
}
