package contracts_test

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/require"

	"vitek/internal/tokens"
)

func TestContract_ComposeIsRenderedFromTokens(t *testing.T) {
	assertFileEqualsRender(t, tokens.PathCompose, tokens.RenderComposeYAML())
}

func TestContract_SQLCIsRenderedFromTokens(t *testing.T) {
	assertFileEqualsRender(t, tokens.PathSQLC, tokens.RenderSQLCYAML())
}

func TestContract_DockerfileIsRenderedFromTokens(t *testing.T) {
	assertFileEqualsRender(t, tokens.PathDockerfile, tokens.RenderDockerfile())
}

func TestContract_GoModToolchainMatchesToken(t *testing.T) {
	root := moduleRoot(t)
	raw, err := os.ReadFile(filepath.Join(root, tokens.PathGoMod))
	require.NoError(t, err)
	require.Contains(t, string(raw), "go "+tokens.GoToolchain)
	require.Contains(t, string(raw), "module "+tokens.ModulePath)
}

func TestContract_READMETitleMatchesProduct(t *testing.T) {
	root := moduleRoot(t)
	raw, err := os.ReadFile(filepath.Join(root, tokens.PathREADME))
	require.NoError(t, err)
	require.Contains(t, string(raw), "# "+tokens.ProductName)
	require.Contains(t, string(raw), tokens.ProductNameLocal)
}

func TestContract_TaskfilePinsSQLCVersion(t *testing.T) {
	root := moduleRoot(t)
	raw, err := os.ReadFile(filepath.Join(root, tokens.PathTaskfile))
	require.NoError(t, err)
	require.Contains(t, string(raw), "sqlc@"+tokens.SQLCVersion)
}

func TestContract_GoBuildImageMatchesToolchain(t *testing.T) {
	require.Equal(t, "golang:1.26-alpine", tokens.ImageGoBuild())
	require.Contains(t, tokens.RenderDockerfile(), "FROM "+tokens.ImageGoBuild()+" AS build")
}

func assertFileEqualsRender(t *testing.T, rel, want string) {
	t.Helper()
	root := moduleRoot(t)
	raw, err := os.ReadFile(filepath.Join(root, rel))
	require.NoError(t, err)
	require.Equal(t, want, string(raw), rel)
}
