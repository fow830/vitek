package contracts_test

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"

	"vitek/internal/domain"
	"vitek/internal/repository"
	"vitek/internal/service"
	"vitek/internal/tokens"
	httpapi "vitek/internal/transport/httpapi"
)

type failThenOKProcessor struct {
	failSubstr string
	okItems    []service.SimilarListing
}

func (p *failThenOKProcessor) FindSimilar(_ context.Context, listingURL string) ([]service.SimilarListing, error) {
	if strings.Contains(listingURL, p.failSubstr) {
		return nil, domain.ErrListingSearchAvitoFetch
	}
	return p.okItems, nil
}

type alwaysFailProcessor struct{}

func (alwaysFailProcessor) FindSimilar(context.Context, string) ([]service.SimilarListing, error) {
	return nil, domain.ErrListingSearchAvitoFetch
}

// CONTRACT-LISTING-040: PATCH pause removes watch from due set.
func TestContract_ListingWatch_PauseNotDue(t *testing.T) {
	ctx := context.Background()
	pool, q := queries(t)
	mock := startAvitoMock(t)
	t.Cleanup(mock.Close)

	client := service.NewAvitoClient(service.WithAvitoHTTPBase(mock.URL))
	worker := service.NewListingSearchWorker(pool, service.NewAvitoListingProcessor(pool, client))
	watches := service.NewFilterWatches(pool)

	user, err := service.NewUsers(pool).CreateUser(ctx, tokens.ProductEmail("watch-pause"), repository.PlanTypePRO)
	require.NoError(t, err)
	_, err = service.NewProxies(pool).Create(ctx, mock.URL, repository.ProxyStatusACTIVE, "p")
	require.NoError(t, err)
	_, err = service.NewAvitoAccounts(pool).CreateWithSecret(ctx, "a", repository.AvitoAccountStatusACTIVE, "l", "s")
	require.NoError(t, err)

	watch, err := watches.Start(ctx, user.ID, tokens.FixtureFilterListingURL)
	require.NoError(t, err)

	watch, err = watches.UpdateStatus(ctx, user.ID, watch.ID, tokens.ListingWatchStatusPaused)
	require.NoError(t, err)
	require.Equal(t, repository.ListingWatchStatusPAUSED, watch.Status)

	n, err := worker.ProcessWatchPolls(ctx)
	require.NoError(t, err)
	require.Equal(t, 0, n)

	got, err := q.GetFilterWatch(ctx, watch.ID)
	require.NoError(t, err)
	require.False(t, got.LastPolledAt.Valid)
}

// CONTRACT-LISTING-041: resume ACTIVE keeps seen baseline.
func TestContract_ListingWatch_ResumeKeepsSeen(t *testing.T) {
	ctx := context.Background()
	pool, q := queries(t)
	mock := startAvitoMock(t)
	t.Cleanup(mock.Close)

	client := service.NewAvitoClient(service.WithAvitoHTTPBase(mock.URL))
	worker := service.NewListingSearchWorker(pool, service.NewAvitoListingProcessor(pool, client))
	watches := service.NewFilterWatches(pool)

	user, err := service.NewUsers(pool).CreateUser(ctx, tokens.ProductEmail("watch-resume"), repository.PlanTypePRO)
	require.NoError(t, err)
	_, err = service.NewProxies(pool).Create(ctx, mock.URL, repository.ProxyStatusACTIVE, "p")
	require.NoError(t, err)
	_, err = service.NewAvitoAccounts(pool).CreateWithSecret(ctx, "a", repository.AvitoAccountStatusACTIVE, "l", "s")
	require.NoError(t, err)

	watch, err := watches.Start(ctx, user.ID, tokens.FixtureFilterListingURL)
	require.NoError(t, err)
	_, err = worker.ProcessWatchPolls(ctx)
	require.NoError(t, err)

	var seenBefore int
	require.NoError(t, pool.QueryRow(ctx, `SELECT count(*) FROM listing_filter_seen WHERE user_id=$1`, user.ID).Scan(&seenBefore))
	require.Equal(t, 2, seenBefore)

	_, err = watches.UpdateStatus(ctx, user.ID, watch.ID, tokens.ListingWatchStatusPaused)
	require.NoError(t, err)
	_, err = watches.UpdateStatus(ctx, user.ID, watch.ID, tokens.ListingWatchStatusActive)
	require.NoError(t, err)

	var seenAfter int
	require.NoError(t, pool.QueryRow(ctx, `SELECT count(*) FROM listing_filter_seen WHERE user_id=$1`, user.ID).Scan(&seenAfter))
	require.Equal(t, seenBefore, seenAfter)

	hits, err := q.ListWatchHits(ctx, watch.ID)
	require.NoError(t, err)
	require.Empty(t, hits)
}

