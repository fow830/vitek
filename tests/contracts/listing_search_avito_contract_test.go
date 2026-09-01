package contracts_test

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"

	"vitek/internal/repository"
	"vitek/internal/service"
	"vitek/internal/tokens"
)

func avitoFixtureURL(itemID string) string {
	return tokens.AvitoListingURL(tokens.AvitoListingHostPrimary, "moskva/telefony/item_"+itemID)
}

func startAvitoMock(t *testing.T) *httptest.Server {
	t.Helper()
	return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch {
		case strings.HasPrefix(r.URL.Path, tokens.AvitoWebPathItemDetails):
			itemID := strings.TrimPrefix(r.URL.Path, tokens.AvitoWebPathItemDetails)
			_ = json.NewEncoder(w).Encode(map[string]any{
				tokens.AvitoJSONFieldID:         itemID,
				tokens.AvitoJSONFieldTitle:      tokens.FixtureAvitoItemTitle1,
				tokens.AvitoJSONFieldCategoryID: tokens.FixtureAvitoCategoryID,
				tokens.AvitoJSONFieldLocationID: tokens.FixtureAvitoLocationID,
			})
		case r.URL.Path == tokens.AvitoWebPathItemSearch:
			if r.URL.Query().Get(tokens.AvitoQueryFilterF) != "" {
				_ = json.NewEncoder(w).Encode(map[string]any{
					tokens.AvitoJSONFieldItems: []map[string]any{
						{tokens.AvitoJSONFieldID: tokens.FixtureAvitoItemID1, tokens.AvitoJSONFieldTitle: tokens.FixtureAvitoItemTitle1},
						{tokens.AvitoJSONFieldID: tokens.FixtureAvitoItemID2, tokens.AvitoJSONFieldTitle: "Filter hit 2"},
					},
				})
				return
			}
			_ = json.NewEncoder(w).Encode(map[string]any{
				tokens.AvitoJSONFieldItems: []map[string]any{
					{tokens.AvitoJSONFieldID: tokens.FixtureListingID1, tokens.AvitoJSONFieldTitle: "skip source"},
					{tokens.AvitoJSONFieldID: tokens.FixtureAvitoItemID1, tokens.AvitoJSONFieldTitle: tokens.FixtureAvitoItemTitle1},
					{tokens.AvitoJSONFieldID: tokens.FixtureAvitoItemID2, tokens.AvitoJSONFieldTitle: "Similar 2"},
				},
			})
		default:
			if strings.Contains(r.URL.Path, "apple-") {
				w.Header().Set("Content-Type", "text/html; charset=utf-8")
				_, _ = w.Write([]byte(tokens.FixtureAvitoFilterPageHTML))
				return
			}
			http.NotFound(w, r)
		}
	}))
}

// CONTRACT-LISTING-006: Avito processor completes task with numeric avito_id results.
func TestContract_ListingSearch_AvitoProcessorCompletesTask(t *testing.T) {
	ctx := context.Background()
	pool, q := queries(t)
	mock := startAvitoMock(t)
	t.Cleanup(mock.Close)

	client := service.NewAvitoClient(service.WithAvitoHTTPBase(mock.URL))
	worker := service.NewListingSearchWorker(pool, service.NewAvitoListingProcessor(pool, client))

	users := service.NewUsers(pool)
	user, err := users.CreateUser(ctx, tokens.ProductEmail("listing-avito"), repository.PlanTypeFREE)
	require.NoError(t, err)

	proxies := service.NewProxies(pool)
	_, err = proxies.Create(ctx, mock.URL, repository.ProxyStatusACTIVE, "mock-proxy")
	require.NoError(t, err)

	_, err = service.NewAvitoAccounts(pool).CreateWithSecret(ctx, "acc-1", repository.AvitoAccountStatusACTIVE, "login@test", "secret")
	require.NoError(t, err)

	listingURL := avitoFixtureURL(tokens.FixtureListingID1)
	require.True(t, tokens.ValidListingURL(listingURL))

	task, err := service.NewTasks(pool).CreateTask(ctx, user.ID, listingURL)
	require.NoError(t, err)

	ok, err := worker.ProcessOne(ctx)
	require.NoError(t, err)
	require.True(t, ok)

	got, err := q.GetTask(ctx, task.ID)
	require.NoError(t, err)
	require.Equal(t, repository.TaskStatusCOMPLETED, got.Status)

	items, err := q.ListTaskItems(ctx, task.ID)
	require.NoError(t, err)
	require.Len(t, items, 2)
	require.Equal(t, tokens.FixtureAvitoItemID1, items[0].AvitoID)
	require.NotContains(t, items[0].AvitoID, tokens.ListingStubAvitoIDPrefix)
}

