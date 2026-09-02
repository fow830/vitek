package config

import (
	"fmt"
	"os"
	"strings"
	"time"

	"vitek/internal/tokens"
)

// Config is runtime configuration loaded exclusively via tokens.Env* keys.
type Config struct {
	AppEnv                  string
	HTTPAddr                string
	DatabaseURL             string
	WorkerTick              time.Duration
	ListingSearchProcessor  string
	AvitoHTTPBase           string
}

// Load reads process environment. Missing optional keys fall back to tokens.Default*.
func Load() (Config, error) {
	tickRaw := get(tokens.EnvWorkerTick, tokens.DefaultWorkerTick)
	tick, err := time.ParseDuration(tickRaw)
	if err != nil {
		return Config{}, fmt.Errorf("%s: %w", tokens.EnvWorkerTick, err)
	}

	cfg := Config{
		AppEnv:                 get(tokens.EnvAppEnv, tokens.DefaultAppEnv),
		HTTPAddr:               get(tokens.EnvHTTPAddr, tokens.DefaultHTTPAddr()),
		DatabaseURL:            get(tokens.EnvDatabaseURL, tokens.DefaultDatabaseURL()),
		WorkerTick:             tick,
		ListingSearchProcessor: get(tokens.EnvListingSearchProcessor, tokens.DefaultListingSearchProcessor),
		AvitoHTTPBase:          get(tokens.EnvAvitoHTTPBase, tokens.AvitoHTTPSBase),
	}

	switch cfg.ListingSearchProcessor {
	case tokens.ListingSearchProcessorStub, tokens.ListingSearchProcessorAvito, tokens.ListingSearchProcessorRod:
	default:
		return Config{}, fmt.Errorf("%s: unsupported value %q", tokens.EnvListingSearchProcessor, cfg.ListingSearchProcessor)
	}

	switch cfg.AppEnv {
	case tokens.AppEnvLocal, tokens.AppEnvStaging, tokens.AppEnvProduction:
	default:
		return Config{}, fmt.Errorf("%s: unsupported value %q", tokens.EnvAppEnv, cfg.AppEnv)
	}

	if strings.TrimSpace(cfg.DatabaseURL) == "" {
		return Config{}, fmt.Errorf("%s is required", tokens.EnvDatabaseURL)
	}
	if cfg.WorkerTick <= 0 {
		return Config{}, fmt.Errorf("%s must be > 0", tokens.EnvWorkerTick)
	}

	return cfg, nil
}

func get(key, fallback string) string {
	if v, ok := os.LookupEnv(key); ok && strings.TrimSpace(v) != "" {
		return v
	}
	return fallback
}
