package main

import (
	"context"
	"errors"
	"net"
	"net/http"
	"os/signal"
	"sync"
	"syscall"
	"time"

	"github.com/fergalhk-lab/apps/dogcam/gen/dogcampb"
	"github.com/fergalhk-lab/apps/dogcam/server/internal/broadcast"
	"github.com/fergalhk-lab/apps/dogcam/server/internal/config"
	"github.com/fergalhk-lab/apps/dogcam/server/internal/grpchandler"
	"github.com/fergalhk-lab/apps/dogcam/server/internal/httpserver"
	"github.com/fergalhk-lab/apps/dogcam/server/internal/metricsstore"
	"go.uber.org/zap"
	"google.golang.org/grpc"
)

func main() {
	log, _ := zap.NewProduction()
	defer log.Sync()

	cfg, err := config.Load()
	if err != nil {
		log.Fatal("failed to load config", zap.Error(err))
	}

	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	b := broadcast.New(cfg.FrameIntervalMs)
	ms := metricsstore.New()

	// gRPC server
	grpcSrv := grpc.NewServer()
	dogcampb.RegisterStreamServiceServer(grpcSrv, grpchandler.NewStreamHandler(b, cfg.CamAPIKey))
	dogcampb.RegisterTelemetryServiceServer(grpcSrv, grpchandler.NewTelemetryHandler(ms, cfg.CamAPIKey))

	grpcLis, err := net.Listen("tcp", ":"+cfg.GRPCPort)
	if err != nil {
		log.Fatal("failed to listen on grpc port", zap.Error(err))
	}

	// HTTP server (browsers)
	httpSrv := &http.Server{
		Addr:    ":" + cfg.HTTPPort,
		Handler: httpserver.New(b, cfg.ViewerPassword).Handler(),
	}

	// Metrics server (internal, no auth)
	metricsSrv := &http.Server{
		Addr:    ":" + cfg.MetricsPort,
		Handler: metricsstore.Handler(ms),
	}

	go func() {
		log.Info("grpc server listening", zap.String("port", cfg.GRPCPort))
		if err := grpcSrv.Serve(grpcLis); err != nil {
			log.Error("grpc server stopped", zap.Error(err))
		}
	}()

	go func() {
		log.Info("http server listening", zap.String("port", cfg.HTTPPort))
		if err := httpSrv.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			log.Error("http server stopped", zap.Error(err))
		}
	}()

	go func() {
		log.Info("metrics server listening", zap.String("port", cfg.MetricsPort))
		if err := metricsSrv.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			log.Error("metrics server stopped", zap.Error(err))
		}
	}()

	<-ctx.Done()
	log.Info("shutting down")
	shutCtx, shutCancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer shutCancel()
	var wg sync.WaitGroup
	wg.Add(3)
	go func() { defer wg.Done(); grpcSrv.GracefulStop() }()
	go func() { defer wg.Done(); _ = httpSrv.Shutdown(shutCtx) }()
	go func() { defer wg.Done(); _ = metricsSrv.Shutdown(shutCtx) }()
	wg.Wait()
	log.Info("shutdown complete")
}
