package tokens

// Infrastructure image and volume tokens (docker compose / CI).
const (
	ImagePostgres = "postgres:17-alpine"
	ImageRedis    = "redis:8-alpine"
	ImageGoBuild  = "golang:1.26-alpine"
	ImageRuntime  = "gcr.io/distroless/static-debian12"

	VolumePostgres = "vitek_pg"

	// Ports the containers listen on internally (image defaults).
	ContainerPostgresPort = "5432"
	ContainerRedisPort    = "6379"

	DefaultWorkerTick = "5s"
)
