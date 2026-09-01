package contracts_test

import (
	"bufio"
	"context"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	"vitek/internal/repository"
	"vitek/internal/service"
	"vitek/internal/tokens"
	httpapi "vitek/internal/transport/httpapi"
)

// CONTRACT-APP-LIVE-001: app host / serves tokenized HTML; /sse streams Datastar events for admin.
func TestContract_AppLiveSurface(t *testing.T) {
	pool, q := queries(t)
	mailer := service.NewMemoryMagicLinkMailer()
	handler := httpapi.NewServer(pool, httpapi.WithMagicLinkMailer(mailer)).Handler()
	ctx := t.Context()

	_, err := q.CreateUserWithRole(ctx, repository.CreateUserWithRoleParams{
		Email: tokens.ProductEmail("live-admin"),
		Role:  repository.UserRoleADMIN,
	})
	require.NoError(t, err)
	cookie := loginCookie(t, handler, mailer, tokens.ProductEmail("live-admin"))

	req := appHostRequest(http.MethodGet, tokens.PathRoot)
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)
	require.Equal(t, http.StatusOK, rec.Code)
	require.Contains(t, rec.Body.String(), tokens.ProductBrandStem())

	req = appHostRequest(http.MethodGet, tokens.PathRoot)
	req.Header.Set(tokens.HeaderCookie, tokens.CookieSessionName+"="+cookie)
	rec = httptest.NewRecorder()
	handler.ServeHTTP(rec, req)
	require.Equal(t, http.StatusOK, rec.Code)
	require.Equal(t, tokens.MIMETextHTML, rec.Header().Get(tokens.HeaderContentType))
	body := rec.Body.String()
	require.Contains(t, body, tokens.ProductBrandStem())
	require.Contains(t, body, tokens.ProductBrandAccent())
	require.Contains(t, body, tokens.ServiceCodeListingSearch)
	require.Contains(t, strings.ToLower(body), tokens.AttrDataStar)
	require.Contains(t, body, tokens.AppDOMScreenPlatform)
	require.Contains(t, body, tokens.AppClassIsActive)

	sreq := appHostRequest(http.MethodGet, tokens.PathAppSSE)
	srec := httptest.NewRecorder()
	handler.ServeHTTP(srec, sreq)
	require.Equal(t, http.StatusUnauthorized, srec.Code)

	_, err = q.CreateUserWithRole(ctx, repository.CreateUserWithRoleParams{
		Email: tokens.ProductEmail("live-user"),
		Role:  repository.UserRoleUSER,
	})
	require.NoError(t, err)
	userCookie := loginCookie(t, handler, mailer, tokens.ProductEmail("live-user"))
	ureq := appHostRequest(http.MethodGet, tokens.PathAppSSE)
	ureq.Header.Set(tokens.HeaderCookie, tokens.CookieSessionName+"="+userCookie)
	urec := httptest.NewRecorder()
	handler.ServeHTTP(urec, ureq)
	require.Equal(t, http.StatusForbidden, urec.Code)

	ts := httptest.NewServer(handler)
	t.Cleanup(ts.Close)

	streamCtx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	t.Cleanup(cancel)
	sreq, err = http.NewRequestWithContext(streamCtx, http.MethodGet, ts.URL+tokens.PathAppSSE, nil)
	require.NoError(t, err)
	sreq.Host = tokens.ProductDomainApp
	sreq.Header.Set(tokens.HeaderCookie, tokens.CookieSessionName+"="+cookie)

	sresp, err := http.DefaultClient.Do(sreq)
	require.NoError(t, err)
	t.Cleanup(func() { _ = sresp.Body.Close() })
	require.Equal(t, http.StatusOK, sresp.StatusCode)
	require.Contains(t, sresp.Header.Get(tokens.HeaderContentType), tokens.MIMETextEventStream)

	got := readSSEUntil(t, sresp.Body, 3*time.Second, func(acc string) bool {
		return strings.Contains(acc, tokens.AppDOMStatAvito)
	})
	require.Contains(t, got, tokens.AppDOMStatAvito)
	require.Contains(t, got, tokens.DatastarPatchMarker)
}

func readSSEUntil(t *testing.T, r io.Reader, wait time.Duration, ok func(string) bool) string {
	t.Helper()
	done := make(chan string, 1)
	go func() {
		sc := bufio.NewScanner(r)
		sc.Buffer(make([]byte, 0, 64*1024), 1024*1024)
		var acc strings.Builder
		for sc.Scan() {
			acc.WriteString(sc.Text())
			acc.WriteByte('\n')
			if ok(acc.String()) {
				done <- acc.String()
				return
			}
		}
		done <- acc.String()
	}()
	select {
	case body := <-done:
		return body
	case <-time.After(wait):
		t.Fatalf("SSE produced no matching events within %s", wait)
		return ""
	}
}
