package contracts_test

import (
	"context"
	"encoding/json"
	"fmt"
	"net/url"
	"strconv"
	"strings"
	"testing"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/stretchr/testify/require"

	"vitek/internal/config"
	"vitek/internal/domain"
	"vitek/internal/repository"
	"vitek/internal/service"
	"vitek/internal/tokens"
)

// CONTRACT-PROXY-012: health probe OK resets streak; N fails → DEAD (+ binding pause).
func TestContract_ProxyHealthProbe(t *testing.T) {
	ctx := context.Background()
	pool, q := queries(t)
	proxies := service.NewProxies(pool)
	bindings := service.NewBindings(pool)

	p, err := proxies.Create(ctx, "socks5://127.0.0.1:18080", repository.ProxyStatusACTIVE, "probe")
	require.NoError(t, err)

	ok, fail, err := service.ProbeActive(ctx, proxies, bindings, func(_ context.Context, _, _ string) error {
		return nil
	})
	require.NoError(t, err)
	require.Equal(t, 1, ok)
	require.Equal(t, 0, fail)

	got, err := q.GetProxy(ctx, p.ID)
	require.NoError(t, err)
	require.Equal(t, repository.ProxyHealthStatusOK, got.Health)
	require.Equal(t, int32(0), got.FailStreak)

	for i := 0; i < tokens.ProxyDeadAfterFails; i++ {
		_, _, err = service.ProbeActive(ctx, proxies, bindings, func(_ context.Context, _, _ string) error {
			return fmt.Errorf("%s", tokens.ErrMsgProxyProbeFailed)
		})
		require.NoError(t, err)
	}
	got, err = q.GetProxy(ctx, p.ID)
	require.NoError(t, err)
	// last fetchable proxy is clamped to DEGRADED (never left DEAD alone)
	require.Equal(t, repository.ProxyHealthStatusDEGRADED, got.Health)
	require.Equal(t, int32(tokens.ProxyDeadAfterFails), got.FailStreak)
}

// CONTRACT-PROXY-013: production boot rejects empty ACTIVE pool; DEAD-only / docker-bridge soft-warn.
func TestContract_ProxyPoolBootGuard(t *testing.T) {
	ctx := context.Background()
	pool, _ := queries(t)
	proxies := service.NewProxies(pool)

	warns, err := service.ProxyPoolBootIssues(ctx, proxies, tokens.AppEnvProduction)
	require.Error(t, err)
	require.Empty(t, warns)

	alive, err := proxies.Create(ctx, "socks5://172.19.0.1:11080", repository.ProxyStatusACTIVE, "bridge")
	require.NoError(t, err)
	warns, err = service.ProxyPoolBootIssues(ctx, proxies, tokens.AppEnvProduction)
	require.NoError(t, err)
	require.Equal(t, []string{tokens.ErrMsgProxyPoolDockerBridgeOnly}, warns)

	for i := 0; i < tokens.ProxyDeadAfterFails; i++ {
		_, err = proxies.RecordHealthFail(ctx, alive.ID, tokens.ErrMsgProxyProbeFailed)
		require.NoError(t, err)
	}
	warns, err = service.ProxyPoolBootIssues(ctx, proxies, tokens.AppEnvProduction)
	require.NoError(t, err)
	require.Contains(t, warns, tokens.ErrMsgProxyPoolAllDead)
	require.Contains(t, warns, tokens.ErrMsgProxyPoolDockerBridgeOnly)

	_, err = proxies.Create(ctx, "socks5://10.8.0.2:1080", repository.ProxyStatusACTIVE, "ok")
	require.NoError(t, err)
	warns, err = service.ProxyPoolBootIssues(ctx, proxies, tokens.AppEnvProduction)
	require.NoError(t, err)
	require.NotContains(t, warns, tokens.ErrMsgProxyPoolAllDead)
	require.Empty(t, warns) // has non-bridge ACTIVE (even if other is DEAD)
	require.True(t, tokens.IsDockerBridgeProxyEndpoint("socks5://172.19.0.1:11080"))
	require.False(t, tokens.IsDockerBridgeProxyEndpoint("socks5://10.8.0.2:1080"))
	require.Equal(t, tokens.DockerBridgeCIDR, "172.16.0.0/12")
}

