// billsplit/cmd/seed/main.go
package main

import (
	"context"
	"fmt"
	"log"

	"github.com/aws/aws-sdk-go-v2/aws"
	awsconfig "github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/fergalhk-lab/apps/billsplit/internal/config"
	"github.com/fergalhk-lab/apps/billsplit/internal/service"
	localstore "github.com/fergalhk-lab/apps/billsplit/internal/store"
)

func main() {
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("config: %v", err)
	}

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

	st := localstore.NewS3Store(s3Client, cfg.S3Bucket)
	invites := service.NewInviteService(st)

	code, err := invites.GenerateInvite(context.Background(), true)
	if err != nil {
		log.Fatalf("generate invite: %v", err)
	}
	fmt.Println(code)
}
