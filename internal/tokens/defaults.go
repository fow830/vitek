package tokens

import "fmt"

// DefaultAppEnv is the local-dev default.
const DefaultAppEnv = AppEnvLocal

// Local host for composed DSNs (dev only).
const DefaultHostLocal = "localhost"

// Postgres SSL mode for local compose / tests.
const DefaultPostgresSSLMode = "disable"

// Atomic local-dev defaults.
const (
	DefaultHTTPPort         = "8080"
	DefaultPostgresUser     = "vitek"
	DefaultPostgresPassword = "vitek"
	DefaultPostgresDB       = "vitek"
	DefaultTestPostgresDB   = "vitek_test"
	DefaultPostgresPort     = "5432"
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
