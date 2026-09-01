package tokens

import "time"

// HTTP server timing.
const (
	HTTPReadHeaderTimeout = 5 * time.Second
	HTTPShutdownTimeout   = 10 * time.Second
)

// HTTP paths (Go 1.22+ ServeMux patterns use METHOD + path via HTTPGet/HTTPPost).
const (
	PathHealthz         = "/healthz"
	PathV1Users         = "/v1/users"
	PathV1Tasks         = "/v1/tasks"
	PathV1ProxiesActive = "/v1/proxies/active"
)

// MIME and JSON field / status tokens.
const (
	MIMEApplicationJSON = "application/json"

	JSONFieldStatus  = "status"
	JSONFieldProduct = "product"
	JSONFieldError   = "error"
	JSONFieldID      = "id"
	JSONFieldEmail   = "email"
	JSONFieldUserID  = "user_id"
	JSONFieldQuery   = "query"
	JSONFieldProxies  = "proxies"
	JSONFieldEndpoint = "endpoint"
	JSONFieldPlanType = "plan_type"

	HealthStatusOK        = "ok"
	HealthStatusUnhealthy = "unhealthy"
)

// HTTP error message tokens (stable API surface).
const (
	ErrMsgInvalidJSON          = "invalid json"
	ErrMsgInvalidEmail         = "invalid or empty email address"
	ErrMsgInvalidPlanType      = "invalid plan_type"
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
)

// HTTPGet / HTTPPost build ServeMux patterns.
func HTTPGet(path string) string  { return "GET " + path }
func HTTPPost(path string) string { return "POST " + path }
