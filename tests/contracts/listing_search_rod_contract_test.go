package contracts_test

import (
	"context"
	"os"
	"path/filepath"
	"strconv"
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	"vitek/internal/domain"
	"vitek/internal/repository"
	"vitek/internal/service"
	"vitek/internal/tokens"
)

// CONTRACT-LISTING-020: factory builds Rod Avito processor from tokens.
func TestContract_ListingSearch_RodProcessorFactory(t *testing.T) {
	pool, _ := queries(t)

	rodProc, err := service.NewListingProcessor(pool, tokens.ListingSearchProcessorRod, tokens.AvitoHTTPSBase, "")
	require.NoError(t, err)
	require.IsType(t, &service.RodAvitoListingProcessor{}, rodProc)

	require.Equal(t, []string{
		tokens.ListingSearchProcessorStub,
		tokens.ListingSearchProcessorAvito,
		tokens.ListingSearchProcessorRod,
	}, tokens.ListingSearchProcessors)
}

type fakeRodPageFetcher struct {
	html string
	err  error
}

func (f *fakeRodPageFetcher) FetchHTML(_ context.Context, _, _ string) (string, error) {
	if f.err != nil {
		return "", f.err
	}
	return f.html, nil
}

// CONTRACT-LISTING-021: Rod processor completes filter task from browser HTML (SERP items).
func TestContract_ListingSearch_RodProcessorFilterFromHTML(t *testing.T) {
	ctx := context.Background()
	pool, q := queries(t)

	fetcher := &fakeRodPageFetcher{html: tokens.FixtureAvitoFilterSERPHTML}
	proc := service.NewRodAvitoListingProcessor(pool, fetcher)
	worker := service.NewListingSearchWorker(pool, proc)

	users := service.NewUsers(pool)
	user, err := users.CreateUser(ctx, tokens.ProductEmail("listing-rod"), repository.PlanTypeFREE)
	require.NoError(t, err)

	proxies := service.NewProxies(pool)
	_, err = proxies.Create(ctx, tokens.FixtureRodProxyEndpoint, repository.ProxyStatusACTIVE, "rod-proxy")
	require.NoError(t, err)

	_, err = service.NewAvitoAccounts(pool).CreateWithSecret(ctx, "acc-rod", repository.AvitoAccountStatusACTIVE, "login@test", "secret")
	require.NoError(t, err)

	task, err := service.NewTasks(pool).CreateTask(ctx, user.ID, tokens.FixtureFilterListingURL)
	require.NoError(t, err)

	ok, err := worker.ProcessOne(ctx)
	require.NoError(t, err)
	require.True(t, ok)

	got, err := q.GetTask(ctx, task.ID)
	require.NoError(t, err)
	require.Equal(t, repository.TaskStatusCOMPLETED, got.Status)

	items, err := q.ListTaskItems(ctx, task.ID)
	require.NoError(t, err)
	require.Empty(t, items)

	var seenCount int
	err = pool.QueryRow(ctx, `SELECT count(*) FROM listing_filter_seen WHERE user_id = $1`, user.ID).Scan(&seenCount)
	require.NoError(t, err)
	require.Equal(t, 2, seenCount)
}

// CONTRACT-LISTING-022: SERP HTML parser extracts avito ids + titles (SoT for Rod path).
func TestContract_ListingSearch_ParseAvitoSERPHTML(t *testing.T) {
	items, err := tokens.ParseAvitoSERPItems(tokens.FixtureAvitoFilterSERPHTML)
	require.NoError(t, err)
	require.Len(t, items, 2)
	require.Equal(t, tokens.FixtureAvitoItemID1, items[0].AvitoID)
	require.Equal(t, tokens.FixtureAvitoItemTitle1, items[0].Title)
	require.Equal(t, tokens.FixtureAvitoItemID2, items[1].AvitoID)
}

// CONTRACT-LISTING-023: Rod processor fails task when no active proxy.
func TestContract_ListingSearch_RodProcessorNoProxyFailsTask(t *testing.T) {
	ctx := context.Background()
	pool, q := queries(t)

	proc := service.NewRodAvitoListingProcessor(pool, &fakeRodPageFetcher{html: tokens.FixtureAvitoFilterSERPHTML})
	worker := service.NewListingSearchWorker(pool, proc)

	users := service.NewUsers(pool)
	user, err := users.CreateUser(ctx, tokens.ProductEmail("listing-rod-noproxy"), repository.PlanTypeFREE)
	require.NoError(t, err)
	_, err = service.NewAvitoAccounts(pool).CreateWithSecret(ctx, "acc-rod-2", repository.AvitoAccountStatusACTIVE, "login@test", "secret")
	require.NoError(t, err)

	task, err := service.NewTasks(pool).CreateTask(ctx, user.ID, tokens.FixtureFilterListingURL)
	require.NoError(t, err)

	ok, err := worker.ProcessOne(ctx)
	require.NoError(t, err)
	require.True(t, ok)

	got, err := q.GetTask(ctx, task.ID)
	require.NoError(t, err)
	require.Equal(t, repository.TaskStatusFAILED, got.Status)
}

