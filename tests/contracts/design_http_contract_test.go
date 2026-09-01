package contracts_test

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"

	"vitek/internal/tokens"
	httpapi "vitek/internal/transport/httpapi"
)

// CONTRACT-DESIGN-HTTP-001: design tokens CSS is served from token render (not disk).
func TestContract_DesignCSSHTTPSurface(t *testing.T) {
	pool, _ := queries(t)
	handler := httpapi.NewServer(pool).Handler()

	req := httptest.NewRequest(http.MethodGet, tokens.PathTokensCSS, nil)
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)
	require.Equal(t, http.StatusOK, rec.Code)
	require.Equal(t, tokens.MIMETextCSS, rec.Header().Get(tokens.HeaderContentType))

	want := tokens.RenderDesignCSS()
	require.Equal(t, want, rec.Body.String())
	require.Contains(t, want, tokens.CSSColorAccent+": "+tokens.ColorAccent)

	app := appHostRequest(http.MethodGet, tokens.PathRoot)
	appRec := httptest.NewRecorder()
	handler.ServeHTTP(appRec, app)
	require.Equal(t, http.StatusOK, appRec.Code)
	require.Contains(t, appRec.Body.String(), `href="`+tokens.PathTokensCSS+`"`)
	require.Contains(t, strings.ToLower(appRec.Body.String()), tokens.CSSColorAccent)
}
