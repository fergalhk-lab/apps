package grpchandler

import (
	"context"

	"github.com/fergalhk-lab/apps/dogcam/gen/dogcampb"
	"github.com/fergalhk-lab/apps/dogcam/server/internal/broadcast"
	"github.com/fergalhk-lab/apps/dogcam/server/internal/metricsstore"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
)

// StreamHandler implements dogcampb.StreamServiceServer.
type StreamHandler struct {
	dogcampb.UnimplementedStreamServiceServer
	broadcaster *broadcast.Broadcaster
	apiKey      string
}

func NewStreamHandler(b *broadcast.Broadcaster, apiKey string) *StreamHandler {
	return &StreamHandler{broadcaster: b, apiKey: apiKey}
}

func (h *StreamHandler) VideoStream(stream dogcampb.StreamService_VideoStreamServer) error {
	if err := h.validateKey(stream.Context()); err != nil {
		return err
	}

	controlCh := make(chan *dogcampb.ControlMessage, 4)
	h.broadcaster.RegisterCamera(controlCh)
	defer func() {
		h.broadcaster.UnregisterCamera()
		close(controlCh)
	}()

	// Forward control messages from broadcaster to camera.
	go func() {
		for msg := range controlCh {
			if err := stream.Send(msg); err != nil {
				return
			}
		}
	}()

	for {
		frame, err := stream.Recv()
		if err != nil {
			return err
		}
		h.broadcaster.Publish(frame.JpegData)
	}
}

// TelemetryHandler implements dogcampb.TelemetryServiceServer.
type TelemetryHandler struct {
	dogcampb.UnimplementedTelemetryServiceServer
	store  *metricsstore.Store
	apiKey string
}

func NewTelemetryHandler(s *metricsstore.Store, apiKey string) *TelemetryHandler {
	return &TelemetryHandler{store: s, apiKey: apiKey}
}

func (h *TelemetryHandler) ReportMetrics(ctx context.Context, p *dogcampb.MetricsPayload) (*dogcampb.MetricsAck, error) {
	if err := h.validateKey(ctx); err != nil {
		return nil, err
	}
	h.store.Update(p)
	return &dogcampb.MetricsAck{}, nil
}

func validateAPIKey(ctx context.Context, apiKey string) error {
	md, ok := metadata.FromIncomingContext(ctx)
	if !ok {
		return status.Error(codes.Unauthenticated, "missing metadata")
	}
	vals := md.Get("authorization")
	if len(vals) == 0 || vals[0] != "Bearer "+apiKey {
		return status.Error(codes.Unauthenticated, "invalid API key")
	}
	return nil
}

func (h *StreamHandler) validateKey(ctx context.Context) error {
	return validateAPIKey(ctx, h.apiKey)
}

func (h *TelemetryHandler) validateKey(ctx context.Context) error {
	return validateAPIKey(ctx, h.apiKey)
}
