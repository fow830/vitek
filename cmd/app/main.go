package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"

	"github.com/jackc/pgx/v5/pgxpool"

	"vitek/internal/config"
	"vitek/internal/service"
	"vitek/internal/tokens"
	httpapi "vitek/internal/transport/httpapi"
)

func main() {
	cfg, err := config.Load()
	if err != nil {
		fmt.Fprintf(os.Stderr, "%s %s: %v\n", tokens.ProductName, tokens.BinaryAPI, err)
		os.Exit(1)
	}

	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	pool, err := pgxpool.New(ctx, cfg.DatabaseURL)
	if err != nil {
		log.Fatalf("db pool: %v", err)
	}
	defer pool.Close()

	local := cfg.AppEnv == tokens.AppEnvLocal
	srv := &http.Server{
		Addr: cfg.HTTPAddr,
		Handler: httpapi.NewServer(pool,
			httpapi.WithMagicLinkMailer(service.NewMemoryMagicLinkMailer()),
			httpapi.WithExposeMagicLinkTokens(local),
			httpapi.WithSecureCookies(!local),
		).Handler(),
		ReadHeaderTimeout: tokens.HTTPReadHeaderTimeout,
	}

	go func() {
		log.Printf("%s %s listening on %s env=%s", tokens.ProductName, tokens.BinaryAPI, cfg.HTTPAddr, cfg.AppEnv)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("http: %v", err)
		}
	}()

	<-ctx.Done()
	shutdownCtx, cancel := context.WithTimeout(context.Background(), tokens.HTTPShutdownTimeout)
	defer cancel()
	_ = srv.Shutdown(shutdownCtx)
}