// CONTRACT-PROXY-021/022 + SESSION-004: Rod uses READY binding proxy+profile; dead proxy pauses binding.
func TestContract_RodUsesReadyBinding(t *testing.T) {
	ctx := context.Background()
	pool, _ := queries(t)

	acc, err := service.NewAvitoAccounts(pool).CreateWithSecret(ctx, "rod-bind", repository.AvitoAccountStatusACTIVE, "l", "s")
	require.NoError(t, err)
	proxy, err := service.NewProxies(pool).Create(ctx, tokens.FixtureRodProxyEndpoint, repository.ProxyStatusACTIVE, "rp")
	require.NoError(t, err)
	bindings := service.NewBindings(pool)
	b, err := bindings.Create(ctx, acc.ID, proxy.ID, tokens.FixtureBindingUserDataDir)
	require.NoError(t, err)
	_, err = bindings.WarmSession(ctx, b.ID, &contractPageFetcher{html: tokens.FixtureWarmSessionHTML})
	require.NoError(t, err)

	fetcher := &contractPageFetcher{html: tokens.FixtureAvitoFilterSERPHTML}
	proc := service.NewRodAvitoListingProcessor(pool, fetcher)
	out, err := proc.FindSimilar(ctx, tokens.FixtureFilterListingURL)
	require.NoError(t, err)
	require.Len(t, out, 2)
	require.Equal(t, tokens.FixtureRodProxyEndpoint, fetcher.lastProxy)
	require.Equal(t, tokens.FixtureBindingUserDataDir, fetcher.lastUserDataDir)

	proxies := service.NewProxies(pool)
	for i := 0; i < tokens.ProxyDeadAfterFails; i++ {
		_, _, err = service.ProbeActive(ctx, proxies, bindings, func(_ context.Context, _, _ string) error {
			return fmt.Errorf("%s", tokens.ErrMsgProxyProbeFailed)
		})
		require.NoError(t, err)
	}
	require.NoError(t, bindings.PauseForProxy(ctx, proxy.ID))

	got, err := bindings.List(ctx)
	require.NoError(t, err)
	require.Equal(t, repository.ListingBindingStatusPAUSED, got[0].Status)

	_, err = proc.FindSimilar(ctx, tokens.FixtureFilterListingURL)
	require.ErrorIs(t, err, domain.ErrListingSearchNoBinding)
}

// CONTRACT-SESSION-001/003: session enum + WarmSession via fetcher (no secrets in JSON).
func TestContract_SessionWarmViaFetcher(t *testing.T) {
	require.Equal(t, tokens.SchemaListingSessionStatuses, []string{
		tokens.ListingSessionLoggedOut,
		tokens.ListingSessionLoggingIn,
		tokens.ListingSessionReady,
		tokens.ListingSessionChallenge,
		tokens.ListingSessionError,
	})

	ctx := context.Background()
	pool, _ := queries(t)
	acc, err := service.NewAvitoAccounts(pool).CreateWithSecret(ctx, "warm-acc", repository.AvitoAccountStatusACTIVE, "login@x", "sekrit")
	require.NoError(t, err)
	proxy, err := service.NewProxies(pool).Create(ctx, "socks5://127.0.0.1:9", repository.ProxyStatusACTIVE, "wp")
	require.NoError(t, err)
	bindings := service.NewBindings(pool)
	b, err := bindings.Create(ctx, acc.ID, proxy.ID, tokens.FixtureBindingUserDataDir)
	require.NoError(t, err)

	ready, err := bindings.WarmSession(ctx, b.ID, &contractPageFetcher{html: tokens.FixtureWarmSessionHTML})
	require.NoError(t, err)
	require.Equal(t, repository.ListingSessionStatusREADY, ready.SessionStatus)
	js := service.BindingJSON(ready)
	require.NotContains(t, fmt.Sprintf("%v", js), "sekrit")
	require.NotContains(t, fmt.Sprintf("%v", js), "login@x")

	challenged, err := bindings.WarmSession(ctx, b.ID, &contractPageFetcher{html: tokens.FixtureAvitoChallengeHTML})
	require.NoError(t, err)
	require.Equal(t, repository.ListingSessionStatusCHALLENGE, challenged.SessionStatus)
	require.Equal(t, tokens.ErrMsgSessionChallenge, *challenged.SessionErr)

	failed, err := bindings.WarmSession(ctx, b.ID, &contractPageFetcher{err: fmt.Errorf("boom")})
	require.Error(t, err)
	require.Equal(t, repository.ListingSessionStatusERROR, failed.SessionStatus)
}

