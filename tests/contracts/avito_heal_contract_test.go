package contracts_test

import (
	"context"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"sync/atomic"
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	"vitek/internal/domain"
	"vitek/internal/repository"
	"vitek/internal/service"
	"vitek/internal/tokens"
)

// CONTRACT-AVITO-HEAL-001: HTTP status+body classifies into tokenized fetch states.
func TestContract_AvitoHeal_ClassifyResponse(t *testing.T) {
	require.Equal(t, tokens.AvitoFetchStateContent, tokens.ClassifyAvitoHTTPResponse(http.StatusOK, tokens.FixtureAvitoFilterSERPHTML))
	require.Equal(t, tokens.AvitoFetchStateContent, tokens.ClassifyAvitoHTTPResponse(http.StatusOK, tokens.FixtureAvitoMFECatalogHTML))
	require.Equal(t, tokens.AvitoFetchStateEmpty, tokens.ClassifyAvitoHTTPResponse(http.StatusOK, "<html><body>no cards</body></html>"))
	require.Equal(t, tokens.AvitoFetchStateChallenge, tokens.ClassifyAvitoHTTPResponse(http.StatusOK, tokens.FixtureAvitoChallengeHTML))
	require.Equal(t, tokens.AvitoFetchStateRateLimited, tokens.ClassifyAvitoHTTPResponse(http.StatusTooManyRequests, tokens.FixtureAvitoIPBlockHTML))
	require.Equal(t, tokens.AvitoFetchStateChallengePoW, tokens.ClassifyAvitoHTTPResponse(tokens.AvitoHTTPStatusPoW, tokens.FixtureAvitoPoWHTML))
	require.Equal(t, tokens.AvitoFetchStateHeaderTooLarge, tokens.ClassifyAvitoHTTPResponse(tokens.AvitoHTTPStatusHeaderTooLarge, tokens.FixtureAvitoHeaderTooLargeHTML))
	require.Equal(t, tokens.AvitoFetchStateForbidden, tokens.ClassifyAvitoHTTPResponse(http.StatusForbidden, "Доступ ограничен"))
	require.True(t, tokens.AvitoFetchStateHealable(tokens.AvitoFetchStateRateLimited))
	require.True(t, tokens.AvitoFetchStateHealable(tokens.AvitoFetchStateChallengePoW))
	require.True(t, tokens.AvitoFetchStateHealable(tokens.AvitoFetchStateHeaderTooLarge))
	require.True(t, tokens.AvitoFetchStateHealable(tokens.AvitoFetchStateChallenge))
	require.False(t, tokens.AvitoFetchStateHealable(tokens.AvitoFetchStateContent))
	require.False(t, tokens.AvitoFetchStateHealable(tokens.AvitoFetchStateEmpty))
}

// CONTRACT-AVITO-HEAL-002: storage_state cookies are slim allowlist only (never fat dump).
func TestContract_AvitoHeal_SlimStorageStateCookies(t *testing.T) {
	dir := t.TempDir()
	raw := []byte(`{"cookies":[` +
		`{"name":"srv_id","value":"ok1","domain":".avito.ru","path":"/","expires":-1,"secure":true},` +
		`{"name":"ft","value":"ok2","domain":".avito.ru","path":"/","expires":-1,"secure":true},` +
		`{"name":"qrator_jsid","value":"ok3","domain":".avito.ru","path":"/","expires":-1,"secure":true},` +
		`{"name":"_ga","value":"fat","domain":".avito.ru","path":"/","expires":-1,"secure":false},` +
		`{"name":"yandexuid","value":"fat2","domain":".avito.ru","path":"/","expires":-1,"secure":false}` +
		`]}`)
	require.NoError(t, os.WriteFile(filepath.Join(dir, tokens.AvitoStorageStateFile), raw, 0o600))
	header := service.CookieHeaderFromStorageState(dir, tokens.AvitoHTTPSBase+"/")
	require.Contains(t, header, "srv_id=ok1")
	require.Contains(t, header, "ft=ok2")
	require.Contains(t, header, "qrator_jsid=ok3")
	require.NotContains(t, header, "_ga=")
	require.NotContains(t, header, "yandexuid=")
	require.LessOrEqual(t, len(header), tokens.AvitoCookieHeaderMaxBytes)
	require.Contains(t, tokens.AvitoCookieAllowlist, "srv_id")
	require.Contains(t, tokens.AvitoCookieAllowlist, "ft")
	require.Contains(t, tokens.AvitoCookieAllowlist, "f")
	require.Contains(t, tokens.AvitoCookieAllowlist, "qrator_jsid")
}

