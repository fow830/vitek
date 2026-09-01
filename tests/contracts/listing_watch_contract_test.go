package contracts_test

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/require"

	"vitek/internal/repository"
	"vitek/internal/service"
	"vitek/internal/tokens"
	httpapi "vitek/internal/transport/httpapi"
)

// CONTRACT-LISTING-016: filter POST starts watch; worker baseline then poll adds new hits only.
func TestContract_ListingSearch_FilterWatchPoll(t *testing.T) {
	ctx := context.Background()
	pool, q := queries(t)
	mock := startAvitoMock(t)
	t.Cleanup(mock.Close)

	client := service.NewAvitoClient(service.WithAvitoHTTPBase(mock.URL))
	worker := service.NewListingSearchWorker(pool, service.NewAvitoListingProcessor(pool, client))
	watches := service.NewFilterWatches(pool)

	users := service.NewUsers(pool)
	user, err := users.CreateUser(ctx, tokens.ProductEmail("listing-watch"), repository.PlanTypeFREE)
	require.NoError(t, err)

	proxies := service.NewProxies(pool)
	_, err = proxies.Create(ctx, mock.URL, repository.ProxyStatusACTIVE, "mock-proxy")
	require.NoError(t, err)

	_, err = service.NewAvitoAccounts(pool).CreateWithSecret(ctx, "acc-watch", repository.AvitoAccountStatusACTIVE, "login@test", "secret")
	require.NoError(t, err)

	filterURL := tokens.FixtureFilterListingURL
	watch, err := watches.Start(ctx, user.ID, filterURL)
	require.NoError(t, err)

	n, err := worker.ProcessWatchPolls(ctx)
	require.NoError(t, err)
	require.Equal(t, 1, n)

	hits, err := q.ListWatchHits(ctx, watch.ID)
	require.NoError(t, err)
	require.Empty(t, hits)

	var seenCount int
	err = pool.QueryRow(ctx, `SELECT count(*) FROM listing_filter_seen WHERE user_id = $1`, user.ID).Scan(&seenCount)
	require.NoError(t, err)
	require.Equal(t, 2, seenCount)

	n, err = worker.ProcessWatchPolls(ctx)
	require.NoError(t, err)
	require.Equal(t, 0, n)
}

// CONTRACT-LISTING-017: POST filter URL via /v1/me/tasks returns watch kind.
func TestContract_ListingSearch_FilterWatchHTTP(t *testing.T) {
	ctx := context.Background()
	pool, _ := queries(t)
	mailer := service.NewMemoryMagicLinkMailer()
	handler := httpapi.NewServer(pool, httpapi.WithMagicLinkMailer(mailer)).Handler()

	email := tokens.ProductEmail("listing-watch-http")
	_, err := service.NewUsers(pool).CreateUser(ctx, email, repository.PlanTypeFREE)
	require.NoError(t, err)
	cookie := loginCookie(t, handler, mailer, email)

	body, err := json.Marshal(map[string]any{tokens.JSONFieldQuery: tokens.FixtureFilterListingURL})
	require.NoError(t, err)
	req := withAppHost(httptest.NewRequest(http.MethodPost, tokens.PathV1MeTasks, bytes.NewReader(body)))
	req.Header.Set(tokens.HeaderContentType, tokens.MIMEApplicationJSON)
	req.Header.Set(tokens.HeaderCookie, tokens.CookieSessionName+"="+cookie)

	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)
	require.Equal(t, http.StatusCreated, rec.Code)

	var out map[string]any
	require.NoError(t, json.NewDecoder(rec.Body).Decode(&out))
	require.Equal(t, tokens.ListingSearchKindWatch, out[tokens.JSONFieldKind])
	require.Equal(t, tokens.ListingWatchStatusActive, out[tokens.JSONFieldStatus])
}