// CONTRACT-LISTING-030/031: SERP multi-page depth > 10.
func TestContract_RodSERPMultiPageDepth(t *testing.T) {
	require.Greater(t, tokens.ListingSearchSERPMaxItems, 10)
	require.Greater(t, tokens.ListingSearchSERPMaxPages, 1)
	require.Equal(t, tokens.RodInterRequestDelay, 200*time.Millisecond)

	ctx := context.Background()
	pool, _ := queries(t)
	seedReadyRodBinding(t, pool)

	pages := map[int]string{
		1: serpHTMLPage(1, 8),
		2: serpHTMLPage(9, 6),
	}
	fetcher := &pagingRodFetcher{pages: pages}
	proc := service.NewRodAvitoListingProcessor(pool, fetcher)
	out, err := proc.FindSimilar(ctx, tokens.FixtureFilterListingURL)
	require.NoError(t, err)
	require.GreaterOrEqual(t, len(out), 12)
	require.GreaterOrEqual(t, len(fetcher.calls), 2)
	require.Contains(t, fetcher.calls[1], tokens.ListingSearchSERPPageParam+"=")
}

// CONTRACT-LISTING-036 + meta from SERP HTML same channel.
func TestContract_RodMetaFromSERPHTML(t *testing.T) {
	ctx := context.Background()
	pool, q := queries(t)
	seedReadyRodBinding(t, pool)

	fetcher := &contractPageFetcher{html: tokens.FixtureAvitoFilterSERPHTML}
	proc := service.NewRodAvitoListingProcessor(pool, fetcher)
	worker := service.NewListingSearchWorker(pool, proc)
	user, err := service.NewUsers(pool).CreateUser(ctx, tokens.ProductEmail("meta-rod"), repository.PlanTypePRO)
	require.NoError(t, err)
	watches := service.NewFilterWatches(pool)
	watch, err := watches.Start(ctx, user.ID, tokens.FixtureFilterListingURL)
	require.NoError(t, err)

	_, err = worker.ProcessWatchPolls(ctx)
	require.NoError(t, err)

	got, err := q.GetFilterWatch(ctx, watch.ID)
	require.NoError(t, err)
	require.Equal(t, repository.ListingWatchMetaStatusREADY, got.MetaStatus)
	var meta tokens.ListingFilterMeta
	require.NoError(t, json.Unmarshal(got.MetaJson, &meta))
	require.Equal(t, tokens.FixtureAvitoLocationID, fmt.Sprintf("%v", meta.Extras[tokens.AvitoJSONFieldLocationID]))
	require.Equal(t, tokens.FixtureAvitoCategoryID, fmt.Sprintf("%v", meta.Extras[tokens.AvitoJSONFieldCategoryID]))

	metaDirect := tokens.MergeFilterMetaFromSERPHTML(tokens.ParseListingFilterMeta(tokens.FixtureFilterListingURL), tokens.FixtureAvitoFilterSERPHTML)
	require.Equal(t, tokens.FixtureAvitoLocationID, fmt.Sprintf("%v", metaDirect.Extras[tokens.AvitoJSONFieldLocationID]))
}

