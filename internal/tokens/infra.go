package tokens

// Infrastructure image and volume tokens (docker compose / CI).
const (
	ImageRedis   = "redis:8-alpine"
	ImageRuntime = "gcr.io/distroless/static-debian12"

	VolumePostgres = "vitek_pg"

	// Ports the containers listen on internally (image defaults).
	ContainerPostgresPort = "5432"
	ContainerRedisPort    = "6379"

	DefaultWorkerTick = "5s"
)

// ImagePostgres is the compose/testcontainers Postgres image.
const ImagePostgres = "postgres:17-alpine"

// ImageTagLocal builds local docker tags: vitek-api:local.
func ImageTagLocal(binary string) string {
	return ModulePath + "-" + binary + ":local"
}
