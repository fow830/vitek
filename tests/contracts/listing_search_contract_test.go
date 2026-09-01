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

// CONTRACT-LISTING-001: tasks.query must be a valid Avito listing URL.
func TestContract_ListingSearch_InvalidURLRejected(t *testing.T) {
	require.False(t, tokens.ValidListingURL(tokens.FixtureInvalidListingURL))
	require.True(t, tokens.ValidListingURL(tokens.FixtureListingURL))

	pool, _ := queries(t)
	handler := httpapi.NewServer(pool).Handler()
	ctx := t.Context()

	users := service.NewUsers(pool)
	user, err := users.CreateUser(ctx, tokens.ProductEmail("listing-url"), repository.PlanTypeFREE)
	require.NoError(t, err)

	body, err := json.Marshal(map[string]any{
		tokens.JSONFieldUserID: pgUUIDString(user.ID),
		tokens.JSONFieldQuery:  tokens.FixtureInvalidListingURL,
	})
	require.NoError(t, err)
	req := withAppHost(httptest.NewRequest(http.MethodPost, tokens.PathV1Tasks, bytes.NewReader(body)))
	req.Header.Set(tokens.HeaderContentType, tokens.MIMEApplicationJSON)
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)
	require.Equal(t, http.StatusBadRequest, rec.Code)

	var errBody map[string]string
	require.NoError(t, json.NewDecoder(rec.Body).Decode(&errBody))
	require.Equal(t, tokens.ErrMsgInvalidListingURL, errBody[tokens.JSONFieldError])
}

// CONTRACT-LISTING-002: valid Avito URL creates a PENDING task.
func TestContract_ListingSearch_ValidURLCreatesTask(t *testing.T) {
	ctx := context.Background()
	pool, _ := queries(t)

	users := service.NewUsers(pool)
	tasks := service.NewTasks(pool)
	user, err := users.CreateUser(ctx, tokens.ProductEmail("listing-create"), repository.PlanTypeFREE)
	require.NoError(t, err)

	task, err := tasks.CreateTask(ctx, user.ID, tokens.FixtureListingURL)
	require.NoError(t, err)
	require.Equal(t, repository.TaskStatusPENDING, task.Status)
	require.Equal(t, tokens.FixtureListingURL, task.Query)
}

// CONTRACT-LISTING-003: worker stub completes task and links similar items.
func TestContract_ListingSearch_WorkerStubCompletesTask(t *testing.T) {
	ctx := context.Background()
	pool, q := queries(t)

	users := service.NewUsers(pool)
	tasks := service.NewTasks(pool)
	worker := service.NewListingSearchWorker(pool, service.NewStubListingProcessor())

	user, err := users.CreateUser(ctx, tokens.ProductEmail("listing-worker"), repository.PlanTypeFREE)
	require.NoError(t, err)

	task, err := tasks.CreateTask(ctx, user.ID, tokens.FixtureListingURL)
	require.NoError(t, err)

	ok, err := worker.ProcessOne(ctx)
	require.NoError(t, err)
	require.True(t, ok)

	got, err := q.GetTask(ctx, task.ID)
	require.NoError(t, err)
	require.Equal(t, repository.TaskStatusCOMPLETED, got.Status)

	count, err := q.CountTaskItems(ctx, task.ID)
	require.NoError(t, err)
	require.Equal(t, int64(tokens.ListingSearchStubResultCount), count)

	items, err := q.ListTaskItems(ctx, task.ID)
	require.NoError(t, err)
	require.Len(t, items, tokens.ListingSearchStubResultCount)
	require.Equal(t, tokens.StubSimilarAvitoID(tokens.FixtureListingID1, 1), items[0].AvitoID)
	for _, it := range items {
		require.NotEmpty(t, it.AvitoID)
		require.NotEmpty(t, it.Title)
	}
}

// CONTRACT-LISTING-004: listing_search schema surface (task_items + COMPLETED status).
func TestContract_ListingSearch_SchemaSurface(t *testing.T) {
	ctx := context.Background()
	pool, _ := queries(t)

	for _, table := range tokens.SchemaListingSearchTables {
		var exists bool
		err := pool.QueryRow(ctx,
			`SELECT EXISTS (
				SELECT 1 FROM information_schema.tables
				WHERE table_schema = 'public' AND table_name = $1
			)`, table).Scan(&exists)
		require.NoError(t, err)
		require.Truef(t, exists, "listing_search table missing: %s", table)
	}

	for _, status := range tokens.SchemaListingSearchTaskStatuses {
		var exists bool
		err := pool.QueryRow(ctx,
			`SELECT EXISTS (
				SELECT 1 FROM pg_enum e
				JOIN pg_type t ON t.oid = e.enumtypid
				WHERE t.typname = 'task_status' AND e.enumlabel = $1
			)`, status).Scan(&exists)
		require.NoError(t, err)
		require.Truef(t, exists, "task_status enum missing value: %s", status)
	}
}

// CONTRACT-LISTING-005: URL/id/stub tokens stay aligned.
func TestContract_ListingSearch_TokensAligned(t *testing.T) {
	require.True(t, tokens.ValidListingURL(tokens.FixtureListingURL))
	require.True(t, tokens.ValidListingURL(tokens.FixtureListingURL2))
	require.Equal(t, tokens.FixtureListingID1, tokens.ListingIDFromURL(tokens.FixtureListingURL))
	require.Equal(t, tokens.FixtureListingID2, tokens.ListingIDFromURL(tokens.FixtureListingURL2))
	require.Equal(
		t,
		tokens.StubSimilarAvitoID(tokens.FixtureListingID1, 1),
		tokens.ListingStubAvitoIDPrefix+"-"+tokens.FixtureListingID1+"-1",
	)
}

// CONTRACT-LISTING-006: mobile listing URLs accepted; catalog URLs rejected.
func TestContract_ListingSearch_MobileListingURL(t *testing.T) {
	require.True(t, tokens.ValidListingURL(tokens.FixtureMobileListingURL))
	require.Equal(t, tokens.FixtureListingID1, tokens.ListingIDFromURL(tokens.FixtureMobileListingURL))
	require.False(t, tokens.ValidListingURL(tokens.FixtureCategoryListingURL))
}