// CONTRACT-LISTING-042: DELETE → DISABLED, hidden from default list.
func TestContract_ListingWatch_DeleteDisables(t *testing.T) {
	ctx := context.Background()
	pool, _ := queries(t)
	watches := service.NewFilterWatches(pool)

	user, err := service.NewUsers(pool).CreateUser(ctx, tokens.ProductEmail("watch-del"), repository.PlanTypePRO)
	require.NoError(t, err)
	watch, err := watches.Start(ctx, user.ID, tokens.FixtureFilterListingURL)
	require.NoError(t, err)

	require.NoError(t, watches.Disable(ctx, user.ID, watch.ID))
	list, err := watches.ListForUser(ctx, user.ID)
	require.NoError(t, err)
	require.Empty(t, list)
}

// CONTRACT-LISTING-043: re-POST same filter does not clear seen (amends LISTING-018).
func TestContract_ListingWatch_RePostKeepsSeen(t *testing.T) {
	ctx := context.Background()
	pool, q := queries(t)
	mock := startAvitoMock(t)
	t.Cleanup(mock.Close)

	client := service.NewAvitoClient(service.WithAvitoHTTPBase(mock.URL))
	worker := service.NewListingSearchWorker(pool, service.NewAvitoListingProcessor(pool, client))
	watches := service.NewFilterWatches(pool)

	user, err := service.NewUsers(pool).CreateUser(ctx, tokens.ProductEmail("watch-repost"), repository.PlanTypePRO)
	require.NoError(t, err)
	_, err = service.NewProxies(pool).Create(ctx, mock.URL, repository.ProxyStatusACTIVE, "p")
	require.NoError(t, err)
	_, err = service.NewAvitoAccounts(pool).CreateWithSecret(ctx, "a", repository.AvitoAccountStatusACTIVE, "l", "s")
	require.NoError(t, err)

	filterURL := tokens.FixtureFilterListingURL
	canonical := tokens.CanonicalListingSearchQuery(filterURL)
	watch, err := watches.Start(ctx, user.ID, filterURL)
	require.NoError(t, err)
	require.Contains(t, watch.Query, tokens.AvitoQueryFilterF+"=")

	require.NoError(t, q.InsertFilterSeen(ctx, repository.InsertFilterSeenParams{
		UserID: user.ID, FilterKey: canonical, AvitoID: "keep-me",
	}))

	watch, err = watches.Start(ctx, user.ID, filterURL)
	require.NoError(t, err)

	var seenCount int
	require.NoError(t, pool.QueryRow(ctx,
		`SELECT count(*) FROM listing_filter_seen WHERE user_id=$1 AND filter_key=$2 AND avito_id='keep-me'`,
		user.ID, canonical,
	).Scan(&seenCount))
	require.Equal(t, 1, seenCount)

	n, err := worker.ProcessWatchPolls(ctx)
	require.NoError(t, err)
	require.Equal(t, 1, n)

	hits, err := q.ListWatchHits(ctx, watch.ID)
	require.NoError(t, err)
	require.Len(t, hits, 2)
}

