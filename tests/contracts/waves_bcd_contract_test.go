package contracts_test

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/require"

	"vitek/internal/repository"
	"vitek/internal/service"
	"vitek/internal/tokens"
	httpapi "vitek/internal/transport/httpapi"
)

// CONTRACT-PROXY-010/011: proxy health schema + ListActive excludes DEAD.
func TestContract_ProxyHealthAndFetchable(t *testing.T) {
	ctx := context.Background()
	pool, q := queries(t)

	rows, err := pool.Query(ctx, `
		SELECT e.enumlabel FROM pg_enum e
		JOIN pg_type t ON e.enumtypid = t.oid
		WHERE t.typname = 'proxy_health_status'
		ORDER BY e.enumsortorder`)
	require.NoError(t, err)
	defer rows.Close()
	var vals []string
	for rows.Next() {
		var v string
		require.NoError(t, rows.Scan(&v))
		vals = append(vals, v)
	}
	require.Equal(t, tokens.SchemaProxyHealthStatuses, vals)

	alive, err := service.NewProxies(pool).Create(ctx, "socks5://127.0.0.1:1", repository.ProxyStatusACTIVE, "alive")
	require.NoError(t, err)
	dead, err := service.NewProxies(pool).Create(ctx, "socks5://127.0.0.1:2", repository.ProxyStatusACTIVE, "dead")
	require.NoError(t, err)
	_, err = pool.Exec(ctx, `UPDATE proxies SET health='DEAD' WHERE id=$1`, dead.ID)
	require.NoError(t, err)

	list, err := q.ListActiveProxies(ctx)
	require.NoError(t, err)
	require.Len(t, list, 1)
	require.Equal(t, alive.ID, list[0].ID)
	require.Equal(t, tokens.ProxyDeadAfterFails, 3)
	require.Equal(t, tokens.ProxyPoolMinActive, 1)
}

// CONTRACT-PROXY-020/021: bindings schema + create/list/session ready.
func TestContract_ListingFetchBindings(t *testing.T) {
	ctx := context.Background()
	pool, _ := queries(t)

	acc, err := service.NewAvitoAccounts(pool).CreateWithSecret(ctx, "bind-acc", repository.AvitoAccountStatusACTIVE, "ref", "pw")
	require.NoError(t, err)
	proxy, err := service.NewProxies(pool).Create(ctx, "socks5://127.0.0.1:3", repository.ProxyStatusACTIVE, "bind-p")
	require.NoError(t, err)

	bindings := service.NewBindings(pool)
	b, err := bindings.Create(ctx, acc.ID, proxy.ID, tokens.FixtureBindingUserDataDir)
	require.NoError(t, err)
	require.Equal(t, repository.ListingSessionStatusLOGGEDOUT, b.SessionStatus)

	ready, err := bindings.WarmSession(ctx, b.ID, &contractPageFetcher{})
	require.NoError(t, err)
	require.Equal(t, repository.ListingSessionStatusREADY, ready.SessionStatus)

	picked, err := bindings.PickReady(ctx)
	require.NoError(t, err)
	require.Equal(t, b.ID, picked.ID)
	require.Equal(t, proxy.Endpoint, picked.ProxyEndpoint)
}

// CONTRACT-ROD-PROD-001: worker Dockerfile stage uses chrome runtime token.
func TestContract_RodProdDockerfile(t *testing.T) {
	body := tokens.RenderDockerfile()
	require.Contains(t, body, "AS worker")
	require.Contains(t, body, tokens.ImageWorkerRuntime)
	require.Contains(t, body, "AS api")
	require.Contains(t, body, tokens.ImageRuntime)
	require.Equal(t, tokens.ListingSearchSERPMaxItems, 50)
	require.Equal(t, tokens.ListingSearchSERPMaxPages, 5)
	require.Equal(t, tokens.AvitoSimilarSearchLimit, 50)
}

