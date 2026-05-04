package client_test

import (
	"context"
	"errors"
	"net"
	"sync/atomic"
	"testing"
	"time"

	"github.com/fergalhk-lab/apps/dogcam/cam/internal/client"
	"github.com/fergalhk-lab/apps/dogcam/cam/internal/reporter"
	"github.com/fergalhk-lab/apps/dogcam/gen/dogcampb"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/test/bufconn"
)

const bufSize = 1024 * 1024

type fakeReporterClient struct{}

func (f *fakeReporterClient) ReportMetrics(_ context.Context, _ *dogcampb.MetricsPayload) error {
	return nil
}

// fakeStreamServer sends START then receives frames until context done.
type fakeStreamServer struct {
	dogcampb.UnimplementedStreamServiceServer
	dogcampb.UnimplementedTelemetryServiceServer
	received atomic.Int64
}

func (s *fakeStreamServer) VideoStream(stream dogcampb.StreamService_VideoStreamServer) error {
	err := stream.Send(&dogcampb.ControlMessage{
		Command:         dogcampb.ControlMessage_START,
		FrameIntervalMs: 50,
	})
	if err != nil {
		return err
	}
	for {
		_, err := stream.Recv()
		if err != nil {
			return err
		}
		s.received.Add(1)
	}
}

func newTestServer(t *testing.T) (*client.Client, *fakeStreamServer) {
	t.Helper()
	lis := bufconn.Listen(bufSize)
	fake := &fakeStreamServer{}
	srv := grpc.NewServer()
	dogcampb.RegisterStreamServiceServer(srv, fake)
	dogcampb.RegisterTelemetryServiceServer(srv, fake)
	go func() { _ = srv.Serve(lis) }()
	t.Cleanup(func() { srv.Stop() })

	conn, err := grpc.NewClient("passthrough://bufnet",
		grpc.WithContextDialer(func(ctx context.Context, _ string) (net.Conn, error) {
			return lis.DialContext(ctx)
		}),
		grpc.WithTransportCredentials(insecure.NewCredentials()),
	)
	require.NoError(t, err)
	t.Cleanup(func() { conn.Close() })

	r := reporter.New(&fakeReporterClient{})
	c := client.New(conn, "test-key", r)
	return c, fake
}

type fakeCapturer struct{ data []byte }

func (f *fakeCapturer) Capture() ([]byte, error) { return f.data, nil }

func TestClient_SendsFramesOnStart(t *testing.T) {
	c, fake := newTestServer(t)
	capturer := &fakeCapturer{data: []byte{0xFF, 0xD8}}

	ctx, cancel := context.WithTimeout(context.Background(), 500*time.Millisecond)
	defer cancel()

	err := c.RunOnce(ctx, capturer)
	// Context cancellation or deadline is expected
	assert.True(t, errors.Is(err, context.DeadlineExceeded) || err != nil)
	assert.Greater(t, fake.received.Load(), int64(0))
}
