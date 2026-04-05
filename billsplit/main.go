// billsplit/main.go
package main

import (
	"context"
	"embed"
	"errors"
	"fmt"
	"io/fs"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/fergalhk-lab/apps/billsplit/internal/config"
	"github.com/fergalhk-lab/apps/billsplit/internal/dependencies"
	"github.com/fergalhk-lab/apps/billsplit/internal/fxrates"
	"github.com/fergalhk-lab/apps/billsplit/internal/handler"
	"github.com/fergalhk-lab/apps/billsplit/internal/middleware"
	"github.com/fergalhk-lab/apps/billsplit/internal/service"
	localstore "github.com/fergalhk-lab/apps/billsplit/internal/store"
	"go.uber.org/zap"
)

// frontend/dist is populated by `npm run build` in billsplit/frontend/.
// Use `make build` (not `go build` directly) to ensure the frontend is built first.
//go:embed frontend/dist
var frontendDist embed.FS

func main() {
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("config: %v", err)
	}

	zapCfg := zap.NewProductionConfig()
	zapCfg.Sampling = nil
	logger, err := zapCfg.Build()
	if err != nil {
		log.Fatalf("logger: %v", err)
	}
	defer logger.Sync() //nolint:errcheck

	s3Client, err := dependencies.NewS3Client(context.Background())
	if err != nil {
		log.Fatalf("s3 client: %v", err)
	}

	store := localstore.NewS3Store(s3Client, cfg.S3Bucket)

	svc := handler.Services{
		Auth:        service.NewAuthService(store, cfg.JWTSecret),
		Groups:      service.NewGroupService(store),
		Expenses:    service.NewExpenseService(store),
		Settlements: service.NewSettlementService(store),
		Invites:     service.NewInviteService(store),
		FXRates:     fxrates.NewCache(store),
	}

	apiRouter := handler.NewRouter(svc, cfg.SecureCookie)

	distFS, err := fs.Sub(frontendDist, "frontend/dist")
	if err != nil {
		log.Fatalf("embed fs: %v", err)
	}
	fileServer := http.FileServer(http.FS(distFS))

	mux := http.NewServeMux()
	mux.Handle("/api/", apiRouter)
	mux.HandleFunc("/healthz", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})
	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		_, err := fs.Stat(distFS, r.URL.Path[1:])
		if err != nil {
			r.URL.Path = "/"
		}
		fileServer.ServeHTTP(w, r)
	})

	addr := fmt.Sprintf(":%s", cfg.Port)
	server := &http.Server{
		Addr:              addr,
		Handler:           middleware.RecoverPanic(logger)(middleware.RequestLogger(logger)(mux)),
		ReadHeaderTimeout: 5 * time.Second,
		ReadTimeout:       30 * time.Second,
		WriteTimeout:      60 * time.Second,
		IdleTimeout:       120 * time.Second,
	}

	sigCtx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	go func() {
		<-sigCtx.Done()
		stop()
		shutdownCtx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
		defer cancel()
		if err := server.Shutdown(shutdownCtx); err != nil {
			logger.Error("shutdown error", zap.Error(err))
		}
	}()

	log.Printf("listening on %s", addr)
	if err := server.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
		log.Fatalf("server: %v", err)
	}
	logger.Info("server stopped")
}