// CONTRACT-LISTING-033/035: fetch URL keeps f=; fidelity drops non-apple titles.
func TestContract_FilterFidelityAndDepthTokens(t *testing.T) {
	require.Contains(t, tokens.NormalizeListingSearchURL(tokens.FixtureFilterListingURL), tokens.AvitoQueryFilterF+"=")
	require.Equal(t, tokens.ListingFilterAppleLabel, "apple")
	require.Contains(t, tokens.ListingFilterAppleDenySubstrings, "vivo")
	require.True(t, tokens.ListingTitleAllowedForFilter(tokens.FixtureFilterListingURL, tokens.FixtureAvitoItemTitle1))
	require.False(t, tokens.ListingTitleAllowedForFilter(tokens.FixtureFilterListingURL, tokens.FixtureNonAppleTitle))
	require.False(t, tokens.ListingTitleAllowedForFilter(tokens.FixtureFilterListingURL, "vivo phone"))
}

// CONTRACT-LISTING-060/062: watch meta PENDING → READY after successful poll.
func TestContract_WatchMetaReadyAfterPoll(t *testing.T) {
	ctx := context.Background()
	pool, q := queries(t)
	mock := startAvitoMock(t)
	t.Cleanup(mock.Close)

	client := service.NewAvitoClient(service.WithAvitoHTTPBase(mock.URL))
	worker := service.NewListingSearchWorker(pool, service.NewAvitoListingProcessor(pool, client))
	watches := service.NewFilterWatches(pool)

	user, err := service.NewUsers(pool).CreateUser(ctx, tokens.ProductEmail("watch-meta"), repository.PlanTypePRO)
	require.NoError(t, err)
	_, err = service.NewProxies(pool).Create(ctx, mock.URL, repository.ProxyStatusACTIVE, "p")
	require.NoError(t, err)
	_, err = service.NewAvitoAccounts(pool).CreateWithSecret(ctx, "a", repository.AvitoAccountStatusACTIVE, "l", "s")
	require.NoError(t, err)

	watch, err := watches.Start(ctx, user.ID, tokens.FixtureFilterListingURL)
	require.NoError(t, err)
	require.Equal(t, repository.ListingWatchMetaStatusPENDING, watch.MetaStatus)

	_, err = worker.ProcessWatchPolls(ctx)
	require.NoError(t, err)

	got, err := q.GetFilterWatch(ctx, watch.ID)
	require.NoError(t, err)
	require.Equal(t, repository.ListingWatchMetaStatusREADY, got.MetaStatus)
	require.NotEmpty(t, got.MetaJson)
	out := service.WatchJSON(got)
	require.Equal(t, tokens.WatchMetaStatusReady, out[tokens.JSONFieldMetaStatus])
}

// CONTRACT-DAY0-002b + NOTIFY-001/002/003: outbox enqueue + stub sender.
func TestContract_NotifyOutboxAndAllowlist(t *testing.T) {
	require.Contains(t, tokens.AllowedNotifyGoModFragments, tokens.GoModRedisClient)
	require.Contains(t, tokens.AllowedNotifyGoModFragments, tokens.GoModTelegramBotAPI)

	ctx := context.Background()
	pool, q := queries(t)

	sender := &service.LogNotificationSender{}
	notify := service.NewNotifications(pool, sender)

	user, err := service.NewUsers(pool).CreateUser(ctx, tokens.ProductEmail("notify-user"), repository.PlanTypePRO)
	require.NoError(t, err)
	_, err = notify.UpsertSettings(ctx, user.ID, "12345", true)
	require.NoError(t, err)

	watches := service.NewFilterWatches(pool)
	watch, err := watches.Start(ctx, user.ID, tokens.FixtureFilterListingURL)
	require.NoError(t, err)

	item, err := q.UpsertItem(ctx, repository.UpsertItemParams{
		AvitoID: tokens.FixtureAvitoItemID1,
		Title:   tokens.FixtureAvitoItemTitle1,
	})
	require.NoError(t, err)
	require.NoError(t, notify.EnqueueWatchHit(ctx, q, user.ID, watch.ID, item.ID))

	n, err := notify.ProcessOutbox(ctx, tokens.ListingSearchWatchDueLimit)
	require.NoError(t, err)
	require.Equal(t, 1, n)
	require.Len(t, sender.Sent, 1)
	require.Contains(t, sender.Sent[0], "12345")
}