// CONTRACT-AVITO-HEAL-003: MFE catalog JSON yields listing items.
func TestContract_AvitoHeal_ParseMFECatalog(t *testing.T) {
	items, err := tokens.ParseAvitoSERPItems(tokens.FixtureAvitoMFECatalogHTML)
	require.NoError(t, err)
	require.Len(t, items, 2)
	require.Equal(t, tokens.FixtureAvitoItemID1, items[0].AvitoID)
	require.Equal(t, tokens.FixtureAvitoItemTitle1, items[0].Title)
	require.Equal(t, tokens.FixtureAvitoItemID2, items[1].AvitoID)
}

// CONTRACT-AVITO-HEAL-004: heal rotates change-IP until content (capped by token).
func TestContract_AvitoHeal_RotateUntilClean(t *testing.T) {
	require.Equal(t, 18, tokens.AvitoHealMaxRotates)
	require.Equal(t, 8*time.Second, tokens.AvitoRateLimitBackoff)

	var hits atomic.Int32
	var changes atomic.Int32
	mock := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/change-ip" {
			changes.Add(1)
			w.WriteHeader(http.StatusOK)
			return
		}
		n := hits.Add(1)
		if n < 3 {
			w.WriteHeader(http.StatusTooManyRequests)
			_, _ = w.Write([]byte(tokens.FixtureAvitoIPBlockHTML))
			return
		}
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		_, _ = w.Write([]byte(tokens.FixtureAvitoFilterSERPHTML))
	}))
	t.Cleanup(mock.Close)

	client := service.NewAvitoClient(
		service.WithAvitoHTTPBase(mock.URL),
		service.WithProxyChangeIPURL(mock.URL+"/change-ip"),
	)
	client.SetChangeIPFunc(func(ctx context.Context, changeURL string) error {
		req, err := http.NewRequestWithContext(ctx, http.MethodGet, changeURL, nil)
		require.NoError(t, err)
		res, err := http.DefaultClient.Do(req)
		require.NoError(t, err)
		res.Body.Close()
		return nil
	})

	client.SetHealBackoff(0)
	items, err := client.FindFromFilterURL(context.Background(), "", tokens.FixtureFilterListingURL, "")
	require.NoError(t, err)
	require.NotEmpty(t, items)
	require.GreaterOrEqual(t, int(changes.Load()), 1)
	require.GreaterOrEqual(t, int(hits.Load()), 3)
}

// CONTRACT-AVITO-HEAL-005: 431 is healable (change-IP), not opaque fetch failed without rotate.
func TestContract_AvitoHeal_HeaderTooLargeTriggersHeal(t *testing.T) {
	var changes atomic.Int32
	var hits atomic.Int32
	mock := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/change-ip" {
			changes.Add(1)
			w.WriteHeader(http.StatusOK)
			return
		}
		n := hits.Add(1)
		if n == 1 {
			w.WriteHeader(tokens.AvitoHTTPStatusHeaderTooLarge)
			_, _ = w.Write([]byte(tokens.FixtureAvitoHeaderTooLargeHTML))
			return
		}
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		_, _ = w.Write([]byte(tokens.FixtureAvitoFilterSERPHTML))
	}))
	t.Cleanup(mock.Close)

	client := service.NewAvitoClient(
		service.WithAvitoHTTPBase(mock.URL),
		service.WithProxyChangeIPURL(mock.URL+"/change-ip"),
	)
	client.SetChangeIPFunc(func(ctx context.Context, changeURL string) error {
		res, err := http.Get(changeURL)
		require.NoError(t, err)
		res.Body.Close()
		return nil
	})
	client.SetHealBackoff(0)

	items, err := client.FindFromFilterURL(context.Background(), "", tokens.FixtureFilterListingURL, "")
	require.NoError(t, err)
	require.NotEmpty(t, items)
	require.Equal(t, int32(1), changes.Load())
}

