package contracts_test

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/require"

	"vitek/internal/tokens"
	httpapi "vitek/internal/transport/httpapi"
)

// CONTRACT-HTTP-POLICY-001: probe paths from tokens.HTTPPathProbe404 always 404 (any host).
func TestContract_UnlistedPathsAlways404(t *testing.T) {
	pool, _ := queries(t)
	handler := httpapi.NewServer(pool).Handler()

	hosts := []string{
		tokens.ProductDomainLanding,
		tokens.ProductDomainWWW,
		tokens.ProductDomainApp,
		"evil.example",
	}
	for _, path := range tokens.HTTPPathProbe404 {
		for _, host := range hosts {
			req := httptest.NewRequest(http.MethodGet, path, nil)
			req.Host = host
			rec := httptest.NewRecorder()
			handler.ServeHTTP(rec, req)
			require.Equal(t, http.StatusNotFound, rec.Code, "GET %s host=%s", path, host)
		}
	}
}

// CONTRACT-HTTP-POLICY-002: landing host rejects platform-only API paths.
func TestContract_LandingHostRejectsAppAPI(t *testing.T) {
	pool, _ := queries(t)
	handler := httpapi.NewServer(pool).Handler()

	probes := []struct {
		method string
		path   string
	}{
		{http.MethodPost, tokens.PathV1AuthMagicLinkConsume},
		{http.MethodGet, tokens.PathV1AdminProxies},
		{http.MethodPost, tokens.PathV1Users},
		{http.MethodGet, tokens.PathAppSSE},
	}
	for _, host := range []string{tokens.ProductDomainLanding, tokens.ProductDomainWWW} {
		for _, p := range probes {
			req := httptest.NewRequest(p.method, p.path, nil)
			req.Host = host
			rec := httptest.NewRecorder()
			handler.ServeHTTP(rec, req)
			require.Equal(t, http.StatusNotFound, rec.Code, "%s %s host=%s", p.method, p.path, host)
		}
	}
}

// CONTRACT-HTTP-POLICY-003: HTTP policy routes cover HTTPPathAllowlist paths.
func TestContract_HTTPPolicyCoversAllowlistPaths(t *testing.T) {
	allowedPaths := map[string]struct{}{}
	for _, route := range tokens.HTTPGlobalRoutes {
		allowedPaths[route.Path] = struct{}{}
	}
	for _, route := range tokens.HTTPLandingRoutes {
		allowedPaths[route.Path] = struct{}{}
	}
	for _, route := range tokens.HTTPAppRoutes {
		allowedPaths[route.Path] = struct{}{}
	}
	for _, path := range tokens.HTTPPathAllowlist {
		_, ok := allowedPaths[path]
		require.Truef(t, ok, "HTTPPathAllowlist path %q missing from route policy", path)
	}
}

// CONTRACT-HTTP-POLICY-004: probe paths never overlap allowlist.
func TestContract_Probe404NotInAllowlist(t *testing.T) {
	allow := map[string]struct{}{}
	for _, path := range tokens.HTTPPathAllowlist {
		allow[path] = struct{}{}
	}
	for _, path := range tokens.HTTPPathProbe404 {
		_, ok := allow[tokens.NormalizeRequestPath(path)]
		require.Falsef(t, ok, "probe path %q must not be allowlisted", path)
	}
}

// CONTRACT-HTTP-POLICY-005: wrong HTTP method on contracted path → 404.
func TestContract_WrongMethodOnValidPath404(t *testing.T) {
	pool, _ := queries(t)
	handler := httpapi.NewServer(pool).Handler()

	req := withAppHost(httptest.NewRequest(http.MethodGet, tokens.PathV1Users, nil))
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)
	require.Equal(t, http.StatusNotFound, rec.Code)
}
