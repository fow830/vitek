package contracts_test

import (
	"html"
	"strconv"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"

	"vitek/internal/tokens"
)

// CONTRACT-APP-FACE-001: app/face.html is generated from tokens (no hand-edited SoT).
func TestContract_AppFaceIsRenderedFromTokens(t *testing.T) {
	assertFileEqualsRender(t, tokens.PathAppFace, tokens.RenderAppFaceHTML())
}

// CONTRACT-APP-FACE-002: brand parts, catalog, nav, auth method appear in render.
func TestContract_AppFaceEmbedsTokenIdentity(t *testing.T) {
	body := tokens.RenderAppFaceHTML()

	require.Equal(t, tokens.ProductNameLocal, tokens.ProductBrandStem()+tokens.ProductBrandAccent())
	require.Contains(t, body, tokens.ProductBrandStem())
	require.Contains(t, body, tokens.ProductBrandAccent())
	require.Contains(t, body, tokens.AuthMethodMagicLink)
	require.Contains(t, body, tokens.FixtureSessionEmail())
	require.Contains(t, body, html.EscapeString(tokens.FontsGoogleCSSURL))
	require.Contains(t, body, tokens.PathTokensCSS)
	require.Contains(t, body, tokens.AppCopyLiveClock)
	require.Contains(t, strings.ToLower(body), tokens.AttrDataStar)
	require.Contains(t, body, tokens.DatastarCDNURL)
	require.NotContains(t, body, "LIVE SSE")

	for _, s := range tokens.ProductServiceCatalog {
		require.Contains(t, body, s.Code)
		require.Contains(t, body, s.Title)
	}
	for _, n := range tokens.AppNav {
		require.Contains(t, body, n.Label)
		require.Contains(t, body, `data-view="`+n.ID+`"`)
		require.Contains(t, body, tokens.AppDOMViewPrefix+n.ID)
	}
	require.Contains(t, body, tokens.PathV1AuthLogout)
	require.Contains(t, body, tokens.AppClassIsActive)
}

// CONTRACT-APP-FACE-004: generated JS uses tokenized auth open URL + task terminal statuses.
func TestContract_AppFaceScriptUsesTokens(t *testing.T) {
	body := tokens.RenderAppFaceHTML()
	require.Contains(t, body, tokens.JSONFieldMagicLinkURL)

	loggedIn := tokens.RenderAppFaceHTMLLoggedIn(tokens.ProductEmail("face"), "", false)
	require.Contains(t, loggedIn, "'"+tokens.TaskStatusCompleted+"'")
	require.Contains(t, loggedIn, "'"+tokens.TaskStatusFailed+"'")
	require.Contains(t, loggedIn, strconv.Itoa(tokens.ListingSearchPollMaxAttempts))
}

// CONTRACT-APP-FACE-003: design CSS carries app chrome color tokens.
func TestContract_DesignCSSIncludesAppChromeColors(t *testing.T) {
	css := tokens.RenderDesignCSS()
	require.Contains(t, css, tokens.CSSColorOnAccent+": "+tokens.ColorOnAccent)
	require.Contains(t, css, tokens.CSSColorCanvasHi+": "+tokens.ColorCanvasHi)
	require.Contains(t, css, tokens.CSSColorCanvasLo+": "+tokens.ColorCanvasLo)
}

// CONTRACT-LANDING-001: landing HTML is token-only and has no platform shell.
func TestContract_LandingHTMLFromTokens(t *testing.T) {
	body := tokens.RenderLandingHTML()
	require.Contains(t, body, tokens.LandingDOMMagicForm)
	require.Contains(t, body, tokens.ProductDomainApp)
	require.NotContains(t, body, tokens.AppDOMScreenPlatform)
	require.NotContains(t, body, tokens.AppNavIDOverview)
}