// CONTRACT-LISTING-044: reset-baseline clears seen+hits.
func TestContract_ListingWatch_ResetBaseline(t *testing.T) {
	ctx := context.Background()
	pool, q := queries(t)
	mock := startAvitoMock(t)
	t.Cleanup(mock.Close)

	client := service.NewAvitoClient(service.WithAvitoHTTPBase(mock.URL))
	worker := service.NewListingSearchWorker(pool, service.NewAvitoListingProcessor(pool, client))
	watches := service.NewFilterWatches(pool)

	user, err := service.NewUsers(pool).CreateUser(ctx, tokens.ProductEmail("watch-reset-bl"), repository.PlanTypePRO)
	require.NoError(t, err)
	_, err = service.NewProxies(pool).Create(ctx, mock.URL, repository.ProxyStatusACTIVE, "p")
	require.NoError(t, err)
	_, err = service.NewAvitoAccounts(pool).CreateWithSecret(ctx, "a", repository.AvitoAccountStatusACTIVE, "l", "s")
	require.NoError(t, err)

	watch, err := watches.Start(ctx, user.ID, tokens.FixtureFilterListingURL)
	require.NoError(t, err)
	_, err = worker.ProcessWatchPolls(ctx)
	require.NoError(t, err)

	require.NoError(t, watches.ResetBaseline(ctx, user.ID, watch.ID))

	var seenCount int
	require.NoError(t, pool.QueryRow(ctx, `SELECT count(*) FROM listing_filter_seen WHERE user_id=$1`, user.ID).Scan(&seenCount))
	require.Equal(t, 0, seenCount)
	hits, err := q.ListWatchHits(ctx, watch.ID)
	require.NoError(t, err)
	require.Empty(t, hits)
}

// CONTRACT-LISTING-050: poll error on one watch does not abort siblings.
func TestContract_ListingWatch_PollErrorIsolated(t *testing.T) {
	ctx := context.Background()
	pool, q := queries(t)

	proc := &failThenOKProcessor{
		failSubstr: "apple",
		okItems: []service.SimilarListing{
			{AvitoID: tokens.FixtureAvitoItemID1, Title: tokens.FixtureAvitoItemTitle1},
			{AvitoID: tokens.FixtureAvitoItemID2, Title: tokens.FixtureAvitoItemTitle2},
		},
	}
	worker := service.NewListingSearchWorker(pool, proc)
	watches := service.NewFilterWatches(pool)

	user, err := service.NewUsers(pool).CreateUser(ctx, tokens.ProductEmail("watch-iso"), repository.PlanTypePRO)
	require.NoError(t, err)

	bad, err := watches.Start(ctx, user.ID, tokens.FixtureFilterListingURL)
	require.NoError(t, err)
	good, err := watches.Start(ctx, user.ID, tokens.FixtureFilterListingURL2)
	require.NoError(t, err)

	n, err := worker.ProcessWatchPolls(ctx)
	require.NoError(t, err)
	require.Equal(t, 2, n)

	badGot, err := q.GetFilterWatch(ctx, bad.ID)
	require.NoError(t, err)
	require.NotNil(t, badGot.LastError)
	require.Equal(t, tokens.ErrMsgListingSearchAvitoFetch, *badGot.LastError)
	require.Equal(t, int32(1), badGot.ConsecutiveFailures)

	goodGot, err := q.GetFilterWatch(ctx, good.ID)
	require.NoError(t, err)
	require.Nil(t, goodGot.LastError)
	require.True(t, goodGot.LastPolledAt.Valid)
}