// CONTRACT-LISTING-007: Avito processor fails task when no active proxy.
func TestContract_ListingSearch_AvitoProcessorNoProxyFailsTask(t *testing.T) {
	ctx := context.Background()
	pool, q := queries(t)
	mock := startAvitoMock(t)
	t.Cleanup(mock.Close)

	client := service.NewAvitoClient(service.WithAvitoHTTPBase(mock.URL))
	worker := service.NewListingSearchWorker(pool, service.NewAvitoListingProcessor(pool, client))

	users := service.NewUsers(pool)
	user, err := users.CreateUser(ctx, tokens.ProductEmail("listing-noproxy"), repository.PlanTypeFREE)
	require.NoError(t, err)
	_, err = service.NewAvitoAccounts(pool).CreateWithSecret(ctx, "acc-2", repository.AvitoAccountStatusACTIVE, "login@test", "secret")
	require.NoError(t, err)

	task, err := service.NewTasks(pool).CreateTask(ctx, user.ID, avitoFixtureURL(tokens.FixtureListingID1))
	require.NoError(t, err)

	ok, err := worker.ProcessOne(ctx)
	require.NoError(t, err)
	require.True(t, ok)

	got, err := q.GetTask(ctx, task.ID)
	require.NoError(t, err)
	require.Equal(t, repository.TaskStatusFAILED, got.Status)
}

// CONTRACT-LISTING-008: Avito processor fails task when no active account secret.
func TestContract_ListingSearch_AvitoProcessorNoAccountFailsTask(t *testing.T) {
	ctx := context.Background()
	pool, q := queries(t)
	mock := startAvitoMock(t)
	t.Cleanup(mock.Close)

	client := service.NewAvitoClient(service.WithAvitoHTTPBase(mock.URL))
	worker := service.NewListingSearchWorker(pool, service.NewAvitoListingProcessor(pool, client))

	users := service.NewUsers(pool)
	user, err := users.CreateUser(ctx, tokens.ProductEmail("listing-noacc"), repository.PlanTypeFREE)
	require.NoError(t, err)

	_, err = service.NewProxies(pool).Create(ctx, mock.URL, repository.ProxyStatusACTIVE, "mock-proxy")
	require.NoError(t, err)

	task, err := service.NewTasks(pool).CreateTask(ctx, user.ID, avitoFixtureURL(tokens.FixtureListingID1))
	require.NoError(t, err)

	ok, err := worker.ProcessOne(ctx)
	require.NoError(t, err)
	require.True(t, ok)

	got, err := q.GetTask(ctx, task.ID)
	require.NoError(t, err)
	require.Equal(t, repository.TaskStatusFAILED, got.Status)
}

// CONTRACT-LISTING-009: avito_account_secrets schema surface.
func TestContract_ListingSearch_AvitoSecretsSchema(t *testing.T) {
	ctx := context.Background()
	pool, _ := queries(t)

	for _, table := range tokens.SchemaAvitoSecretsTables {
		var exists bool
		err := pool.QueryRow(ctx,
			`SELECT EXISTS (
				SELECT 1 FROM information_schema.tables
				WHERE table_schema = 'public' AND table_name = $1
			)`, table).Scan(&exists)
		require.NoError(t, err)
		require.Truef(t, exists, "table missing: %s", table)
	}

	for _, col := range []string{"account_id", "password", "updated_at"} {
		var exists bool
		err := pool.QueryRow(ctx,
			`SELECT EXISTS (
				SELECT 1 FROM information_schema.columns
				WHERE table_schema = 'public'
				  AND table_name = 'avito_account_secrets'
				  AND column_name = $1
			)`, col).Scan(&exists)
		require.NoError(t, err)
		require.Truef(t, exists, "avito_account_secrets.%s missing", col)
	}
}

// CONTRACT-LISTING-010: listing processor factory respects tokens.
func TestContract_ListingSearch_ProcessorFactory(t *testing.T) {
	pool, _ := queries(t)

	stub, err := service.NewListingProcessor(pool, tokens.ListingSearchProcessorStub, "")
	require.NoError(t, err)
	require.IsType(t, &service.StubListingProcessor{}, stub)

	avito, err := service.NewListingProcessor(pool, tokens.ListingSearchProcessorAvito, tokens.AvitoHTTPSBase)
	require.NoError(t, err)
	require.IsType(t, &service.AvitoListingProcessor{}, avito)

	_, err = service.NewListingProcessor(pool, tokens.FixtureInvalidEnum, "")
	require.Error(t, err)
}

