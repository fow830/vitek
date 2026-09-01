package contracts_test

import (
	"testing"

	"vitek/internal/tokens"
)

func TestContract_DesignTokensCSSIsRenderedFromTokens(t *testing.T) {
	assertFileEqualsRender(t, tokens.PathDesignCSS, tokens.RenderDesignCSS())
}