// CONTRACT-AVITO-HEAL-006: rate-limited / heal exhaustion is soft — watch stays ACTIVE;
// soft also clears consecutive_failures so antibot does not advance auto-pause streak.
func TestContract_AvitoHeal_SoftFailDoesNotAutoPause(t *testing.T) {
	ctx := context.Background()
	pool, q := queries(t)
	users := service.NewUsers(pool)
	user, err := users.CreateUser(ctx, tokens.ProductEmail("heal-soft"), repository.PlanTypeFREE)
	require.NoError(t, err)

	watches := service.NewFilterWatches(pool)
	watch, err := watches.Start(ctx, user.ID, tokens.FixtureFilterListingURL)
	require.NoError(t, err)

	_, err = pool.Exec(ctx, `UPDATE listing_filter_watches SET consecutive_failures = $2 WHERE id = $1`, watch.ID, tokens.WatchAutoPauseAfterFails-1)
	require.NoError(t, err)

	worker := service.NewListingSearchWorker(pool, &healSoftFailProcessor{})
	for i := 0; i < tokens.WatchAutoPauseAfterFails+2; i++ {
		_, err := pool.Exec(ctx, `UPDATE listing_filter_watches SET last_polled_at = now() - interval '1 hour' WHERE id = $1`, watch.ID)
		require.NoError(t, err)
		_, err = worker.ProcessWatchPolls(ctx)
		require.NoError(t, err)
	}
	got, err := q.GetFilterWatch(ctx, watch.ID)
	require.NoError(t, err)
	require.Equal(t, repository.ListingWatchStatusACTIVE, got.Status)
	require.Equal(t, repository.ListingWatchMetaStatusPENDING, got.MetaStatus)
	require.NotNil(t, got.LastError)
	require.Contains(t, *got.LastError, tokens.ErrMsgListingSearchRateLimited)
	require.Equal(t, int32(0), got.ConsecutiveFailures)
}

// CONTRACT-AVITO-HEAL-007: hard fetch failures still auto-pause after N.
func TestContract_AvitoHeal_HardFailStillAutoPauses(t *testing.T) {
	ctx := context.Background()
	pool, q := queries(t)
	users := service.NewUsers(pool)
	user, err := users.CreateUser(ctx, tokens.ProductEmail("heal-hard"), repository.PlanTypeFREE)
	require.NoError(t, err)

	watches := service.NewFilterWatches(pool)
	watch, err := watches.Start(ctx, user.ID, tokens.FixtureFilterListingURL)
	require.NoError(t, err)

	worker := service.NewListingSearchWorker(pool, &healHardFailProcessor{})
	for i := 0; i < tokens.WatchAutoPauseAfterFails; i++ {
		_, err := pool.Exec(ctx, `UPDATE listing_filter_watches SET last_polled_at = now() - interval '1 hour' WHERE id = $1`, watch.ID)
		require.NoError(t, err)
		_, err = worker.ProcessWatchPolls(ctx)
		require.NoError(t, err)
	}
	got, err := q.GetFilterWatch(ctx, watch.ID)
	require.NoError(t, err)
	require.Equal(t, repository.ListingWatchStatusPAUSED, got.Status)
	require.Equal(t, int32(tokens.WatchAutoPauseAfterFails), got.ConsecutiveFailures)
}

