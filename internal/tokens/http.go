package tokens

import "time"

// HTTP server timing.
const (
	HTTPReadHeaderTimeout = 5 * time.Second
	HTTPShutdownTimeout   = 10 * time.Second
	SSETickInterval       = 500 * time.Millisecond
)

// HTTP paths (Go 1.22+ ServeMux patterns use METHOD + path via HTTPGet/HTTPPost).
const (
	PathHealthz                = "/healthz"
	PathV1Users                = "/v1/users"
	PathV1Tasks                = "/v1/tasks"
	PathV1TaskResultsSuffix    = "/results"
	PathV1MeTasks              = "/v1/me/tasks"
	PathV1ProxiesActive        = "/v1/proxies/active"
	PathV1AuthMagicLink        = "/v1/auth/magic-link"
	PathV1AuthMagicLinkConsume = "/v1/auth/magic-link/consume"
	PathV1AuthMagicLinkOpen    = "/v1/auth/magic-link/open"
	PathV1AuthLogout           = "/v1/auth/logout"
	PathV1AdminProxies         = "/v1/admin/proxies"
	PathV1AdminAvitoAccounts   = "/v1/admin/avito-accounts"
	PathRoot                   = "/"
	PathAppSSE                 = "/sse"
	PathTokensCSS              = "/tokens.css"
	PathParamID                = "id"
	PathSuffixID               = "/{id}"
)

// PathProbe* are fuzz samples that must always 404 (never registered routes).
const (
	PathProbeLegacyAdmin    = "/admin"
	PathProbeLegacyAdminSlash = "/admin/"
	PathProbeLegacyAdminSSE = "/admin/sse"
	PathProbeFoo            = "/foo"
	PathProbeAPI            = "/api"
	PathProbeV1Hack         = "/v1/hack"
	PathProbeWPLogin        = "/wp-login.php"
	PathProbeDotEnv         = "/.env"
)

// MIME and JSON field / status tokens.
const (
	MIMEApplicationJSON = "application/json"
	MIMETextHTML        = "text/html; charset=utf-8"
	MIMETextCSS         = "text/css; charset=utf-8"
	MIMETextEventStream = "text/event-stream"

	JSONFieldStatus      = "status"
	JSONFieldProduct     = "product"
	JSONFieldError       = "error"
	JSONFieldID          = "id"
	JSONFieldEmail       = "email"
	JSONFieldUserID      = "user_id"
	JSONFieldQuery       = "query"
	JSONFieldProxies     = "proxies"
	JSONFieldEndpoint    = "endpoint"
	JSONFieldPlanType    = "plan_type"
	JSONFieldToken       = "token"
	JSONFieldMagicLinkURL = "magic_link_url"
	JSONFieldPassword     = "password"
	JSONFieldRole        = "role"
	JSONFieldLabel       = "label"
	JSONFieldAccounts    = "accounts"
	JSONFieldExternalRef = "external_ref"
	JSONFieldResults     = "results"
	JSONFieldItems       = "items"
	JSONFieldAvitoID     = "avito_id"
	JSONFieldTitle       = "title"
	JSONFieldRank        = "rank"
	JSONFieldTasks       = "tasks"

	HeaderContentType = "Content-Type"
	HeaderSetCookie   = "Set-Cookie"
	HeaderCookie      = "Cookie"
	HeaderUserAgent   = "User-Agent"
	HeaderAccept      = "Accept"
	HeaderAcceptLanguage = "Accept-Language"

	CookieSessionName  = "vitek_session"
	CookiePath         = "/"
	CookieSameSite     = "Lax"
	CookieAttrHttpOnly = "HttpOnly"
	CookieAttrSecure   = "Secure"
	CookieAttrPath     = "Path="
	CookieAttrSameSite = "SameSite="

	HealthStatusOK        = "ok"
	HealthStatusUnhealthy = "unhealthy"

	DatastarPatchMarker = "datastar-patch"
	AttrDataStar        = "data-star"

	LocaleHTML  = "ru"
	LocaleBCP47 = "ru-RU"

	QueryParamToken = "token"

	BoolStringTrue  = "true"
	BoolStringFalse = "false"

	FixtureInvalidEnum = "NOPE"

	FixtureAdminProxyEndpoint = "http://admin-proxy.example:8080"
)

// Auth / session TTLs.
const (
	MagicLinkTTL = 15 * time.Minute
	SessionTTL   = 7 * 24 * time.Hour
)

// HTTP error message tokens (stable API surface).
const (
	ErrMsgInvalidJSON          = "invalid json"
	ErrMsgInvalidEmail         = "invalid or empty email address"
	ErrMsgInvalidPlanType      = "invalid plan_type"
	ErrMsgInvalidProxyStatus   = "invalid proxy status"
	ErrMsgInvalidAvitoStatus   = "invalid avito account status"
	ErrMsgEmailConflict        = "user with this email already exists"
	ErrMsgCreateUserFailed     = "failed to create user"
	ErrMsgInvalidUserID        = "invalid user_id"
	ErrMsgCreateTaskFailed     = "create task failed"
	ErrMsgListProxiesFailed    = "list proxies failed"
	ErrMsgDatabaseUnavailable  = "database unavailable"
	ErrMsgSubscriptionLimit    = "subscription task limit exceeded"
	ErrMsgNoActiveSubscription = "no active subscription"
	ErrMsgDuplicateAvitoID     = "duplicate avito_id"
	ErrMsgServiceNotEntitled   = "user not entitled to service"
	ErrMsgInvalidMagicToken    = "invalid or expired magic link"
	ErrMsgMagicLinkFailed      = "magic link request failed"
	ErrMsgUnauthorized         = "unauthorized"
	ErrMsgForbidden            = "forbidden"
	ErrMsgAdminProxiesFailed   = "admin proxies failed"
	ErrMsgAdminAvitoFailed     = "admin avito accounts failed"
	ErrMsgInvalidResourceID    = "invalid resource id"
	ErrMsgInvalidListingURL    = "invalid avito listing url"
	ErrMsgTaskNotFound         = "task not found"
	ErrMsgListTasksFailed      = "list tasks failed"
	ErrMsgGetTaskFailed        = "get task failed"
	ErrMsgInvalidPassword      = "invalid password"
)

// HTTPPathTaskResults returns GET /v1/tasks/{id}/results pattern.
func HTTPPathTaskResults() string {
	return HTTPPathID(PathV1Tasks) + PathV1TaskResultsSuffix
}
func HTTPGet(path string) string   { return "GET " + path }
func HTTPPost(path string) string  { return "POST " + path }
func HTTPPatch(path string) string { return "PATCH " + path }

// HTTPPathID appends the ServeMux {id} suffix.
func HTTPPathID(base string) string { return base + PathSuffixID }

// HTTPResourceID joins a collection path with a concrete resource id.
func HTTPResourceID(base, id string) string { return base + "/" + id }