// CONTRACT-LISTING-024: Rod processor fails task when no active account.
func TestContract_ListingSearch_RodProcessorNoAccountFailsTask(t *testing.T) {
	ctx := context.Background()
	pool, q := queries(t)

	proc := service.NewRodAvitoListingProcessor(pool, &fakeRodPageFetcher{html: tokens.FixtureAvitoFilterSERPHTML})
	worker := service.NewListingSearchWorker(pool, proc)

	users := service.NewUsers(pool)
	user, err := users.CreateUser(ctx, tokens.ProductEmail("listing-rod-noacc"), repository.PlanTypeFREE)
	require.NoError(t, err)
	_, err = service.NewProxies(pool).Create(ctx, tokens.FixtureRodProxyEndpoint, repository.ProxyStatusACTIVE, "rod-proxy")
	require.NoError(t, err)

	task, err := service.NewTasks(pool).CreateTask(ctx, user.ID, tokens.FixtureFilterListingURL)
	require.NoError(t, err)

	ok, err := worker.ProcessOne(ctx)
	require.NoError(t, err)
	require.True(t, ok)

	got, err := q.GetTask(ctx, task.ID)
	require.NoError(t, err)
	require.Equal(t, repository.TaskStatusFAILED, got.Status)
}

// CONTRACT-LISTING-025: Rod processor is filter-only (item URL → domain error).
func TestContract_ListingSearch_RodProcessorFilterOnly(t *testing.T) {
	ctx := context.Background()
	pool, _ := queries(t)

	proc := service.NewRodAvitoListingProcessor(pool, &fakeRodPageFetcher{html: tokens.FixtureAvitoFilterSERPHTML})
	_, err := service.NewProxies(pool).Create(ctx, tokens.FixtureRodProxyEndpoint, repository.ProxyStatusACTIVE, "rod-proxy")
	require.NoError(t, err)
	_, err = service.NewAvitoAccounts(pool).CreateWithSecret(ctx, "acc-rod-item", repository.AvitoAccountStatusACTIVE, "login@test", "secret")
	require.NoError(t, err)

	_, err = proc.FindSimilar(ctx, tokens.FixtureListingURL)
	require.ErrorIs(t, err, domain.ErrListingSearchRodFilterOnly)
}

// CONTRACT-LISTING-026: listing_watch_status SoT mirrors PostgreSQL enum.
func TestContract_ListingSearch_WatchStatusSchema(t *testing.T) {
	require.Equal(t, []string{
		tokens.ListingWatchStatusActive,
		tokens.ListingWatchStatusPaused,
		tokens.ListingWatchStatusDisabled,
	}, tokens.SchemaListingWatchStatuses)

	pool, _ := queries(t)
	var vals []string
	rows, err := pool.Query(context.Background(), `
		SELECT e.enumlabel
		FROM pg_type t
		JOIN pg_enum e ON t.oid = e.enumtypid
		WHERE t.typname = 'listing_watch_status'
		ORDER BY e.enumsortorder`)
	require.NoError(t, err)
	defer rows.Close()
	for rows.Next() {
		var v string
		require.NoError(t, rows.Scan(&v))
		vals = append(vals, v)
	}
	require.Equal(t, tokens.SchemaListingWatchStatuses, vals)
}

// CONTRACT-LISTING-027: Rod/Avito timing + SERP marker tokens stay aligned.
func TestContract_ListingSearch_RodAvitoTokens(t *testing.T) {
	require.Equal(t, 30*time.Second, tokens.AvitoHTTPClientTimeout)
	require.Equal(t, 4<<20, tokens.AvitoHTTPMaxBodyBytes)
	require.Equal(t, 45*time.Second, tokens.RodPageFetchTimeout)
	require.True(t, tokens.RodHeadlessDefault)
	require.Equal(t, time.Minute, tokens.ListingSearchWatchPollInterval)
	require.Equal(t, int(tokens.ListingSearchWatchPollInterval/time.Millisecond), tokens.ListingSearchWatchPollIntervalMs)
	require.Equal(t, tokens.ListingSearchTaskListLimit, tokens.ListingSearchWatchDueLimit)
	require.Equal(t, tokens.EnvRodUserDataDir, "ROD_USER_DATA_DIR")
	require.Contains(t, tokens.FixtureAvitoFilterSERPHTML, tokens.AvitoSERPAttrItemID+`="`+tokens.FixtureAvitoItemID1+`"`)
	require.Contains(t, tokens.FixtureAvitoFilterSERPHTML, `data-marker="`+tokens.AvitoSERPMarkerItemTitle+`"`)

	sqlPath := filepath.Join(moduleRoot(t), tokens.PathQueryListingWatches)
	body, err := os.ReadFile(sqlPath)
	require.NoError(t, err)
	require.Contains(t, string(body), tokens.ListingSearchWatchDueSQLInterval)
	require.Contains(t, string(body), "LIMIT "+strconv.FormatInt(int64(tokens.ListingSearchWatchDueLimit), 10))
}
