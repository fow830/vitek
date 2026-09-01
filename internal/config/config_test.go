package config_test

import (
	"testing"

	"github.com/stretchr/testify/require"

	"vitek/internal/config"
	"vitek/internal/tokens"
)

func TestLoad_Defaults(t *testing.T) {
	for _, k := range []string{
		tokens.EnvAppEnv, tokens.EnvHTTPAddr,
		tokens.EnvDatabaseURL, tokens.EnvWorkerTick,
	} {
		t.Setenv(k, "")
	}

	cfg, err := config.Load()
	require.NoError(t, err)
	require.Equal(t, tokens.DefaultAppEnv, cfg.AppEnv)
	require.Equal(t, tokens.DefaultHTTPAddr(), cfg.HTTPAddr)
	require.Equal(t, tokens.DefaultDatabaseURL(), cfg.DatabaseURL)
	require.Equal(t, tokens.DefaultWorkerTick, cfg.WorkerTick.String())
}

func TestLoad_RejectsBadAppEnv(t *testing.T) {
	t.Setenv(tokens.EnvAppEnv, "lab")
	t.Setenv(tokens.EnvDatabaseURL, tokens.DefaultDatabaseURL())

	_, err := config.Load()
	require.Error(t, err)
}