// CONTRACT-AVITO-HEAL-008: monitoring due interval is calm (not 5s hammer).
func TestContract_AvitoHeal_WatchDueInterval(t *testing.T) {
	require.Equal(t, 30*time.Second, tokens.ListingSearchWatchPollInterval)
	require.Equal(t, "interval '30 seconds'", tokens.ListingSearchWatchDueSQLInterval)
	require.Equal(t, "каждые 30 секунд", tokens.ListingSearchWatchPollHumanRU)
	require.Contains(t, tokens.AppCopyWatchStatusActive, tokens.ListingSearchWatchPollHumanRU)
	require.Equal(t, "30s", tokens.DefaultWorkerTick)
}

// CONTRACT-AVITO-HEAL-009: soft failures are typed AvitoFetchFailure (no string sniffing).
func TestContract_AvitoHeal_TypedSoftFailure(t *testing.T) {
	soft := service.NewAvitoFetchFailureForTest(tokens.AvitoFetchStateRateLimited, 429, true)
	require.True(t, soft.Soft())
	require.ErrorIs(t, soft, domain.ErrListingSearchRateLimited)
	require.Contains(t, soft.Error(), tokens.AvitoFetchStateRateLimited)

	hard := service.NewAvitoFetchFailureForTest(tokens.AvitoFetchStateEmpty, 200, false)
	require.False(t, hard.Soft())
	require.ErrorIs(t, hard, domain.ErrListingSearchAvitoFetch)
}

// CONTRACT-AVITO-HEAL-010: http mode = HTTP heal + Chrome; chrome mode = Chrome-only SERP.
func TestContract_AvitoHeal_HTTPModeWiresChromeFallback(t *testing.T) {
	pool, _ := queries(t)

	httpProc, err := service.NewListingProcessor(pool, tokens.ListingSearchProcessorRod, tokens.FixtureAvitoHTTPBaseMock, tokens.FixtureBindingUserDataDir, tokens.RodFetchModeHTTP)
	require.NoError(t, err)
	httpRod := httpProc.(*service.RodAvitoListingProcessor)
	require.True(t, httpRod.HasHTTPHeal())
	_, ok := httpRod.PageFetcher().(*service.RodPageFetcher)
	require.True(t, ok)

	chromeProc, err := service.NewListingProcessor(pool, tokens.ListingSearchProcessorRod, tokens.AvitoHTTPSBase, tokens.FixtureBindingUserDataDir, tokens.RodFetchModeChrome)
	require.NoError(t, err)
	chromeRod := chromeProc.(*service.RodAvitoListingProcessor)
	require.False(t, chromeRod.HasHTTPHeal())
	_, ok = chromeRod.PageFetcher().(*service.RodPageFetcher)
	require.True(t, ok)
}

// CONTRACT-AVITO-HEAL-011: proxy probe must not treat antibot 429 as healthy.
func TestContract_AvitoHeal_ProxyProbeRejectsAntibot(t *testing.T) {
	require.Equal(t, tokens.HTTPStatus2xxMax, tokens.ProxyProbeMaxStatusCode)
	require.True(t, tokens.HTTPStatusOK(http.StatusOK))
	require.False(t, tokens.HTTPStatusOK(http.StatusTooManyRequests))
	mock := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusTooManyRequests)
		_, _ = w.Write([]byte(tokens.FixtureAvitoIPBlockHTML))
	}))
	t.Cleanup(mock.Close)
	err := service.DefaultHTTPProxyProbe(context.Background(), tokens.ProxyEndpointDirect, mock.URL)
	require.Error(t, err)
}

