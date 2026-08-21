package main

import (
	"context"
	"os/signal"
	"syscall"

	"github.com/fergalhk-lab/apps/dogcam/cam/internal/capture"
	"github.com/fergalhk-lab/apps/dogcam/cam/internal/client"
	"github.com/fergalhk-lab/apps/dogcam/cam/internal/config"
	"github.com/fergalhk-lab/apps/dogcam/cam/internal/reporter"
	"github.com/fergalhk-lab/apps/dogcam/gen/dogcampb"
	"go.uber.org/zap"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/metadata"
)

func main() {
	log, _ := zap.NewProduction()
	defer log.Sync()

	cfg, err := config.Load()
	if err != nil {
		log.Fatal("failed to load config", zap.Error(err))
	}

	cap, err := capture.New(cfg.CameraDevice)
	if err != nil {
		log.Fatal("failed to init capturer", zap.Error(err))
	}

	conn, err := grpc.NewClient(cfg.ServerAddr,
		grpc.WithTransportCredentials(insecure.NewCredentials()),
	)
	if err != nil {
		log.Fatal("failed to create grpc client", zap.Error(err))
	}
	defer conn.Close()

	telClient := dogcampb.NewTelemetryServiceClient(conn)
	rep := reporter.New(reporterClientAdapter{telClient, cfg.CamAPIKey})

	c := client.New(conn, cfg.CamAPIKey, rep)

	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	log.Info("starting camera client", zap.String("server", cfg.ServerAddr))
	c.Run(ctx, cap)
	log.Info("camera client stopped")
}

// reporterClientAdapter wraps the gRPC TelemetryServiceClient to satisfy reporter.Client.
type reporterClientAdapter struct {
	c      dogcampb.TelemetryServiceClient
	apiKey string
}

func (a reporterClientAdapter) ReportMetrics(ctx context.Context, p *dogcampb.MetricsPayload) error {
	md := metadata.Pairs("authorization", "Bearer "+a.apiKey)
	ctx = metadata.NewOutgoingContext(ctx, md)
	_, err := a.c.ReportMetrics(ctx, p)
	return err
}