// CONTRACT-ADMIN-UI-010: live poll interval token in admin face script.
func TestContract_AdminLivePollToken(t *testing.T) {
	body := tokens.RenderAppFaceHTMLLoggedIn(tokens.FixtureSessionEmail(), "", true)
	require.Contains(t, body, "setInterval")
	require.Contains(t, body, tokens.JSONFieldHealth)
	require.Equal(t, tokens.AdminLivePollIntervalMs, 5000)
	userBody := tokens.RenderAppFaceHTML()
	require.Contains(t, userBody, tokens.AppClassMetaPending)
}

// CONTRACT-HTTP bindings + notifications routes.
func TestContract_BindingsAndNotificationsHTTP(t *testing.T) {
	ctx := context.Background()
	pool, _ := queries(t)
	mailer := service.NewMemoryMagicLinkMailer()
	handler := httpapi.NewServer(pool, httpapi.WithMagicLinkMailer(mailer), httpapi.WithPageFetcher(&contractPageFetcher{})).Handler()

	adminUser, err := service.NewUsers(pool).CreateUser(ctx, tokens.ProductEmail("bind-admin2"), repository.PlanTypePRO)
	require.NoError(t, err)
	_, err = pool.Exec(ctx, `UPDATE users SET role='ADMIN' WHERE id=$1`, adminUser.ID)
	require.NoError(t, err)
	cookie := loginCookie(t, handler, mailer, tokens.ProductEmail("bind-admin2"))

	acc, err := service.NewAvitoAccounts(pool).CreateWithSecret(ctx, "http-acc", repository.AvitoAccountStatusACTIVE, "r", "p")
	require.NoError(t, err)
	proxy, err := service.NewProxies(pool).Create(ctx, "socks5://127.0.0.1:9", repository.ProxyStatusACTIVE, "hp")
	require.NoError(t, err)

	body, err := json.Marshal(map[string]any{
		tokens.JSONFieldAccountID:   service.UUIDString(acc.ID),
		tokens.JSONFieldProxyID:     service.UUIDString(proxy.ID),
		tokens.JSONFieldUserDataDir: tokens.FixtureBindingUserDataDir,
	})
	require.NoError(t, err)
	req := withAppHost(httptest.NewRequest(http.MethodPost, tokens.PathV1AdminBindings, bytes.NewReader(body)))
	req.Header.Set(tokens.HeaderContentType, tokens.MIMEApplicationJSON)
	req.Header.Set(tokens.HeaderCookie, tokens.CookieSessionName+"="+cookie)
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)
	require.Equal(t, http.StatusCreated, rec.Code)

	var created map[string]any
	require.NoError(t, json.NewDecoder(rec.Body).Decode(&created))
	id := created[tokens.JSONFieldID].(string)

	lreq := withAppHost(httptest.NewRequest(http.MethodPost, tokens.HTTPResourceID(tokens.PathV1AdminBindings, id)+tokens.PathV1AdminBindingSessionLoginSuffix, nil))
	lreq.Header.Set(tokens.HeaderCookie, tokens.CookieSessionName+"="+cookie)
	lrec := httptest.NewRecorder()
	handler.ServeHTTP(lrec, lreq)
	require.Equal(t, http.StatusOK, lrec.Code)

	nbody, err := json.Marshal(map[string]any{
		tokens.JSONFieldTelegramChatID: "999",
		tokens.JSONFieldEnabled:        true,
	})
	require.NoError(t, err)
	nreq := withAppHost(httptest.NewRequest(http.MethodPut, tokens.PathV1MeNotifications, bytes.NewReader(nbody)))
	nreq.Header.Set(tokens.HeaderContentType, tokens.MIMEApplicationJSON)
	nreq.Header.Set(tokens.HeaderCookie, tokens.CookieSessionName+"="+cookie)
	nrec := httptest.NewRecorder()
	handler.ServeHTTP(nrec, nreq)
	require.Equal(t, http.StatusOK, nrec.Code)

	sqlcBody, err := os.ReadFile(filepath.Join(moduleRoot(t), tokens.PathSQLC))
	require.NoError(t, err)
	require.Contains(t, string(sqlcBody), tokens.PathMigrationNotifications)
}