// CONTRACT-AVITO-HEAL-012: after HTTP heal exhaust, Chrome/HTML fallback still recovers SERP.
// WithAvitoHealMaxRotates(0) = exactly one HTTP GET (no rotate), then fallback.
func TestContract_AvitoHeal_ChromeFallbackAfterHTTPExhaust(t *testing.T) {
	ctx := context.Background()
	pool, _ := queries(t)

	// Direct egress so AvitoClient hits the mock (not a dead SOCKS fixture).
	acc, err := service.NewAvitoAccounts(pool).CreateWithSecret(ctx, "heal-fb", repository.AvitoAccountStatusACTIVE, "l", "s")
	require.NoError(t, err)
	proxy, err := service.NewProxies(pool).Create(ctx, tokens.ProxyEndpointDirect, repository.ProxyStatusACTIVE, "heal-fb-p")
	require.NoError(t, err)
	bindings := service.NewBindings(pool)
	b, err := bindings.Create(ctx, acc.ID, proxy.ID, tokens.FixtureBindingUserDataDir)
	require.NoError(t, err)
	_, err = bindings.WarmSession(ctx, b.ID, &contractPageFetcher{html: tokens.FixtureWarmSessionHTML})
	require.NoError(t, err)

	var httpHits atomic.Int32
	mock := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		httpHits.Add(1)
		w.WriteHeader(http.StatusTooManyRequests)
		_, _ = w.Write([]byte(tokens.FixtureAvitoIPBlockHTML))
	}))
	t.Cleanup(mock.Close)

	client := service.NewAvitoClient(
		service.WithAvitoHTTPBase(mock.URL),
		service.WithAvitoHealMaxRotates(0),
	)
	client.SetHealBackoff(0)

	var htmlHits atomic.Int32
	proc := service.NewRodAvitoListingProcessorWithClient(pool, &healFallbackFetcher{hits: &htmlHits, html: tokens.FixtureAvitoFilterSERPHTML}, client)

	items, err := proc.FindSimilar(ctx, tokens.FixtureFilterListingURL)
	require.NoError(t, err)
	require.NotEmpty(t, items)
	require.Equal(t, int32(1), httpHits.Load())
	require.GreaterOrEqual(t, int(htmlHits.Load()), 1)
}

// CONTRACT-AVITO-HEAL-013: face humanizeWatchError uses token ErrMsg* (no hard-coded English).
func TestContract_AvitoHeal_FaceErrorHumanizeUsesTokens(t *testing.T) {
	loggedIn := tokens.RenderAppFaceHTMLLoggedIn(tokens.ProductEmail("heal-face"), "", false)
	require.Contains(t, loggedIn, tokens.ErrMsgListingSearchRateLimited)
	require.Contains(t, loggedIn, tokens.ErrMsgListingSearchAvitoFetch)
	require.Contains(t, loggedIn, tokens.ErrMsgListingSearchAvitoEmpty)
	require.Contains(t, loggedIn, tokens.ErrMsgListingSearchNoBinding)
	require.Contains(t, loggedIn, tokens.AppCopyWatchErrRateLimited)
	anon := tokens.RenderAppFaceHTML()
	require.NotContains(t, anon, "каждую минуту")
	require.Contains(t, anon, tokens.ListingSearchWatchPollHumanRU)
	require.Equal(t, "interval '30 seconds'", tokens.ListingSearchWatchDueSQLInterval)
}

type healFallbackFetcher struct {
	hits *atomic.Int32
	html string
}

func (f *healFallbackFetcher) FetchHTML(context.Context, string, string, string) (string, error) {
	f.hits.Add(1)
	return f.html, nil
}

type healSoftFailProcessor struct{}

func (healSoftFailProcessor) FindSimilar(context.Context, string) ([]service.SimilarListing, error) {
	return nil, domain.ErrListingSearchRateLimited
}

type healHardFailProcessor struct{}

func (healHardFailProcessor) FindSimilar(context.Context, string) ([]service.SimilarListing, error) {
	return nil, domain.ErrListingSearchAvitoFetch
}
