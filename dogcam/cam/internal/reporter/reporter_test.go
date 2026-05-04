package reporter_test

import (
	"context"
	"testing"

	"github.com/fergalhk-lab/apps/dogcam/cam/internal/reporter"
	"github.com/fergalhk-lab/apps/dogcam/gen/dogcampb"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type fakeClient struct {
	calls []*dogcampb.MetricsPayload
}

func (f *fakeClient) ReportMetrics(_ context.Context, p *dogcampb.MetricsPayload) error {
	f.calls = append(f.calls, p)
	return nil
}

func TestReporter_ReportsOnFirstCall(t *testing.T) {
	c := &fakeClient{}
	r := reporter.New(c)
	p := &dogcampb.MetricsPayload{FramesCaptured: 1}
	require.NoError(t, r.Report(context.Background(), p))
	assert.Len(t, c.calls, 1)
}

func TestReporter_SkipsWhenUnchanged(t *testing.T) {
	c := &fakeClient{}
	r := reporter.New(c)
	p := &dogcampb.MetricsPayload{FramesCaptured: 1}
	require.NoError(t, r.Report(context.Background(), p))
	require.NoError(t, r.Report(context.Background(), p))
	assert.Len(t, c.calls, 1) // only one call
}

func TestReporter_ReportsWhenChanged(t *testing.T) {
	c := &fakeClient{}
	r := reporter.New(c)
	require.NoError(t, r.Report(context.Background(), &dogcampb.MetricsPayload{FramesCaptured: 1}))
	require.NoError(t, r.Report(context.Background(), &dogcampb.MetricsPayload{FramesCaptured: 2}))
	assert.Len(t, c.calls, 2)
}

func TestReporter_ReportsStreamingStateChange(t *testing.T) {
	c := &fakeClient{}
	r := reporter.New(c)
	require.NoError(t, r.Report(context.Background(), &dogcampb.MetricsPayload{StreamingActive: false}))
	require.NoError(t, r.Report(context.Background(), &dogcampb.MetricsPayload{StreamingActive: true}))
	assert.Len(t, c.calls, 2)
}
