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
	ProxyHealthTimeout      = 5 * time.Second
	ProxyProbeMaxStatusCode = 499
	AdminLivePollIntervalMs = 5000
	DockerBridgeCIDR        = "172.16.0.0/12"
	ProxyProbePath          = "/"

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
	ErrMsgProxyPoolDockerBridgeOnly = "production proxy pool is docker-bridge only"

	FixtureBindingUserDataDir = "/tmp/vitek-rod-profile"
	FixtureWarmSessionHTML    = "<html>ok</html>"
	FixtureEmptySERPHTML      = "<html></html>"
	LogWorkerProxyProbe       = "proxy_probe: ok=%d fail=%d"
	RodInterRequestDelay      = 200 * time.Millisecond
	DefaultRodUserDataDir     = "/var/lib/vitek/rod"
)

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

// IsDockerBridgeProxyEndpoint reports docker-bridge-style hosts (DockerBridgeCIDR).
func IsDockerBridgeProxyEndpoint(endpoint string) bool {
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
