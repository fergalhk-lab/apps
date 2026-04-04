// billsplit/cmd/fxrates-sync/main.go
package main

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	awsconfig "github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/fergalhk-lab/apps/billsplit/internal/config"
	"github.com/fergalhk-lab/apps/billsplit/internal/fxrates"
	"github.com/fergalhk-lab/apps/billsplit/internal/fxrates/provider"
	localstore "github.com/fergalhk-lab/apps/billsplit/internal/store"
)

const apiURL = "https://open.er-api.com/v6/latest/USD"

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
	ctx := context.Background()

	resp, err := http.Get(apiURL)
	if err != nil {
		log.Fatalf("fetch rates: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		log.Fatalf("fetch rates: unexpected status %s", resp.Status)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		log.Fatalf("read response body: %v", err)
	}

	rates, err := provider.Parse(body)
	if err != nil {
		log.Fatalf("parse rates: %v", err)
	}

	data, err := json.Marshal(rates)
	if err != nil {
		log.Fatalf("marshal rates: %v", err)
	}

	if err := st.ForceWriteObject(ctx, fxrates.S3Key, data); err != nil {
		log.Fatalf("write rates: %v", err)
	}

	fmt.Printf("synced %d exchange rates (provider updated %s)\n",
		len(rates.Rates), rates.ProviderUpdatedAt.Format(time.RFC3339))
}
