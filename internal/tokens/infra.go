package tokens

// Infrastructure image and volume tokens (docker compose / CI).
const (
	ImageRuntime       = "gcr.io/distroless/static-debian12"
	ImageWorkerRuntime = "chromedp/headless-shell:latest"

	VolumePostgres = "vitek_pg"

	ContainerPostgresPort = "5432"

	DefaultWorkerTick = "5s"

	DefaultListingSearchProcessor = ListingSearchProcessorStub
)

// ImagePostgres is the compose/testcontainers Postgres image.
const ImagePostgres = "postgres:17-alpine"
