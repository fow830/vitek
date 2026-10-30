package tokens

import (
	"net"
	"net/url"
	"strings"
	"time"
)

// Proxy pool / health / binding tokens (wave B).
const (
	ProxyPoolMinActive      = 1
	ProxyDeadAfterFails     = 3
	ProxyHealthTimeout      = 15 * time.Second
	// ProxyProbeMaxStatusCode: only 2xx counts as healthy (429/439 antibot must fail).
	ProxyProbeMaxStatusCode = HTTPStatus2xxMax
	AdminLivePollIntervalMs = 5000
	DockerBridgeCIDR        = "172.16.0.0/12"
	ProxyProbePath          = "/"
	// ProxyEndpointDirect means no upstream proxy (egress = host/worker IP).
	ProxyEndpointDirect = "direct"

	ProxyHealthUnknown  = "UNKNOWN"
	ProxyHealthOK       = "OK"
	ProxyHealthDegraded = "DEGRADED"
	ProxyHealthDead     = "DEAD"

	ListingBindingStatusActive   = "ACTIVE"
	ListingBindingStatusPaused   = "PAUSED"
	ListingBindingStatusDisabled = "DISABLED"

	ListingSessionLoggedOut = "LOGGED_OUT"
	ListingSessionLoggingIn = "LOGGING_IN"
	ListingSessionReady     = "READY"
	ListingSessionChallenge = "CHALLENGE"
	ListingSessionError     = "ERROR"

	PathV1AdminBindingSessionLoginSuffix = "/session/login"

	JSONFieldHealth        = "health"
	JSONFieldFailStreak    = "fail_streak"
	JSONFieldLastOkAt      = "last_ok_at"
	JSONFieldLastErr       = "last_err"
	JSONFieldBindings      = "bindings"
	JSONFieldProxyID       = "proxy_id"
	JSONFieldAccountID     = "avito_account_id"
	JSONFieldUserDataDir   = "user_data_dir"
	JSONFieldSessionStatus = "session_status"
	JSONFieldSessionErr    = "session_err"

	ErrMsgListingSearchNoBinding    = "no active listing fetch binding"
	ErrMsgAdminBindingsFailed       = "admin bindings failed"
	ErrMsgInvalidBindingStatus      = "invalid binding status"
	ErrMsgProxyProbeFailed          = "proxy probe failed"
	ErrMsgSessionWarmFailed         = "session warm failed"
	ErrMsgSessionChallenge          = "avito session challenge"
	ErrMsgRodUserDataDirRequired    = "ROD_USER_DATA_DIR is required for rod in production"
	ErrMsgProxyPoolTooSmall         = "proxy pool below minimum active"
	ErrMsgProxyPoolAllDead          = "all ACTIVE proxies are DEAD"
	ErrMsgProxyPoolDockerBridgeOnly = "production proxy pool is docker-bridge only"

	FixtureBindingUserDataDir = "/tmp/vitek-rod-profile"
	FixtureWarmSessionHTML    = "<html>ok</html>"
	FixtureEmptySERPHTML      = "<html></html>"
	LogWorkerProxyProbe       = "proxy_probe: ok=%d fail=%d"
	LogWorkerSessionWarm      = "session_warm: n=%d"
	RodInterRequestDelay      = 200 * time.Millisecond
	DefaultRodUserDataDir = "/var/lib/vitek/rod"
	DefaultRodChromeBin   = "/usr/bin/google-chrome-stable"

	RodFetchModeHTTP    = "http"
	RodFetchModeChrome  = "chrome"
	DefaultRodFetchMode = RodFetchModeHTTP

	// Multi-step Chrome profile warm (home → dwell/scroll → secondary feed).
	RodChromeWarmHomeDwell      = 8 * time.Second
	RodChromeWarmStepDwell      = 5 * time.Second
	RodChromeWarmScrollPause    = 800 * time.Millisecond
	RodChromeWarmScrollSteps    = 3
	RodChromeWarmScrollDelta    = 480
	RodChromeWarmScrollX        = 960.0
	RodChromeWarmScrollY        = 540.0
	RodChromeWarmTimeout        = 2 * time.Minute
	RodChromeWarmSecondaryPath  = "/all"
)

// RodChromeWarmSecondaryURL is the second warm navigation (public feed, not user filters).
func RodChromeWarmSecondaryURL() string {
	return AvitoHTTPSBase + RodChromeWarmSecondaryPath
}

var SchemaProxyHealthStatuses = []string{
	ProxyHealthUnknown, ProxyHealthOK, ProxyHealthDegraded, ProxyHealthDead,
}

var SchemaListingBindingStatuses = []string{
	ListingBindingStatusActive, ListingBindingStatusPaused, ListingBindingStatusDisabled,
}

var SchemaListingSessionStatuses = []string{
	ListingSessionLoggedOut, ListingSessionLoggingIn, ListingSessionReady, ListingSessionChallenge, ListingSessionError,
}

func HTTPPathAdminBindingSessionLogin() string {
	return HTTPPathID(PathV1AdminBindings) + PathV1AdminBindingSessionLoginSuffix
}

// IsDirectProxyEndpoint reports an explicit "no proxy" binding endpoint.
func IsDirectProxyEndpoint(endpoint string) bool {
	e := strings.TrimSpace(strings.ToLower(endpoint))
	return e == "" || e == ProxyEndpointDirect || e == ProxyEndpointDirect+":" || e == ProxyEndpointDirect+"://"
}

// IsDockerBridgeProxyEndpoint reports docker-bridge-style hosts (DockerBridgeCIDR).
func IsDockerBridgeProxyEndpoint(endpoint string) bool {
	if IsDirectProxyEndpoint(endpoint) {
		return false
	}
	u, err := url.Parse(strings.TrimSpace(endpoint))
	if err != nil || u.Hostname() == "" {
		return false
	}
	ip := net.ParseIP(u.Hostname())
	if ip == nil {
		return false
	}
	_, cidr, err := net.ParseCIDR(DockerBridgeCIDR)
	if err != nil {
		return false
	}
	return cidr.Contains(ip)
}
