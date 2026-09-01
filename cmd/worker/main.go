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

	q := repository.New(pool)
	proxies := service.NewProxies(q)
	items := service.NewItems(q)

	log.Printf("%s %s started env=%s tick=%s", tokens.ProductName, tokens.BinaryWorker, cfg.AppEnv, cfg.WorkerTick)

	ticker := time.NewTicker(cfg.WorkerTick)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			log.Printf("%s %s shutting down", tokens.ProductName, tokens.BinaryWorker)
			return
		case <-ticker.C:
			runTick(ctx, proxies, items)
		}
	}
}

func runTick(ctx context.Context, proxies *service.Proxies, items *service.Items) {
	list, err := proxies.ListActive(ctx)
	if err != nil {
		log.Printf("proxies: %v", err)
		return
	}
	log.Printf("tick: active_proxies=%d (parser adapter not wired yet)", len(list))

	// Keep items service referenced so Phase B wiring stays complete; real ingest comes later.
	_ = items
}
