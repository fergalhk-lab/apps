package client

import (
	"context"
	"time"

	"github.com/fergalhk-lab/apps/dogcam/cam/internal/capture"
	"github.com/fergalhk-lab/apps/dogcam/cam/internal/reporter"
	"github.com/fergalhk-lab/apps/dogcam/gen/dogcampb"
	"google.golang.org/grpc"
	"google.golang.org/grpc/metadata"
)

type Client struct {
	conn      *grpc.ClientConn
	apiKey    string
	reporter  *reporter.Reporter
	startedAt time.Time
}

func New(conn *grpc.ClientConn, apiKey string, r *reporter.Reporter) *Client {
	return &Client{conn: conn, apiKey: apiKey, reporter: r, startedAt: time.Now()}
}

// Run connects and streams forever, reconnecting with exponential backoff on error.
func (c *Client) Run(ctx context.Context, cap capture.Capturer) {
	backoff := time.Second
	for {
		if err := c.RunOnce(ctx, cap); err != nil {
			if ctx.Err() != nil {
				return
			}
			time.Sleep(backoff)
			if backoff < 30*time.Second {
				backoff *= 2
			}
			continue
		}
		backoff = time.Second
	}
}

// RunOnce opens a single VideoStream RPC and handles START/STOP signals.
// Returns when the stream ends or ctx is cancelled.
func (c *Client) RunOnce(ctx context.Context, cap capture.Capturer) error {
	md := metadata.Pairs("authorization", "Bearer "+c.apiKey)
	streamCtx := metadata.NewOutgoingContext(ctx, md)

	streamClient := dogcampb.NewStreamServiceClient(c.conn)
	stream, err := streamClient.VideoStream(streamCtx)
	if err != nil {
		return err
	}

	var (
		framesCaptured int64
		framesSent     int64
	)

	var captureCancel context.CancelFunc
	defer func() {
		if captureCancel != nil {
			captureCancel()
		}
	}()

	for {
		msg, err := stream.Recv()
		if err != nil {
			return err
		}

		switch msg.Command {
		case dogcampb.ControlMessage_START:
			if captureCancel != nil {
				captureCancel()
			}
			var captureCtx context.Context
			captureCtx, captureCancel = context.WithCancel(ctx)
			interval := time.Duration(msg.FrameIntervalMs) * time.Millisecond
			go c.captureLoop(captureCtx, stream, cap, interval, &framesCaptured, &framesSent)

		case dogcampb.ControlMessage_STOP:
			if captureCancel != nil {
				captureCancel()
				captureCancel = nil
			}
			_ = c.reporter.Report(ctx, c.buildMetrics(framesCaptured, framesSent, false, ""))
		}
	}
}

func (c *Client) captureLoop(
	ctx context.Context,
	stream dogcampb.StreamService_VideoStreamClient,
	cap capture.Capturer,
	interval time.Duration,
	captured, sent *int64,
) {
	ticker := time.NewTicker(interval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			frame, err := cap.Capture()
			lastErr := ""
			if err != nil {
				lastErr = err.Error()
			} else {
				*captured++
				if sendErr := stream.Send(&dogcampb.FrameMessage{
					JpegData:    frame,
					TimestampMs: time.Now().UnixMilli(),
				}); sendErr != nil {
					return
				}
				*sent++
			}
			_ = c.reporter.Report(ctx, c.buildMetrics(*captured, *sent, true, lastErr))
		}
	}
}

func (c *Client) buildMetrics(captured, sent int64, streaming bool, lastErr string) *dogcampb.MetricsPayload {
	return &dogcampb.MetricsPayload{
		FramesCaptured:  captured,
		FramesSent:      sent,
		StreamingActive: streaming,
		UptimeSeconds:   int64(time.Since(c.startedAt).Seconds()),
		LastError:       lastErr,
	}
}
