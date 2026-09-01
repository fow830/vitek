package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"

	"vitek/internal/config"
	"vitek/internal/repository"
	"vitek/internal/service"
	"vitek/internal/tokens"
)

func main() {
	cfg, err := config.Load()
	if err != nil {
		fmt.Fprintf(os.Stderr, "%s %s: %v\n", tokens.ProductName, tokens.BinaryWorker, err)
		os.Exit(1)
	}

	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	pool, err := pgxpool.New(ctx, cfg.DatabaseURL)
	if err != nil {
		log.Fatalf("db pool: %v", err)
	}
	defer pool.Close()

	proxies := service.NewProxies(repository.New(pool))

	log.Printf("%s %s (%s) started env=%s tick=%s", tokens.ProductName, tokens.ProductNameLocal, tokens.BinaryWorker, cfg.AppEnv, cfg.WorkerTick)

	ticker := time.NewTicker(cfg.WorkerTick)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			log.Printf("%s %s shutting down", tokens.ProductName, tokens.BinaryWorker)
			return
		case <-ticker.C:
			runTick(ctx, proxies)
		}
	}
}

func runTick(ctx context.Context, proxies *service.Proxies) {
	list, err := proxies.ListActive(ctx)
	if err != nil {
		log.Printf(tokens.LogWorkerProxiesErr, err)
		return
	}
	log.Printf(tokens.LogWorkerTickActive, len(list))
}
