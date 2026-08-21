package reporter

import (
	"context"
	"sync"

	"github.com/fergalhk-lab/apps/dogcam/gen/dogcampb"
	"google.golang.org/protobuf/proto"
)

// Client is the subset of the gRPC TelemetryServiceClient used by Reporter.
type Client interface {
	ReportMetrics(ctx context.Context, p *dogcampb.MetricsPayload) error
}

type Reporter struct {
	mu   sync.Mutex
	last *dogcampb.MetricsPayload
	c    Client
}

func New(c Client) *Reporter {
	return &Reporter{c: c}
}

// Report calls ReportMetrics only if the payload differs from the last reported one.
func (r *Reporter) Report(ctx context.Context, p *dogcampb.MetricsPayload) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	if r.last != nil && proto.Equal(r.last, p) {
		return nil
	}
	if err := r.c.ReportMetrics(ctx, p); err != nil {
		return err
	}
	r.last = proto.Clone(p).(*dogcampb.MetricsPayload)
	return nil
}
