package contracts_test

import (
	"html"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"

	"vitek/internal/tokens"
)

// CONTRACT-ADMIN-FACE-001: face.html is generated from tokens (no hand-edited SoT).
func TestContract_AdminFaceIsRenderedFromTokens(t *testing.T) {
	assertFileEqualsRender(t, tokens.PathAdminFace, tokens.RenderAdminFaceHTML())
}

// CONTRACT-ADMIN-FACE-002: brand parts, catalog, nav, auth method appear in render.
func TestContract_AdminFaceEmbedsTokenIdentity(t *testing.T) {
	body := tokens.RenderAdminFaceHTML()

	require.Equal(t, tokens.ProductNameLocal, tokens.ProductBrandStem()+tokens.ProductBrandAccent())
	require.Contains(t, body, tokens.ProductBrandStem())
	require.Contains(t, body, tokens.ProductBrandAccent())
	require.Contains(t, body, tokens.AuthMethodMagicLink)
	require.Contains(t, body, tokens.FixtureAdminEmail())
	require.Contains(t, body, html.EscapeString(tokens.FontsGoogleCSSURL))
	require.Contains(t, body, tokens.PathAdminFaceTokensHref)
	require.Contains(t, body, tokens.AdminMockClockLabel)
	require.NotContains(t, strings.ToLower(body), "datastar")
	require.NotContains(t, body, "LIVE SSE")

	for _, s := range tokens.ProductServiceCatalog {
		require.Contains(t, body, s.Code)
		require.Contains(t, body, s.Title)
	}
	for _, n := range tokens.AdminNav {
		require.Contains(t, body, n.Label)
		require.Contains(t, body, `data-view="`+n.ID+`"`)
	}
}

// CONTRACT-ADMIN-FACE-003: design CSS carries admin chrome color tokens.
func TestContract_DesignCSSIncludesAdminChromeColors(t *testing.T) {
	css := tokens.RenderDesignCSS()
	require.Contains(t, css, tokens.CSSColorOnAccent+": "+tokens.ColorOnAccent)
	require.Contains(t, css, tokens.CSSColorCanvasHi+": "+tokens.ColorCanvasHi)
	require.Contains(t, css, tokens.CSSColorCanvasLo+": "+tokens.ColorCanvasLo)
}
