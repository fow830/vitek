package tokens

// Infrastructure image and volume tokens (docker compose / CI).
const (
	ImagePostgres = "postgres:17-alpine"
	ImageRedis    = "redis:8-alpine"

	VolumePostgres = "vitek_pg"

	// Ports the containers listen on internally (image defaults).
	ContainerPostgresPort = "5432"
	ContainerRedisPort    = "6379"
)
