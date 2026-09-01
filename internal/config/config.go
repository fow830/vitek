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
	AppEnv      string
	HTTPAddr    string
	LogLevel    string
	DatabaseURL string
	RedisURL    string
	WorkerTick  time.Duration
}

// Load reads process environment. Missing optional keys fall back to tokens.Default*.
func Load() (Config, error) {
	tickRaw := get(tokens.EnvWorkerTick, tokens.DefaultWorkerTick)
	tick, err := time.ParseDuration(tickRaw)
	if err != nil {
		return Config{}, fmt.Errorf("%s: %w", tokens.EnvWorkerTick, err)
	}

	cfg := Config{
		AppEnv:      get(tokens.EnvAppEnv, tokens.DefaultAppEnv),
		HTTPAddr:    get(tokens.EnvHTTPAddr, tokens.DefaultHTTPAddr()),
		LogLevel:    get(tokens.EnvLogLevel, tokens.DefaultLogLevel),
		DatabaseURL: get(tokens.EnvDatabaseURL, tokens.DefaultDatabaseURL()),
		RedisURL:    get(tokens.EnvRedisURL, tokens.DefaultRedisURL()),
		WorkerTick:  tick,
	}

	switch cfg.AppEnv {
	case tokens.AppEnvLocal, tokens.AppEnvStaging, tokens.AppEnvProduction:
	default:
		return Config{}, fmt.Errorf("%s: unsupported value %q", tokens.EnvAppEnv, cfg.AppEnv)
	}

	switch cfg.LogLevel {
	case tokens.LogLevelDebug, tokens.LogLevelInfo, tokens.LogLevelWarn, tokens.LogLevelError:
	default:
		return Config{}, fmt.Errorf("%s: unsupported value %q", tokens.EnvLogLevel, cfg.LogLevel)
	}

	if strings.TrimSpace(cfg.DatabaseURL) == "" {
		return Config{}, fmt.Errorf("%s is required", tokens.EnvDatabaseURL)
	}
	if strings.TrimSpace(cfg.RedisURL) == "" {
		return Config{}, fmt.Errorf("%s is required", tokens.EnvRedisURL)
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
