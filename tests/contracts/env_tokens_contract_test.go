package contracts_test

import (
	"testing"

	"vitek/internal/tokens"
)

func TestContract_EnvExampleIsRenderedFromTokens(t *testing.T) {
	assertFileEqualsRender(t, tokens.PathEnvExample, tokens.RenderEnvExample())
}
