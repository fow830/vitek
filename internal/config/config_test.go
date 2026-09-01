package config_test

import (
	"testing"

	"github.com/stretchr/testify/require"

	"vitek/internal/config"
	"vitek/internal/tokens"
)

func TestLoad_Defaults(t *testing.T) {
	for _, k := range []string{
		tokens.EnvAppEnv, tokens.EnvHTTPAddr, tokens.EnvLogLevel,
		tokens.EnvDatabaseURL, tokens.EnvRedisURL,
	} {
		t.Setenv(k, "")
	}

	cfg, err := config.Load()
	require.NoError(t, err)
	require.Equal(t, tokens.DefaultAppEnv, cfg.AppEnv)
	require.Equal(t, tokens.DefaultHTTPAddr(), cfg.HTTPAddr)
	require.Equal(t, tokens.DefaultLogLevel, cfg.LogLevel)
	require.Equal(t, tokens.DefaultDatabaseURL(), cfg.DatabaseURL)
	require.Equal(t, tokens.DefaultRedisURL(), cfg.RedisURL)
}

func TestLoad_RejectsBadAppEnv(t *testing.T) {
	t.Setenv(tokens.EnvAppEnv, "lab")
	t.Setenv(tokens.EnvDatabaseURL, tokens.DefaultDatabaseURL())
	t.Setenv(tokens.EnvRedisURL, tokens.DefaultRedisURL())

	_, err := config.Load()
	require.Error(t, err)
}

func TestLoad_RejectsBadLogLevel(t *testing.T) {
	t.Setenv(tokens.EnvAppEnv, tokens.AppEnvLocal)
	t.Setenv(tokens.EnvLogLevel, "verbose")
	t.Setenv(tokens.EnvDatabaseURL, tokens.DefaultDatabaseURL())
	t.Setenv(tokens.EnvRedisURL, tokens.DefaultRedisURL())

	_, err := config.Load()
	require.Error(t, err)
}
