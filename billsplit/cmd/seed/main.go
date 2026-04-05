// billsplit/cmd/seed/main.go
package main

import (
	"context"
	"flag"
	"fmt"
	"log"

	"github.com/fergalhk-lab/apps/billsplit/internal/config"
	"github.com/fergalhk-lab/apps/billsplit/internal/dependencies"
	"github.com/fergalhk-lab/apps/billsplit/internal/service"
	localstore "github.com/fergalhk-lab/apps/billsplit/internal/store"
)

func main() {
	initial := flag.Bool("initial", false, "only create an invite token if no tokens have ever been created (including used ones)")
	flag.Parse()

	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("config: %v", err)
	}

	s3Client, err := dependencies.NewS3Client(context.Background())
	if err != nil {
		log.Fatalf("s3 client: %v", err)
	}

	st := localstore.NewS3Store(s3Client, cfg.S3Bucket)
	invites := service.NewInviteService(st)
	ctx := context.Background()

	if *initial {
		has, err := invites.HasInvites(ctx)
		if err != nil {
			log.Fatalf("check invites: %v", err)
		}
		if has {
			log.Println("invites already exist, skipping")
			return
		}
	}

	code, err := invites.GenerateInvite(ctx, true)
	if err != nil {
		log.Fatalf("generate invite: %v", err)
	}
	fmt.Println(code)
}
