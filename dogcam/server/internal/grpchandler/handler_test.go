package grpchandler_test

import (
	"context"
	"net"
	"testing"
	"time"

	"github.com/fergalhk-lab/apps/dogcam/gen/dogcampb"
	"github.com/fergalhk-lab/apps/dogcam/server/internal/broadcast"
	"github.com/fergalhk-lab/apps/dogcam/server/internal/grpchandler"
	"github.com/fergalhk-lab/apps/dogcam/server/internal/metricsstore"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
	"google.golang.org/grpc/test/bufconn"
)

const bufSize = 1024 * 1024

func newTestServer(t *testing.T, apiKey string) (dogcampb.StreamServiceClient, dogcampb.TelemetryServiceClient, *broadcast.Broadcaster, *metricsstore.Store) {
	t.Helper()
	lis := bufconn.Listen(bufSize)
	b := broadcast.New(2000)
	ms := metricsstore.New()

	srv := grpc.NewServer()
	dogcampb.RegisterStreamServiceServer(srv, grpchandler.NewStreamHandler(b, apiKey))
	dogcampb.RegisterTelemetryServiceServer(srv, grpchandler.NewTelemetryHandler(ms, apiKey))

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

	return dogcampb.NewStreamServiceClient(conn), dogcampb.NewTelemetryServiceClient(conn), b, ms
}

func authCtx(apiKey string) context.Context {
	md := metadata.Pairs("authorization", "Bearer "+apiKey)
	return metadata.NewOutgoingContext(context.Background(), md)
}

func TestVideoStream_RejectsInvalidAPIKey(t *testing.T) {
	streamClient, _, _, _ := newTestServer(t, "secret")
	stream, err := streamClient.VideoStream(authCtx("wrong"))
	require.NoError(t, err) // stream handshake succeeds; auth rejection arrives async
	// Recv delivers the server's Unauthenticated status error
	_, recvErr := stream.Recv()
	require.Error(t, recvErr)
	assert.Equal(t, codes.Unauthenticated, status.Code(recvErr))
}

func TestVideoStream_FramesPublishedToBroadcaster(t *testing.T) {
	streamClient, _, b, _ := newTestServer(t, "secret")

	// Subscribe a browser client
	frames := b.Subscribe("browser1")
	defer b.Unsubscribe("browser1")

	ctx, cancel := context.WithCancel(authCtx("secret"))
	defer cancel()

	stream, err := streamClient.VideoStream(ctx)
	require.NoError(t, err)

	// Give server time to register camera with broadcaster
	time.Sleep(50 * time.Millisecond)

	frameData := []byte{0xFF, 0xD8, 0xFF, 0xE0}
	err = stream.Send(&dogcampb.FrameMessage{JpegData: frameData, TimestampMs: 1})
	require.NoError(t, err)

	select {
	case got := <-frames:
		assert.Equal(t, frameData, got)
	case <-time.After(500 * time.Millisecond):
		t.Fatal("frame not received by broadcaster subscriber")
	}
}

func TestReportMetrics_RejectsInvalidAPIKey(t *testing.T) {
	_, telClient, _, _ := newTestServer(t, "secret")
	_, err := telClient.ReportMetrics(authCtx("wrong"), &dogcampb.MetricsPayload{})
	require.Error(t, err)
	assert.Equal(t, codes.Unauthenticated, status.Code(err))
}

func TestReportMetrics_StoresPayload(t *testing.T) {
	_, telClient, _, ms := newTestServer(t, "secret")
	payload := &dogcampb.MetricsPayload{FramesCaptured: 42, FramesSent: 40, StreamingActive: true}
	_, err := telClient.ReportMetrics(authCtx("secret"), payload)
	require.NoError(t, err)

	got := ms.Get()
	require.NotNil(t, got)
	assert.Equal(t, int64(42), got.FramesCaptured)
}