// CONTRACT-ROD-PROD-002: production + rod requires ROD_USER_DATA_DIR.
func TestContract_RodProdRequiresUserDataDir(t *testing.T) {
	t.Setenv(tokens.EnvAppEnv, tokens.AppEnvProduction)
	t.Setenv(tokens.EnvDatabaseURL, "postgres://x")
	t.Setenv(tokens.EnvListingSearchProcessor, tokens.ListingSearchProcessorRod)
	t.Setenv(tokens.EnvRodUserDataDir, "")
	_, err := config.Load()
	require.Error(t, err)
	require.Contains(t, err.Error(), tokens.EnvRodUserDataDir)

	t.Setenv(tokens.EnvRodUserDataDir, tokens.FixtureBindingUserDataDir)
	cfg, err := config.Load()
	require.NoError(t, err)
	require.Equal(t, tokens.FixtureBindingUserDataDir, cfg.RodUserDataDir)
	require.Equal(t, tokens.DefaultRodUserDataDir, "/var/lib/vitek/rod")
	require.Equal(t, tokens.ImageWorkerRuntime, "chromedp/headless-shell:latest")
}

type contractPageFetcher struct {
	html            string
	err             error
	lastProxy       string
	lastUserDataDir string
	lastURL         string
}

func (f *contractPageFetcher) FetchHTML(_ context.Context, proxyEndpoint, userDataDir, pageURL string) (string, error) {
	f.lastProxy = proxyEndpoint
	f.lastUserDataDir = userDataDir
	f.lastURL = pageURL
	if f.err != nil {
		return "", f.err
	}
	if f.html != "" {
		return f.html, nil
	}
	return tokens.FixtureWarmSessionHTML, nil
}

type pagingRodFetcher struct {
	pages map[int]string
	calls []string
}

func (f *pagingRodFetcher) FetchHTML(_ context.Context, _, _, pageURL string) (string, error) {
	f.calls = append(f.calls, pageURL)
	u, err := url.Parse(pageURL)
	if err != nil {
		return "", err
	}
	page := 1
	if raw := u.Query().Get(tokens.ListingSearchSERPPageParam); raw != "" {
		page, _ = strconv.Atoi(raw)
	}
	html, ok := f.pages[page]
	if !ok {
		return tokens.FixtureEmptySERPHTML, nil
	}
	return html, nil
}

func serpHTMLPage(startID, n int) string {
	var b strings.Builder
	b.WriteString(`<!doctype html><html><body>`)
	b.WriteString(`"locationId":` + tokens.FixtureAvitoLocationID + `,"categoryId":` + tokens.FixtureAvitoCategoryID)
	for i := 0; i < n; i++ {
		id := strconv.Itoa(7000000000 + startID + i)
		title := "iPhone " + strconv.Itoa(startID+i)
		b.WriteString(`<div ` + tokens.AvitoSERPAttrItemID + `="` + id + `" data-marker="` + tokens.AvitoSERPMarkerItem + `">`)
		b.WriteString(`<h3 data-marker="` + tokens.AvitoSERPMarkerItemTitle + `">` + title + `</h3></div>`)
	}
	b.WriteString(`</body></html>`)
	return b.String()
}

func seedReadyRodBinding(t *testing.T, pool *pgxpool.Pool) {
	t.Helper()
	ctx := context.Background()
	acc, err := service.NewAvitoAccounts(pool).CreateWithSecret(ctx, "seed-rod", repository.AvitoAccountStatusACTIVE, "l", "s")
	require.NoError(t, err)
	proxy, err := service.NewProxies(pool).Create(ctx, tokens.FixtureRodProxyEndpoint, repository.ProxyStatusACTIVE, "seed-p")
	require.NoError(t, err)
	bindings := service.NewBindings(pool)
	b, err := bindings.Create(ctx, acc.ID, proxy.ID, tokens.FixtureBindingUserDataDir)
	require.NoError(t, err)
	_, err = bindings.WarmSession(ctx, b.ID, &contractPageFetcher{html: tokens.FixtureWarmSessionHTML})
	require.NoError(t, err)
}
