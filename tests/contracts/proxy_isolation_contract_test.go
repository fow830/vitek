package contracts_test

import (
	"context"
	"testing"

	"github.com/stretchr/testify/require"

	"vitek/internal/repository"
	"vitek/internal/service"
)

// CONTRACT-PROXY-001: ListActive returns only proxies with status ACTIVE.
func TestContract_ProxyIsolationActiveOnly(t *testing.T) {
	ctx := context.Background()
	pool, _ := queries(t)
	proxies := service.NewProxies(pool)

	_, err := proxies.Create(ctx, "http://active.example:8080", repository.ProxyStatusACTIVE, "")
	require.NoError(t, err)
	_, err = proxies.Create(ctx, "http://banned.example:8080", repository.ProxyStatusBANNED, "")
	require.NoError(t, err)
	_, err = proxies.Create(ctx, "http://disabled.example:8080", repository.ProxyStatusDISABLED, "")
	require.NoError(t, err)

	list, err := proxies.ListActive(ctx)
	require.NoError(t, err)
	require.Len(t, list, 1)
	require.Equal(t, repository.ProxyStatusACTIVE, list[0].Status)
	require.Equal(t, "http://active.example:8080", list[0].Endpoint)
}