// CONTRACT-LISTING-051: WatchJSON exposes last_error fields.
func TestContract_ListingWatch_JSONExposesError(t *testing.T) {
	ctx := context.Background()
	pool, _ := queries(t)
	watches := service.NewFilterWatches(pool)
	user, err := service.NewUsers(pool).CreateUser(ctx, tokens.ProductEmail("watch-json-err"), repository.PlanTypePRO)
	require.NoError(t, err)
	watch, err := watches.Start(ctx, user.ID, tokens.FixtureFilterListingURL)
	require.NoError(t, err)

	worker := service.NewListingSearchWorker(pool, alwaysFailProcessor{})
	_, err = worker.ProcessWatchPolls(ctx)
	require.NoError(t, err)

	got, err := watches.GetForUser(ctx, user.ID, watch.ID)
	require.NoError(t, err)
	out := service.WatchJSON(got)
	require.Equal(t, tokens.ErrMsgListingSearchAvitoFetch, out[tokens.JSONFieldLastError])
	require.EqualValues(t, 1, out[tokens.JSONFieldConsecutiveFailures])
}

// CONTRACT-LISTING-052: auto-pause after N consecutive failures.
func TestContract_ListingWatch_AutoPauseAfterFails(t *testing.T) {
	ctx := context.Background()
	pool, q := queries(t)
	watches := service.NewFilterWatches(pool)
	user, err := service.NewUsers(pool).CreateUser(ctx, tokens.ProductEmail("watch-autopause"), repository.PlanTypePRO)
	require.NoError(t, err)
	watch, err := watches.Start(ctx, user.ID, tokens.FixtureFilterListingURL)
	require.NoError(t, err)

	worker := service.NewListingSearchWorker(pool, alwaysFailProcessor{})
	for i := 0; i < tokens.WatchAutoPauseAfterFails; i++ {
		// Force due each time by clearing last_polled_at if set without clearing failures.
		_, err = pool.Exec(ctx, `UPDATE listing_filter_watches SET last_polled_at = NULL WHERE id=$1`, watch.ID)
		require.NoError(t, err)
		_, err = worker.ProcessWatchPolls(ctx)
		require.NoError(t, err)
	}

	got, err := q.GetFilterWatch(ctx, watch.ID)
	require.NoError(t, err)
	require.Equal(t, repository.ListingWatchStatusPAUSED, got.Status)
	require.Equal(t, int32(tokens.WatchAutoPauseAfterFails), got.ConsecutiveFailures)
}

// CONTRACT-LISTING-053: success clears last_error.
func TestContract_ListingWatch_SuccessClearsError(t *testing.T) {
	ctx := context.Background()
	pool, q := queries(t)
	mock := startAvitoMock(t)
	t.Cleanup(mock.Close)

	watches := service.NewFilterWatches(pool)
	user, err := service.NewUsers(pool).CreateUser(ctx, tokens.ProductEmail("watch-clear-err"), repository.PlanTypePRO)
	require.NoError(t, err)
	_, err = service.NewProxies(pool).Create(ctx, mock.URL, repository.ProxyStatusACTIVE, "p")
	require.NoError(t, err)
	_, err = service.NewAvitoAccounts(pool).CreateWithSecret(ctx, "a", repository.AvitoAccountStatusACTIVE, "l", "s")
	require.NoError(t, err)

	watch, err := watches.Start(ctx, user.ID, tokens.FixtureFilterListingURL)
	require.NoError(t, err)

	failWorker := service.NewListingSearchWorker(pool, alwaysFailProcessor{})
	_, err = failWorker.ProcessWatchPolls(ctx)
	require.NoError(t, err)

	_, err = pool.Exec(ctx, `UPDATE listing_filter_watches SET last_polled_at = NULL WHERE id=$1`, watch.ID)
	require.NoError(t, err)

	okWorker := service.NewListingSearchWorker(pool, service.NewAvitoListingProcessor(pool, service.NewAvitoClient(service.WithAvitoHTTPBase(mock.URL))))
	_, err = okWorker.ProcessWatchPolls(ctx)
	require.NoError(t, err)

	got, err := q.GetFilterWatch(ctx, watch.ID)
	require.NoError(t, err)
	require.Nil(t, got.LastError)
	require.Equal(t, int32(0), got.ConsecutiveFailures)
	require.True(t, got.LastSuccessAt.Valid)
}

