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
	ComposeServiceRedis    = "redis"
)

// RenderComposeYAML returns the canonical docker-compose.yml body.
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
		"      test: [\"CMD-SHELL\", \"pg_isready -U $$POSTGRES_USER -d $$POSTGRES_DB\"]\n" +
		"      interval: " + HealthInterval + "\n" +
		"      timeout: " + HealthTimeout + "\n" +
		"      retries: " + HealthRetries + "\n" +
		"\n" +
		"  " + ComposeServiceRedis + ":\n" +
		"    image: " + ImageRedis + "\n" +
		"    ports:\n" +
		"      - \"${" + EnvRedisPort + ":-" + DefaultRedisPort + "}:" + ContainerRedisPort + "\"\n" +
		"    healthcheck:\n" +
		"      test: [\"CMD\", \"redis-cli\", \"ping\"]\n" +
		"      interval: " + HealthInterval + "\n" +
		"      timeout: " + HealthTimeout + "\n" +
		"      retries: " + HealthRetries + "\n" +
		"\n" +
		"volumes:\n" +
		"  " + VolumePostgres + ":\n"
}
