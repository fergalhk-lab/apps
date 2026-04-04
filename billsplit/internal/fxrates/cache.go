package fxrates

import (
	"context"
	"encoding/json"
	"sync"
	"time"

	"github.com/fergalhk-lab/apps/billsplit/internal/store"
	"go.uber.org/zap"
)

const cacheTTL = time.Hour

// Cache lazily loads Rates from S3 and holds them in memory for cacheTTL (1 hour).
// Thread-safe; safe to share across goroutines.
type Cache struct {
	store     store.Store
	logger    *zap.Logger
	mu        sync.Mutex
	data      *Rates
	fetchedAt time.Time
}

// NewCache returns a Cache backed by s.
func NewCache(s store.Store, logger *zap.Logger) *Cache {
	return &Cache{store: s, logger: logger.Named("fxrates")}
}

// Get returns the cached Rates if fetched within the last hour, otherwise
// fetches from S3, caches the result, and returns it.
func (c *Cache) Get(ctx context.Context) (*Rates, error) {
	c.mu.Lock()
	defer c.mu.Unlock()
	if c.data != nil && time.Since(c.fetchedAt) < cacheTTL {
		return c.data, nil
	}
	raw, _, err := c.store.ReadObject(ctx, S3Key)
	if err != nil {
		return nil, err
	}
	var r Rates
	if err := json.Unmarshal(raw, &r); err != nil {
		return nil, err
	}
	c.data = &r
	c.fetchedAt = time.Now()
	return c.data, nil
}