// CONTRACT-LISTING-011: Avito processor completes task for filter/stream URL.
func TestContract_ListingSearch_AvitoProcessorFilterURL(t *testing.T) {
	ctx := context.Background()
	pool, q := queries(t)
	mock := startAvitoMock(t)
	t.Cleanup(mock.Close)

	client := service.NewAvitoClient(service.WithAvitoHTTPBase(mock.URL))
	worker := service.NewListingSearchWorker(pool, service.NewAvitoListingProcessor(pool, client))

	users := service.NewUsers(pool)
	user, err := users.CreateUser(ctx, tokens.ProductEmail("listing-filter"), repository.PlanTypeFREE)
	require.NoError(t, err)

	proxies := service.NewProxies(pool)
	_, err = proxies.Create(ctx, mock.URL, repository.ProxyStatusACTIVE, "mock-proxy")
	require.NoError(t, err)

	_, err = service.NewAvitoAccounts(pool).CreateWithSecret(ctx, "acc-filter", repository.AvitoAccountStatusACTIVE, "login@test", "secret")
	require.NoError(t, err)

	filterURL := tokens.FixtureFilterListingURL
	require.True(t, tokens.IsListingFilterURL(filterURL))

	task, err := service.NewTasks(pool).CreateTask(ctx, user.ID, filterURL)
	require.NoError(t, err)

	ok, err := worker.ProcessOne(ctx)
	require.NoError(t, err)
	require.True(t, ok)

	got, err := q.GetTask(ctx, task.ID)
	require.NoError(t, err)
	require.Equal(t, repository.TaskStatusCOMPLETED, got.Status)

	items, err := q.ListTaskItems(ctx, task.ID)
	require.NoError(t, err)
	require.Len(t, items, 2)
	require.Equal(t, tokens.FixtureAvitoItemID1, items[0].AvitoID)
}

// CONTRACT-LISTING-012: filter re-run returns only avito_ids not seen in prior completed tasks.
func TestContract_ListingSearch_FilterURLNewOnly(t *testing.T) {
	ctx := context.Background()
	pool, q := queries(t)
	mock := startAvitoMock(t)
	t.Cleanup(mock.Close)

	client := service.NewAvitoClient(service.WithAvitoHTTPBase(mock.URL))
	worker := service.NewListingSearchWorker(pool, service.NewAvitoListingProcessor(pool, client))

	users := service.NewUsers(pool)
	user, err := users.CreateUser(ctx, tokens.ProductEmail("listing-newonly"), repository.PlanTypeFREE)
	require.NoError(t, err)

	proxies := service.NewProxies(pool)
	_, err = proxies.Create(ctx, mock.URL, repository.ProxyStatusACTIVE, "mock-proxy")
	require.NoError(t, err)

	_, err = service.NewAvitoAccounts(pool).CreateWithSecret(ctx, "acc-newonly", repository.AvitoAccountStatusACTIVE, "login@test", "secret")
	require.NoError(t, err)

	filterURL := tokens.FixtureFilterListingURL
	require.True(t, tokens.IsListingFilterURL(filterURL))

	task1, err := service.NewTasks(pool).CreateTask(ctx, user.ID, filterURL)
	require.NoError(t, err)
	ok, err := worker.ProcessOne(ctx)
	require.NoError(t, err)
	require.True(t, ok)

	items1, err := q.ListTaskItems(ctx, task1.ID)
	require.NoError(t, err)
	require.Len(t, items1, 2)

	task2, err := service.NewTasks(pool).CreateTask(ctx, user.ID, filterURL)
	require.NoError(t, err)
	ok, err = worker.ProcessOne(ctx)
	require.NoError(t, err)
	require.True(t, ok)

	got2, err := q.GetTask(ctx, task2.ID)
	require.NoError(t, err)
	require.Equal(t, repository.TaskStatusCOMPLETED, got2.Status)

	items2, err := q.ListTaskItems(ctx, task2.ID)
	require.NoError(t, err)
	require.Empty(t, items2)

	mobileURL := tokens.FixtureMobileFilterListingURL
	require.True(t, tokens.IsListingFilterURL(mobileURL))
	task3, err := service.NewTasks(pool).CreateTask(ctx, user.ID, mobileURL)
	require.NoError(t, err)
	require.Equal(t, filterURL, task3.Query)

	ok, err = worker.ProcessOne(ctx)
	require.NoError(t, err)
	require.True(t, ok)

	items3, err := q.ListTaskItems(ctx, task3.ID)
	require.NoError(t, err)
	require.Empty(t, items3)
}
