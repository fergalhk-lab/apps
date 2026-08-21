package metricsstore_test

import (
	"testing"

	"github.com/fergalhk-lab/apps/dogcam/gen/dogcampb"
	"github.com/fergalhk-lab/apps/dogcam/server/internal/metricsstore"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestStore_StartsNil(t *testing.T) {
	s := metricsstore.New()
	assert.Nil(t, s.Get())
}

func TestStore_UpdateAndGet(t *testing.T) {
	s := metricsstore.New()
	payload := &dogcampb.MetricsPayload{
		FramesCaptured:  10,
		FramesSent:      9,
		StreamingActive: true,
		UptimeSeconds:   60,
	}
	s.Update(payload)
	got := s.Get()
	require.NotNil(t, got)
	assert.Equal(t, int64(10), got.FramesCaptured)
	assert.Equal(t, int64(9), got.FramesSent)
	assert.True(t, got.StreamingActive)
}

func TestStore_UpdateOverwrites(t *testing.T) {
	s := metricsstore.New()
	s.Update(&dogcampb.MetricsPayload{FramesCaptured: 1})
	s.Update(&dogcampb.MetricsPayload{FramesCaptured: 2})
	assert.Equal(t, int64(2), s.Get().FramesCaptured)
}
