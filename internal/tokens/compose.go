package tokens

// Healthcheck timing for local compose.
const (
	HealthInterval = "5s"
	HealthTimeout  = "5s"
	HealthRetries  = "10"
)

// Compose service names.
const (
	ComposeServicePostgres = "postgres"
)

// RenderComposeYAML returns the canonical docker-compose.yml body (Postgres only until Redis is contracted).
func RenderComposeYAML() string {
	return "" +
		"services:\n" +
		"  " + ComposeServicePostgres + ":\n" +
		"    image: " + ImagePostgres + "\n" +
		"    environment:\n" +
		"      POSTGRES_USER: ${" + EnvPostgresUser + ":-" + DefaultPostgresUser + "}\n" +
		"      POSTGRES_PASSWORD: ${" + EnvPostgresPassword + ":-" + DefaultPostgresPassword + "}\n" +
		"      POSTGRES_DB: ${" + EnvPostgresDB + ":-" + DefaultPostgresDB + "}\n" +
		"    ports:\n" +
		"      - \"${" + EnvPostgresPort + ":-" + DefaultPostgresPort + "}:" + ContainerPostgresPort + "\"\n" +
		"    volumes:\n" +
		"      - " + VolumePostgres + ":/var/lib/postgresql/data\n" +
		"    healthcheck:\n" +
		"      test: [\"CMD-SHELL\", \"pg_isready -U $$" + EnvPostgresUser + " -d $$" + EnvPostgresDB + "\"]\n" +
		"      interval: " + HealthInterval + "\n" +
		"      timeout: " + HealthTimeout + "\n" +
		"      retries: " + HealthRetries + "\n" +
		"\n" +
		"volumes:\n" +
		"  " + VolumePostgres + ":\n"
}
