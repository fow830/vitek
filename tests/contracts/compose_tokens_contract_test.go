package contracts_test

import (
	"os"
	"path/filepath"
	"strings"
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
	require.Contains(t, string(raw), tokens.ProductDomain)
	require.Contains(t, string(raw), tokens.ProductDomainApp)
}

func TestContract_DEPLOYMentionsProductionDomainTokens(t *testing.T) {
	root := moduleRoot(t)
	raw, err := os.ReadFile(filepath.Join(root, tokens.PathDEPLOY))
	require.NoError(t, err)
	body := string(raw)
	require.Contains(t, body, tokens.ProductDomain)
	require.Contains(t, body, tokens.ProductDomainWWW)
	require.Contains(t, body, tokens.ProductDomainApp)
	require.Contains(t, body, tokens.ComposeContainerPostgresProd)
	require.Contains(t, body, tokens.ComposeNetworkProd)
}

func TestContract_TaskfilePinsSQLCVersion(t *testing.T) {
	root := moduleRoot(t)
	raw, err := os.ReadFile(filepath.Join(root, tokens.PathTaskfile))
	require.NoError(t, err)
	require.Contains(t, string(raw), "sqlc@"+tokens.SQLCVersion)
}

func TestContract_GoBuildImageMatchesToolchain(t *testing.T) {
	parts := strings.Split(tokens.GoToolchain, ".")
	require.GreaterOrEqual(t, len(parts), 2)
	want := "golang:" + parts[0] + "." + parts[1] + "-alpine"
	require.Equal(t, want, tokens.ImageGoBuild())
}

func assertFileEqualsRender(t *testing.T, rel, want string) {
	t.Helper()
	root := moduleRoot(t)
	raw, err := os.ReadFile(filepath.Join(root, rel))
	require.NoError(t, err)
	require.Equal(t, want, string(raw), rel)
}
