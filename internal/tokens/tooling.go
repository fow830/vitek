package tokens

// Toolchain and codegen path tokens.
const (
	GoToolchain = "1.26.0"

	PathMigrations     = "db/migrations"
	PathMigrationInit  = "db/migrations/000001_init.up.sql"
	PathQueries        = "db/queries"
	PathRepository     = "internal/repository"
	PathSQLC           = "sqlc.yaml"
	PathCompose        = "docker-compose.yml"
	PathEnvExample     = ".env.example"
	PathDesignCSS      = "web/tokens.css"
	PathGoMod          = "go.mod"
	PathREADME         = "README.md"
	PathDockerfile     = "Dockerfile"
	PathDEPLOY         = "DEPLOY.md"
)

// RenderSQLCYAML returns the canonical sqlc.yaml body.
// Schema points at *.up.sql only so down migrations are not parsed.
func RenderSQLCYAML() string {
	return "" +
		"version: \"2\"\n" +
		"sql:\n" +
		"  - engine: \"postgresql\"\n" +
		"    queries: \"" + PathQueries + "\"\n" +
		"    schema:\n" +
		"      - \"" + PathMigrationInit + "\"\n" +
		"    gen:\n" +
		"      go:\n" +
		"        package: \"repository\"\n" +
		"        out: \"" + PathRepository + "\"\n" +
		"        sql_package: \"pgx/v5\"\n" +
		"        emit_json_tags: true\n" +
		"        emit_empty_slices: true\n" +
		"        emit_pointers_for_null_types: true\n"
}
