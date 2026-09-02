package contracts_test

import (
	"context"
	"testing"

	"github.com/stretchr/testify/require"

	"vitek/internal/repository"
	"vitek/internal/service"
	"vitek/internal/tokens"
)

// CONTRACT-LISTING-020: factory builds Rod Avito processor from tokens.
func TestContract_ListingSearch_RodProcessorFactory(t *testing.T) {
	pool, _ := queries(t)

	rodProc, err := service.NewListingProcessor(pool, tokens.ListingSearchProcessorRod, tokens.AvitoHTTPSBase)
	require.NoError(t, err)
	require.IsType(t, &service.RodAvitoListingProcessor{}, rodProc)

	cfgOK := []string{
		tokens.ListingSearchProcessorStub,
		tokens.ListingSearchProcessorAvito,
		tokens.ListingSearchProcessorRod,
	}
	require.Equal(t, cfgOK, tokens.ListingSearchProcessors)
}

// fakeRodPageFetcher is a test double for browser page fetch (no Chrome in CI).
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
	_, err = proxies.Create(ctx, "socks5://127.0.0.1:9", repository.ProxyStatusACTIVE, "rod-proxy")
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
	// baseline: first filter run seeds seen, zero task_items
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
