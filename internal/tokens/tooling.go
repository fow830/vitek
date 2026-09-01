package tokens

// Toolchain and codegen path tokens.
const (
	GoToolchain = "1.26.0"

	PathMigrations = "db/migrations"
	PathQueries    = "db/queries"
	PathRepository = "internal/repository"
	PathSQLC       = "sqlc.yaml"
	PathCompose    = "docker-compose.yml"
	PathEnvExample = ".env.example"
	PathDesignCSS  = "web/tokens.css"
	PathGoMod      = "go.mod"
	PathREADME     = "README.md"
)

// RenderSQLCYAML returns the canonical sqlc.yaml body.
func RenderSQLCYAML() string {
	return "" +
		"version: \"2\"\n" +
		"sql:\n" +
		"  - engine: \"postgresql\"\n" +
		"    queries: \"" + PathQueries + "\"\n" +
		"    schema: \"" + PathMigrations + "\"\n" +
		"    gen:\n" +
		"      go:\n" +
		"        package: \"repository\"\n" +
		"        out: \"" + PathRepository + "\"\n" +
		"        sql_package: \"pgx/v5\"\n" +
		"        emit_json_tags: true\n" +
		"        emit_empty_slices: true\n"
}
