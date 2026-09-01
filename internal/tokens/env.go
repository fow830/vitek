package tokens

// Environment variable names (no string literals for keys outside this package).
const (
	EnvAppEnv      = "APP_ENV"
	EnvHTTPAddr    = "HTTP_ADDR"
	EnvDatabaseURL = "DATABASE_URL"

	EnvPostgresUser     = "POSTGRES_USER"
	EnvPostgresPassword = "POSTGRES_PASSWORD"
	EnvPostgresDB       = "POSTGRES_DB"
	EnvPostgresPort     = "POSTGRES_PORT"
	EnvHTTPPort         = "HTTP_PORT"
	EnvWorkerTick       = "WORKER_TICK"
)

// AppEnv allowed values.
const (
	AppEnvLocal      = "local"
	AppEnvStaging    = "staging"
	AppEnvProduction = "production"
)
