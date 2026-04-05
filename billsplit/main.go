// billsplit/main.go
package main

import (
	"context"
	"embed"
	"fmt"
	"io/fs"
	"log"
	"net/http"

	"github.com/aws/aws-sdk-go-v2/aws"
	awsconfig "github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/fergalhk-lab/apps/billsplit/internal/config"
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

	awsOpts := []func(*awsconfig.LoadOptions) error{}
	if cfg.S3Endpoint != "" {
		awsOpts = append(awsOpts, awsconfig.WithEndpointResolverWithOptions(
			aws.EndpointResolverWithOptionsFunc(func(svc, region string, _ ...interface{}) (aws.Endpoint, error) {
				return aws.Endpoint{URL: cfg.S3Endpoint, HostnameImmutable: true}, nil
			}),
		))
	}

	awsCfg, err := awsconfig.LoadDefaultConfig(context.Background(), awsOpts...)
	if err != nil {
		log.Fatalf("aws config: %v", err)
	}

	s3Client := s3.NewFromConfig(awsCfg, func(o *s3.Options) {
		if cfg.S3Endpoint != "" {
			o.UsePathStyle = true
		}
	})

	store := localstore.NewS3Store(s3Client, cfg.S3Bucket)

	svc := handler.Services{
		Auth:        service.NewAuthService(store, cfg.JWTSecret),
		Groups:      service.NewGroupService(store),
		Expenses:    service.NewExpenseService(store),
		Settlements: service.NewSettlementService(store),
		Invites:     service.NewInviteService(store),
	}

	apiRouter := handler.NewRouter(svc)

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
	log.Printf("listening on %s", addr)
	if err := http.ListenAndServe(addr, middleware.RecoverPanic(logger)(middleware.RequestLogger(logger)(mux))); err != nil {
		log.Fatalf("server: %v", err)
	}
}
