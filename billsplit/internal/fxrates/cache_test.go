package fxrates_test

import (
	"context"
	"encoding/json"
	"testing"
	"time"

	"github.com/fergalhk-lab/apps/billsplit/internal/fxrates"
	"github.com/fergalhk-lab/apps/billsplit/internal/store"
	"github.com/stretchr/testify/require"
)

// fakeStore is a minimal in-memory store.Store for testing — no MinIO needed.
type fakeStore struct {
	objects map[string][]byte
	calls   int
}

func (f *fakeStore) ReadObject(_ context.Context, key string) ([]byte, string, error) {
	f.calls++
	d, ok := f.objects[key]
	if !ok {
		return nil, "", store.ErrNotFound
	}
	return d, "etag", nil
}

func (f *fakeStore) WriteObject(_ context.Context, _ string, _ []byte, _ string) error { return nil }
func (f *fakeStore) ForceWriteObject(_ context.Context, _ string, _ []byte) error      { return nil }

func marshalRates(t *testing.T, r fxrates.Rates) []byte {
	t.Helper()
	b, err := json.Marshal(r)
	require.NoError(t, err)
	return b
}

func TestCache_FetchesOnFirstCall(t *testing.T) {
	fs := &fakeStore{objects: map[string][]byte{
		fxrates.S3Key: marshalRates(t, fxrates.Rates{
			Base:              "USD",
			Rates:             map[string]float64{"USD": 1.0, "EUR": 0.9},
			ProviderUpdatedAt: time.Now(),
		}),
	}}

	c := fxrates.NewCache(fs)
	rates, err := c.Get(context.Background())
	require.NoError(t, err)
	require.Equal(t, "USD", rates.Base)
	require.Len(t, rates.Rates, 2)
	require.Equal(t, 1, fs.calls)
}

func TestCache_ReturnsCachedOnSubsequentCall(t *testing.T) {
	fs := &fakeStore{objects: map[string][]byte{
		fxrates.S3Key: marshalRates(t, fxrates.Rates{
			Base:  "USD",
			Rates: map[string]float64{"USD": 1.0},
		}),
	}}

	c := fxrates.NewCache(fs)
	first, _ := c.Get(context.Background())
	second, err := c.Get(context.Background())
	require.NoError(t, err)
	require.Same(t, first, second, "expected same pointer on cache hit")
	require.Equal(t, 1, fs.calls, "second call should use cache, not hit store again")
}

func TestCache_ErrorWhenRatesAbsent(t *testing.T) {
	fs := &fakeStore{objects: map[string][]byte{}} // no fxrates key stored

	c := fxrates.NewCache(fs)
	_, err := c.Get(context.Background())
	require.Error(t, err)
}
