package tokens

// Environment variable names (no string literals for keys outside this package).
const (
	EnvAppEnv      = "APP_ENV"
	EnvHTTPAddr    = "HTTP_ADDR"
	EnvLogLevel    = "LOG_LEVEL"
	EnvDatabaseURL = "DATABASE_URL"
	EnvRedisURL    = "REDIS_URL"

	EnvPostgresUser     = "POSTGRES_USER"
	EnvPostgresPassword = "POSTGRES_PASSWORD"
	EnvPostgresDB       = "POSTGRES_DB"
	EnvPostgresPort     = "POSTGRES_PORT"
	EnvRedisPort        = "REDIS_PORT"
	EnvHTTPPort         = "HTTP_PORT"
)

// AppEnv allowed values.
const (
	AppEnvLocal      = "local"
	AppEnvStaging    = "staging"
	AppEnvProduction = "production"
)
