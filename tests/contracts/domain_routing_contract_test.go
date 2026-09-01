package contracts_test

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"

	"vitek/internal/repository"
	"vitek/internal/service"
	"vitek/internal/tokens"
	httpapi "vitek/internal/transport/httpapi"
)

// CONTRACT-DOMAIN-001: landing / app / www domains are tokenized SoT.
func TestContract_ProductDomains(t *testing.T) {
	require.Equal(t, "vitek.tech", tokens.ProductDomainLanding)
	require.Equal(t, "www.vitek.tech", tokens.ProductDomainWWW)
	require.Equal(t, "app.vitek.tech", tokens.ProductDomainApp)
	require.Equal(t, tokens.ProductDomainLanding, tokens.ProductDomain)
}

// CONTRACT-DOMAIN-002: legacy admin probe is not on the HTTP allowlist.
func TestContract_LegacyAdminNotOnAllowlist(t *testing.T) {
	require.NotContains(t, tokens.HTTPPathAllowlist, tokens.PathProbeLegacyAdmin)
	require.Contains(t, tokens.HTTPPathProbe404, tokens.PathProbeLegacyAdmin)
}

// CONTRACT-DOMAIN-003: GET legacy admin probe returns 404 on every host.
func TestContract_LegacyAdminPathGone(t *testing.T) {
	pool, _ := queries(t)
	handler := httpapi.NewServer(pool).Handler()

	for _, host := range tokens.HTTPPolicyProbeHosts {
		req := httptest.NewRequest(http.MethodGet, tokens.PathProbeLegacyAdmin, nil)
		req.Host = host
		rec := httptest.NewRecorder()
		handler.ServeHTTP(rec, req)
		require.Equal(t, http.StatusNotFound, rec.Code, "host=%s", host)
	}
}

// CONTRACT-DOMAIN-004: landing host serves magic-link landing only (no platform shell).
func TestContract_LandingHostRoot(t *testing.T) {
	pool, _ := queries(t)
	handler := httpapi.NewServer(pool).Handler()

	for _, host := range tokens.HTTPLandingHosts {
		req := httptest.NewRequest(http.MethodGet, tokens.PathRoot, nil)
		req.Host = host
		rec := httptest.NewRecorder()
		handler.ServeHTTP(rec, req)
		require.Equal(t, http.StatusOK, rec.Code, "host=%s", host)
		body := rec.Body.String()
		require.Contains(t, body, tokens.LandingDOMMagicForm)
		require.Contains(t, body, tokens.ProductBrandStem())
		require.NotContains(t, body, tokens.AppDOMScreenPlatform)
		require.NotContains(t, body, tokens.AppNavIDOverview)
	}
}

// CONTRACT-DOMAIN-005: app host serves platform shell at /.
func TestContract_AppHostRoot(t *testing.T) {
	pool, _ := queries(t)
	handler := httpapi.NewServer(pool).Handler()

	req := httptest.NewRequest(http.MethodGet, tokens.PathRoot, nil)
	req.Host = tokens.ProductDomainApp
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)
	require.Equal(t, http.StatusOK, rec.Code)
	body := rec.Body.String()
	require.Contains(t, body, tokens.AppDOMScreenAuth)
	require.Contains(t, body, tokens.AppDOMScreenPlatform)
	require.Contains(t, body, tokens.ServiceCodeListingSearch)
	require.NotContains(t, body, tokens.LandingDOMMagicForm)
}

// CONTRACT-DOMAIN-006: unknown host on / returns 404.
func TestContract_UnknownHostRootNotFound(t *testing.T) {
	pool, _ := queries(t)
	handler := httpapi.NewServer(pool).Handler()

	req := httptest.NewRequest(http.MethodGet, tokens.PathRoot, nil)
	req.Host = tokens.ProductDomainUnknownHost
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)
	require.Equal(t, http.StatusNotFound, rec.Code)
}

// CONTRACT-DOMAIN-007: mailer consume URL targets app domain.
func TestContract_MagicLinkConsumeURLUsesAppDomain(t *testing.T) {
	url := tokens.MagicLinkConsumeURL()
	require.True(t, strings.HasPrefix(url, tokens.HTTPSAppBase()))
	require.Contains(t, url, tokens.PathV1AuthMagicLinkConsume)
}

// CONTRACT-DOMAIN-008: landing host must not serve app SSE.
func TestContract_LandingHostSSEForbidden(t *testing.T) {
	pool, _ := queries(t)
	handler := httpapi.NewServer(pool).Handler()

	req := landingHostRequest(http.MethodGet, tokens.PathAppSSE)
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)
	require.Equal(t, http.StatusNotFound, rec.Code)
}

// CONTRACT-DOMAIN-009: USER session on app host sees platform shell without SSE boot.
func TestContract_AppHostUserSessionShell(t *testing.T) {
	pool, q := queries(t)
	mailer := service.NewMemoryMagicLinkMailer()
	handler := httpapi.NewServer(pool, httpapi.WithMagicLinkMailer(mailer)).Handler()
	ctx := t.Context()

	_, err := q.CreateUserWithRole(ctx, repository.CreateUserWithRoleParams{
		Email: tokens.ProductEmail("app-user"),
		Role:  repository.UserRoleUSER,
	})
	require.NoError(t, err)
	cookie := loginCookie(t, handler, mailer, tokens.ProductEmail("app-user"))

	req := appHostRequest(http.MethodGet, tokens.PathRoot)
	req.Header.Set(tokens.HeaderCookie, tokens.CookieSessionName+"="+cookie)
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)
	require.Equal(t, http.StatusOK, rec.Code)
	body := rec.Body.String()
	require.Contains(t, body, tokens.AppDOMScreenPlatform)
	require.Contains(t, body, tokens.AppClassIsActive)
	require.NotContains(t, body, `data-on:load="@get('`+tokens.PathAppSSE+`')"`)
}
