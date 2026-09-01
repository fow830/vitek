package tokens

import (
	"net/http"
	"strings"
)

// HTTPRouteSpec is a contracted method + path (path may include PathSuffixID).
type HTTPRouteSpec struct {
	Method string
	Path   string
}

// HTTPGlobalRoutes work on any known or unknown host (ops probes).
var HTTPGlobalRoutes = []HTTPRouteSpec{
	{Method: http.MethodGet, Path: PathHealthz},
	{Method: http.MethodGet, Path: PathTokensCSS},
}

// HTTPLandingRoutes are the only routes on marketing hosts.
var HTTPLandingRoutes = []HTTPRouteSpec{
	{Method: http.MethodGet, Path: PathRoot},
	{Method: http.MethodPost, Path: PathV1AuthMagicLink},
}

// HTTPAppRoutes are the full platform surface on app.vitek.tech.
var HTTPAppRoutes = []HTTPRouteSpec{
	{Method: http.MethodPost, Path: PathV1Users},
	{Method: http.MethodPost, Path: PathV1Tasks},
	{Method: http.MethodGet, Path: PathV1ProxiesActive},
	{Method: http.MethodPost, Path: PathV1AuthMagicLink},
	{Method: http.MethodPost, Path: PathV1AuthMagicLinkConsume},
	{Method: http.MethodPost, Path: PathV1AuthLogout},
	{Method: http.MethodGet, Path: PathV1AdminProxies},
	{Method: http.MethodPost, Path: PathV1AdminProxies},
	{Method: http.MethodPatch, Path: HTTPPathID(PathV1AdminProxies)},
	{Method: http.MethodGet, Path: PathV1AdminAvitoAccounts},
	{Method: http.MethodPost, Path: PathV1AdminAvitoAccounts},
	{Method: http.MethodPatch, Path: HTTPPathID(PathV1AdminAvitoAccounts)},
	{Method: http.MethodGet, Path: PathRoot},
	{Method: http.MethodGet, Path: PathAppSSE},
}

// HTTPPathProbe404 paths that must never resolve (contract fuzz samples).
var HTTPPathProbe404 = []string{
	PathProbeLegacyAdmin,
	PathProbeLegacyAdmin + "/",
	PathProbeLegacyAdminSSE,
	PathProbeFoo,
	PathProbeAPI,
	PathProbeV1Hack,
	PathProbeWPLogin,
	PathProbeDotEnv,
}

// HTTPPathAllowlist is the contracted path surface in stable order (derived from route policy).
var HTTPPathAllowlist = []string{
	PathHealthz,
	PathV1Users,
	PathV1Tasks,
	PathV1ProxiesActive,
	PathV1AuthMagicLink,
	PathV1AuthMagicLinkConsume,
	PathV1AuthLogout,
	PathV1AdminProxies,
	HTTPPathID(PathV1AdminProxies),
	PathV1AdminAvitoAccounts,
	HTTPPathID(PathV1AdminAvitoAccounts),
	PathRoot,
	PathAppSSE,
	PathTokensCSS,
}

func init() {
	covered := map[string]struct{}{}
	for _, routes := range [][]HTTPRouteSpec{HTTPGlobalRoutes, HTTPLandingRoutes, HTTPAppRoutes} {
		for _, route := range routes {
			covered[route.Path] = struct{}{}
		}
	}
	for _, path := range HTTPPathAllowlist {
		if _, ok := covered[path]; !ok {
			panic("tokens: HTTPPathAllowlist path not in route policy: " + path)
		}
	}
	for _, path := range HTTPPathProbe404 {
		if _, ok := covered[path]; ok {
			panic("tokens: HTTPPathProbe404 path overlaps allowlist: " + path)
		}
	}
}

// NormalizeRequestPath canonicalizes URL paths for policy checks.
func NormalizeRequestPath(path string) string {
	if path == "" {
		return PathRoot
	}
	if len(path) > 1 && strings.HasSuffix(path, "/") {
		path = strings.TrimSuffix(path, "/")
	}
	return path
}

// RoutePathMatches reports whether requestPath matches a route path pattern.
func RoutePathMatches(routePath, requestPath string) bool {
	requestPath = NormalizeRequestPath(requestPath)
	if !strings.Contains(routePath, PathSuffixID) {
		return routePath == requestPath
	}
	prefix := strings.TrimSuffix(routePath, PathSuffixID)
	if requestPath == prefix {
		return false
	}
	if !strings.HasPrefix(requestPath, prefix+"/") {
		return false
	}
	rest := strings.TrimPrefix(requestPath, prefix+"/")
	return rest != "" && !strings.Contains(rest, "/")
}

// RequestMatchesRoute compares method + path against a route spec.
func RequestMatchesRoute(method, path string, route HTTPRouteSpec) bool {
	return method == route.Method && RoutePathMatches(route.Path, path)
}

// IsAllowedHTTPRequest is the iron allowlist gate (unknown → false → 404).
func IsAllowedHTTPRequest(method, path, host string) bool {
	path = NormalizeRequestPath(path)
	host = NormalizeHost(host)

	for _, route := range HTTPGlobalRoutes {
		if RequestMatchesRoute(method, path, route) {
			return true
		}
	}

	if IsLandingHost(host) {
		for _, route := range HTTPLandingRoutes {
			if RequestMatchesRoute(method, path, route) {
				return true
			}
		}
		return false
	}

	if IsAppHost(host) {
		for _, route := range HTTPAppRoutes {
			if RequestMatchesRoute(method, path, route) {
				return true
			}
		}
		return false
	}

	return false
}
