package contracts_test

import (
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

// CONTRACT-TASK-VIEW-001: session user can read own completed task and results.
func TestContract_TaskView_SessionUserFlow(t *testing.T) {
	ctx := context.Background()
	pool, _ := queries(t)
	mailer := service.NewMemoryMagicLinkMailer()
	handler := httpapi.NewServer(pool, httpapi.WithMagicLinkMailer(mailer)).Handler()

	users := service.NewUsers(pool)
	user, err := users.CreateUser(ctx, tokens.ProductEmail("task-view"), repository.PlanTypeFREE)
	require.NoError(t, err)
	cookie := loginCookie(t, handler, mailer, tokens.ProductEmail("task-view"))

	worker := service.NewListingSearchWorker(pool, service.NewStubListingProcessor())
	task, err := service.NewTasks(pool).CreateTask(ctx, user.ID, tokens.FixtureListingURL)
	require.NoError(t, err)
	ok, err := worker.ProcessOne(ctx)
	require.NoError(t, err)
	require.True(t, ok)

	req := appHostRequest(http.MethodGet, tokens.HTTPResourceID(tokens.PathV1Tasks, pgUUIDString(task.ID)))
	req.Header.Set(tokens.HeaderCookie, tokens.CookieSessionName+"="+cookie)
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)
	require.Equal(t, http.StatusOK, rec.Code)

	var got map[string]any
	require.NoError(t, json.NewDecoder(rec.Body).Decode(&got))
	require.Equal(t, string(repository.TaskStatusCOMPLETED), got[tokens.JSONFieldStatus])

	req = appHostRequest(http.MethodGet, tokens.HTTPResourceID(tokens.PathV1Tasks, pgUUIDString(task.ID))+tokens.PathV1TaskResultsSuffix)
	req.Header.Set(tokens.HeaderCookie, tokens.CookieSessionName+"="+cookie)
	rec = httptest.NewRecorder()
	handler.ServeHTTP(rec, req)
	require.Equal(t, http.StatusOK, rec.Code)

	var results map[string]any
	require.NoError(t, json.NewDecoder(rec.Body).Decode(&results))
	list := results[tokens.JSONFieldResults].([]any)
	require.Len(t, list, tokens.ListingSearchStubResultCount)
}

// CONTRACT-TASK-VIEW-002: another user's task returns forbidden.
func TestContract_TaskView_ForbiddenForOtherUser(t *testing.T) {
	ctx := context.Background()
	pool, _ := queries(t)
	mailer := service.NewMemoryMagicLinkMailer()
	handler := httpapi.NewServer(pool, httpapi.WithMagicLinkMailer(mailer)).Handler()

	users := service.NewUsers(pool)
	owner, err := users.CreateUser(ctx, tokens.ProductEmail("task-owner"), repository.PlanTypeFREE)
	require.NoError(t, err)
	_, err = users.CreateUser(ctx, tokens.ProductEmail("task-other"), repository.PlanTypeFREE)
	require.NoError(t, err)

	task, err := service.NewTasks(pool).CreateTask(ctx, owner.ID, tokens.FixtureListingURL)
	require.NoError(t, err)

	otherCookie := loginCookie(t, handler, mailer, tokens.ProductEmail("task-other"))
	req := appHostRequest(http.MethodGet, tokens.HTTPResourceID(tokens.PathV1Tasks, pgUUIDString(task.ID)))
	req.Header.Set(tokens.HeaderCookie, tokens.CookieSessionName+"="+otherCookie)
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)
	require.Equal(t, http.StatusForbidden, rec.Code)
}

// CONTRACT-TASK-VIEW-003: task view paths are on HTTP allowlist.
func TestContract_TaskView_PathsOnAllowlist(t *testing.T) {
	require.Contains(t, tokens.HTTPPathAllowlist, tokens.HTTPPathID(tokens.PathV1Tasks))
	require.Contains(t, tokens.HTTPPathAllowlist, tokens.HTTPPathTaskResults())
	require.Contains(t, tokens.HTTPPathAllowlist, tokens.PathV1MeTasks)
}
