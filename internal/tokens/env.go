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
	EnvWorkerTick              = "WORKER_TICK"
	EnvListingSearchProcessor  = "LISTING_SEARCH_PROCESSOR"
	EnvAvitoHTTPBase           = "AVITO_HTTP_BASE"
	EnvRodUserDataDir          = "ROD_USER_DATA_DIR"
	EnvRodBrowser              = "ROD_BROWSER"
)

// AppEnv allowed values.
const (
	AppEnvLocal      = "local"
	AppEnvStaging    = "staging"
	AppEnvProduction = "production"
)
