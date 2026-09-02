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

	proxies := service.NewProxies(pool)
	bindings := service.NewBindings(pool)
	if warns, err := service.ProxyPoolBootIssues(ctx, proxies, cfg.AppEnv); err != nil {
		log.Fatalf("proxy pool: %v", err)
	} else {
		for _, w := range warns {
			log.Printf("proxy pool: %s", w)
		}
	}
	processor, err := service.NewListingProcessor(pool, cfg.ListingSearchProcessor, cfg.AvitoHTTPBase, cfg.RodUserDataDir)
	if err != nil {
		log.Fatalf("listing processor: %v", err)
	}
	notify := service.NewNotifications(pool, nil)
	listing := service.NewListingSearchWorkerWithNotify(pool, processor, notify)

	log.Printf("%s %s (%s) started env=%s tick=%s", tokens.ProductName, tokens.ProductNameLocal, tokens.BinaryWorker, cfg.AppEnv, cfg.WorkerTick)

	ticker := time.NewTicker(cfg.WorkerTick)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			log.Printf("%s %s shutting down", tokens.ProductName, tokens.BinaryWorker)
			return
		case <-ticker.C:
			runTick(ctx, proxies, bindings, listing, notify)
		}
	}
}

func runTick(ctx context.Context, proxies *service.Proxies, bindings *service.Bindings, listing *service.ListingSearchWorker, notify *service.Notifications) {
	if ok, fail, err := service.ProbeActive(ctx, proxies, bindings, nil); err != nil {
		log.Printf(tokens.LogWorkerProxiesErr, err)
	} else {
		log.Printf(tokens.LogWorkerProxyProbe, ok, fail)
	}
	if _, err := listing.ProcessOne(ctx); err != nil {
		log.Printf(tokens.LogWorkerListingSearchErr, err)
	}
	if n, err := listing.ProcessWatchPolls(ctx); err != nil {
		log.Printf(tokens.LogWorkerWatchPollErr, err)
	} else if n > 0 {
		log.Printf(tokens.LogWorkerWatchPollDone, n)
	}
	if n, err := notify.ProcessOutbox(ctx, tokens.ListingSearchWatchDueLimit); err != nil {
		log.Printf(tokens.LogWorkerNotifyErr, err)
	} else if n > 0 {
		log.Printf(tokens.LogWorkerNotifyDone, n)
	}
}
