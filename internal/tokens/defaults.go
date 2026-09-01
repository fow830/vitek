package tokens

import "fmt"

// Log levels.
const (
	LogLevelDebug = "debug"
	LogLevelInfo  = "info"
	LogLevelWarn  = "warn"
	LogLevelError = "error"
)

// DefaultLogLevel is the local-dev default.
const DefaultLogLevel = LogLevelInfo

// DefaultAppEnv is the local-dev default.
const DefaultAppEnv = AppEnvLocal

// Local host for composed DSNs (dev only).
const DefaultHostLocal = "localhost"

// Postgres SSL mode for local compose.
const DefaultPostgresSSLMode = "disable"

// Redis DB index for local compose.
const DefaultRedisDB = "0"

// Atomic local-dev defaults.
const (
	DefaultHTTPPort         = "8080"
	DefaultPostgresUser     = "vitek"
	DefaultPostgresPassword = "vitek"
	DefaultPostgresDB       = "vitek"
	DefaultTestPostgresDB   = "vitek_test"
	DefaultPostgresPort     = "5432"
	DefaultRedisPort        = "6379"
)

// DefaultHTTPAddr is derived from DefaultHTTPPort.
func DefaultHTTPAddr() string {
	return ":" + DefaultHTTPPort
}

// DefaultDatabaseURL is derived from atomic postgres defaults.
func DefaultDatabaseURL() string {
	return fmt.Sprintf(
		"postgres://%s:%s@%s:%s/%s?sslmode=%s",
		DefaultPostgresUser,
		DefaultPostgresPassword,
		DefaultHostLocal,
		DefaultPostgresPort,
		DefaultPostgresDB,
		DefaultPostgresSSLMode,
	)
}

// DefaultRedisURL is derived from atomic redis defaults.
func DefaultRedisURL() string {
	return fmt.Sprintf(
		"redis://%s:%s/%s",
		DefaultHostLocal,
		DefaultRedisPort,
		DefaultRedisDB,
	)
}
