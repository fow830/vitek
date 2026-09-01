package contracts_test

import (
	"testing"

	"github.com/stretchr/testify/require"

	"vitek/internal/tokens"
)

func TestContract_EnvExampleIsRenderedFromTokens(t *testing.T) {
	assertFileEqualsRender(t, tokens.PathEnvExample, tokens.RenderEnvExample())
}

func TestContract_EnvExampleOmitsDeadHTTPPort(t *testing.T) {
	for _, k := range tokens.EnvExampleKeys {
		require.NotEqual(t, "HTTP_PORT", k)
	}
	require.NotContains(t, tokens.RenderEnvExample(), "HTTP_PORT=")
}
