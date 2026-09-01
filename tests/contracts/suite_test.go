package contracts_test

import (
	"context"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"testing"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/stretchr/testify/require"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/modules/postgres"
	"github.com/testcontainers/testcontainers-go/wait"

	"vitek/internal/repository"
	"vitek/internal/tokens"
)

func setupTestDB(t *testing.T) *pgxpool.Pool {
	t.Helper()
	ctx := context.Background()

	pgContainer, err := postgres.Run(ctx,
		tokens.ImagePostgres,
		postgres.WithDatabase(tokens.DefaultTestPostgresDB),
		postgres.WithUsername(tokens.DefaultPostgresUser),
		postgres.WithPassword(tokens.DefaultPostgresPassword),
		testcontainers.WithWaitStrategy(
			wait.ForLog("database system is ready to accept connections").
				WithOccurrence(2).
				WithStartupTimeout(60*time.Second),
		),
	)
	require.NoError(t, err)

	t.Cleanup(func() {
		require.NoError(t, pgContainer.Terminate(ctx))
	})

	connStr, err := pgContainer.ConnectionString(ctx, "sslmode="+tokens.DefaultPostgresSSLMode)
	require.NoError(t, err)

	pool, err := pgxpool.New(ctx, connStr)
	require.NoError(t, err)
	t.Cleanup(pool.Close)

	runMigrations(t, pool)
	return pool
}

func runMigrations(t *testing.T, pool *pgxpool.Pool) {
	t.Helper()
	root := moduleRoot(t)
	dir := filepath.Join(root, tokens.PathMigrations)

	entries, err := os.ReadDir(dir)
	require.NoError(t, err)

	var ups []string
	for _, e := range entries {
		name := e.Name()
		if strings.HasSuffix(name, ".up.sql") {
			ups = append(ups, filepath.Join(dir, name))
		}
	}
	sort.Strings(ups)
	require.NotEmpty(t, ups, "no *.up.sql migrations found")

	ctx := context.Background()
	for _, path := range ups {
		sqlBytes, err := os.ReadFile(path)
		require.NoError(t, err)
		_, err = pool.Exec(ctx, string(sqlBytes))
		require.NoError(t, err, "migration %s", path)
	}
}

func queries(t *testing.T) (*pgxpool.Pool, *repository.Queries) {
	t.Helper()
	pool := setupTestDB(t)
	return pool, repository.New(pool)
}