// CONTRACT-LIMITS-010: FREE plan cannot exceed max_watches.
func TestContract_WatchQuota_FREE(t *testing.T) {
	ctx := context.Background()
	pool, _ := queries(t)
	watches := service.NewFilterWatches(pool)
	user, err := service.NewUsers(pool).CreateUser(ctx, tokens.ProductEmail("watch-quota-free"), repository.PlanTypeFREE)
	require.NoError(t, err)

	_, err = watches.Start(ctx, user.ID, tokens.FixtureFilterListingURL)
	require.NoError(t, err)
	_, err = watches.Start(ctx, user.ID, tokens.FixtureFilterListingURL2)
	require.ErrorIs(t, err, domain.ErrWatchLimitExceeded)
}

// CONTRACT-LIMITS-011/012: DISABLED not counted; HTTP patch/delete/reset on allowlist.
func TestContract_WatchQuota_DisabledNotCounted_AndHTTPLifecycle(t *testing.T) {
	ctx := context.Background()
	pool, _ := queries(t)
	mailer := service.NewMemoryMagicLinkMailer()
	handler := httpapi.NewServer(pool, httpapi.WithMagicLinkMailer(mailer)).Handler()

	email := tokens.ProductEmail("watch-quota-http")
	user, err := service.NewUsers(pool).CreateUser(ctx, email, repository.PlanTypeFREE)
	require.NoError(t, err)
	cookie := loginCookie(t, handler, mailer, email)
	watches := service.NewFilterWatches(pool)

	watch, err := watches.Start(ctx, user.ID, tokens.FixtureFilterListingURL)
	require.NoError(t, err)
	require.NoError(t, watches.Disable(ctx, user.ID, watch.ID))

	watch2, err := watches.Start(ctx, user.ID, tokens.FixtureFilterListingURL2)
	require.NoError(t, err)
	id := service.UUIDString(watch2.ID)

	patchBody, err := json.Marshal(map[string]any{tokens.JSONFieldStatus: tokens.ListingWatchStatusPaused})
	require.NoError(t, err)
	preq := withAppHost(httptest.NewRequest(http.MethodPatch, tokens.HTTPResourceID(tokens.PathV1MeWatches, id), bytes.NewReader(patchBody)))
	preq.Header.Set(tokens.HeaderContentType, tokens.MIMEApplicationJSON)
	preq.Header.Set(tokens.HeaderCookie, tokens.CookieSessionName+"="+cookie)
	prec := httptest.NewRecorder()
	handler.ServeHTTP(prec, preq)
	require.Equal(t, http.StatusOK, prec.Code)

	rreq := withAppHost(httptest.NewRequest(http.MethodPost, tokens.HTTPResourceID(tokens.PathV1MeWatches, id)+tokens.PathV1WatchResetBaselineSuffix, nil))
	rreq.Header.Set(tokens.HeaderCookie, tokens.CookieSessionName+"="+cookie)
	rrec := httptest.NewRecorder()
	handler.ServeHTTP(rrec, rreq)
	require.Equal(t, http.StatusOK, rrec.Code)

	dreq := withAppHost(httptest.NewRequest(http.MethodDelete, tokens.HTTPResourceID(tokens.PathV1MeWatches, id), nil))
	dreq.Header.Set(tokens.HeaderCookie, tokens.CookieSessionName+"="+cookie)
	drec := httptest.NewRecorder()
	handler.ServeHTTP(drec, dreq)
	require.Equal(t, http.StatusOK, drec.Code)

	require.Contains(t, tokens.HTTPPathAllowlist, tokens.HTTPPathID(tokens.PathV1MeWatches))
	require.Contains(t, tokens.HTTPPathAllowlist, tokens.HTTPPathMeWatchResetBaseline())
	require.Equal(t, tokens.PlanMaxWatchesFREE, int32(1))
	require.Equal(t, tokens.WatchAutoPauseAfterFails, 3)
}
