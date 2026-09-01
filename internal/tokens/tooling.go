package tokens

import (
	"fmt"
	"strings"
)

// Toolchain and codegen path tokens.
const (
	GoToolchain = "1.26.0"

	SQLCVersion       = "v1.31.1"
	PackageRepository = "repository"
	SQLEnginePostgres = "postgresql"
	SQLPackagePgx     = "pgx/v5"

	PathMigrations        = "db/migrations"
	PathMigrationInit     = "db/migrations/000001_init.up.sql"
	PathMigrationPlatform = "db/migrations/000002_platform.up.sql"
	PathMigrationSessions     = "db/migrations/000003_sessions.up.sql"
	PathMigrationListingSearch = "db/migrations/000004_listing_search.up.sql"
	PathMigrationAvitoSecrets  = "db/migrations/000005_avito_secrets.up.sql"
	PathMigrationListingFilterSeen = "db/migrations/000006_listing_filter_seen.up.sql"
	PathQueries           = "db/queries"
	PathRepository        = "internal/repository"
	PathSQLC              = "sqlc.yaml"
	PathCompose           = "docker-compose.yml"
	PathEnvExample        = ".env.example"
	PathDesignCSS         = "web/tokens.css"
	PathAppFace           = "web/app/face.html"
	PathGoMod             = "go.mod"
	PathREADME            = "README.md"
	PathDockerfile        = "Dockerfile"
	PathDEPLOY            = "DEPLOY.md"
	PathTaskfile          = "Taskfile.yml"
	PathLRTRule           = ".cursor/rules/lrt.mdc"
	PathVitekPolicyRule   = ".cursor/rules/vitek-policy.mdc"
)

// ImageGoBuild returns the golang build image derived from GoToolchain (major.minor).
func ImageGoBuild() string {
	parts := strings.Split(GoToolchain, ".")
	if len(parts) < 2 {
		panic("tokens: GoToolchain must be major.minor.patch")
	}
	return fmt.Sprintf("golang:%s.%s-alpine", parts[0], parts[1])
}

// RenderSQLCYAML returns the canonical sqlc.yaml body.
// Schema points at *.up.sql only so down migrations are not parsed.
func RenderSQLCYAML() string {
	return "" +
		"version: \"2\"\n" +
		"sql:\n" +
		"  - engine: \"" + SQLEnginePostgres + "\"\n" +
		"    queries: \"" + PathQueries + "\"\n" +
		"    schema:\n" +
		"      - \"" + PathMigrationInit + "\"\n" +
		"      - \"" + PathMigrationPlatform + "\"\n" +
		"      - \"" + PathMigrationSessions + "\"\n" +
		"      - \"" + PathMigrationListingSearch + "\"\n" +
		"      - \"" + PathMigrationAvitoSecrets + "\"\n" +
		"      - \"" + PathMigrationListingFilterSeen + "\"\n" +
		"    gen:\n" +
		"      go:\n" +
		"        package: \"" + PackageRepository + "\"\n" +
		"        out: \"" + PathRepository + "\"\n" +
		"        sql_package: \"" + SQLPackagePgx + "\"\n" +
		"        emit_json_tags: true\n" +
		"        emit_empty_slices: true\n" +
		"        emit_pointers_for_null_types: true\n"
}
