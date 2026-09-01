package contracts_test

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"

	"vitek/internal/repository"
	"vitek/internal/service"
	"vitek/internal/tokens"
	httpapi "vitek/internal/transport/httpapi"
)

// CONTRACT-AUTH-ML-008: clickable open URL consumes token, sets session, redirects to app /.
func TestContract_MagicLinkHTTP_OpenURLLogin(t *testing.T) {
	pool, q := queries(t)
	mailer := service.NewMemoryMagicLinkMailer()
	handler := httpapi.NewServer(pool, httpapi.WithMagicLinkMailer(mailer)).Handler()
	ctx := t.Context()

	_, err := q.CreateUserWithRole(ctx, repository.CreateUserWithRoleParams{
		Email: tokens.ProductEmail("open-admin"),
		Role:  repository.UserRoleADMIN,
	})
	require.NoError(t, err)

	body, err := json.Marshal(map[string]any{tokens.JSONFieldEmail: tokens.ProductEmail("open-admin")})
	require.NoError(t, err)
	req := withAppHost(httptest.NewRequest(http.MethodPost, tokens.PathV1AuthMagicLink, bytes.NewReader(body)))
	req.Header.Set(tokens.HeaderContentType, tokens.MIMEApplicationJSON)
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)
	require.Equal(t, http.StatusAccepted, rec.Code)
	require.NotEmpty(t, mailer.LastToken)

	openPath := tokens.PathV1AuthMagicLinkOpen + "?" + tokens.QueryParamToken + "=" + mailer.LastToken
	openURL := tokens.MagicLinkOpenURL(mailer.LastToken)
	require.True(t, strings.HasPrefix(openURL, tokens.HTTPSAppBase()))
	require.Contains(t, openURL, tokens.PathV1AuthMagicLinkOpen)

	oreq := withAppHost(httptest.NewRequest(http.MethodGet, openPath, nil))
	orec := httptest.NewRecorder()
	handler.ServeHTTP(orec, oreq)
	require.Equal(t, http.StatusFound, orec.Code)
	require.Equal(t, tokens.PathRoot, orec.Header().Get("Location"))
	require.True(t, strings.Contains(orec.Header().Get(tokens.HeaderSetCookie), tokens.CookieSessionName+"="))

	var cookie string
	for _, c := range orec.Result().Cookies() {
		if c.Name == tokens.CookieSessionName {
			cookie = c.Value
		}
	}
	require.NotEmpty(t, cookie)

	root := appHostRequest(http.MethodGet, tokens.PathRoot)
	root.Header.Set(tokens.HeaderCookie, tokens.CookieSessionName+"="+cookie)
	rrec := httptest.NewRecorder()
	handler.ServeHTTP(rrec, root)
	require.Equal(t, http.StatusOK, rrec.Code)
	require.Contains(t, rrec.Body.String(), tokens.AppDOMScreenPlatform)
	require.Contains(t, rrec.Body.String(), tokens.ProductEmail("open-admin"))

	// token is single-use
	oreq2 := withAppHost(httptest.NewRequest(http.MethodGet, openPath, nil))
	orec2 := httptest.NewRecorder()
	handler.ServeHTTP(orec2, oreq2)
	require.Equal(t, http.StatusUnauthorized, orec2.Code)
	require.Contains(t, orec2.Body.String(), tokens.MagicLinkOpenCopyInvalid)
}

// CONTRACT-AUTH-ML-009: MagicLinkOpenURL is the mailer SoT for clickable links.
func TestContract_MagicLinkOpenURL(t *testing.T) {
	raw := "deadbeef"
	url := tokens.MagicLinkOpenURL(raw)
	require.Equal(t, tokens.HTTPSAppBase()+tokens.PathV1AuthMagicLinkOpen+"?"+tokens.QueryParamToken+"="+raw, url)
}

// CONTRACT-AUTH-ML-010: missing/invalid token on GET open returns HTML error (not JSON).
func TestContract_MagicLinkHTTP_OpenInvalid(t *testing.T) {
	pool, _ := queries(t)
	handler := httpapi.NewServer(pool).Handler()

	req := withAppHost(httptest.NewRequest(http.MethodGet, tokens.PathV1AuthMagicLinkOpen, nil))
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)
	require.Equal(t, http.StatusBadRequest, rec.Code)
	require.Contains(t, rec.Header().Get(tokens.HeaderContentType), tokens.MIMETextHTML)
	require.Contains(t, rec.Body.String(), tokens.MagicLinkOpenCopyMissingToken)

	req = withAppHost(httptest.NewRequest(http.MethodGet, tokens.PathV1AuthMagicLinkOpen+"?"+tokens.QueryParamToken+"=not-a-valid-token", nil))
	rec = httptest.NewRecorder()
	handler.ServeHTTP(rec, req)
	require.Equal(t, http.StatusUnauthorized, rec.Code)
	require.Contains(t, rec.Header().Get(tokens.HeaderContentType), tokens.MIMETextHTML)
	require.Contains(t, rec.Body.String(), tokens.MagicLinkOpenCopyInvalid)
}

// CONTRACT-AUTH-ML-011: open path is on HTTP allowlist (app host only).
func TestContract_MagicLinkHTTP_OpenOnAllowlist(t *testing.T) {
	require.Contains(t, tokens.HTTPPathAllowlist, tokens.PathV1AuthMagicLinkOpen)

	pool, _ := queries(t)
	handler := httpapi.NewServer(pool).Handler()

	req := withAppHost(httptest.NewRequest(http.MethodGet, tokens.PathV1AuthMagicLinkOpen+"?"+tokens.QueryParamToken+"=x", nil))
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)
	require.NotEqual(t, http.StatusNotFound, rec.Code)

	req = landingHostRequest(http.MethodGet, tokens.PathV1AuthMagicLinkOpen+"?"+tokens.QueryParamToken+"=x")
	rec = httptest.NewRecorder()
	handler.ServeHTTP(rec, req)
	require.Equal(t, http.StatusNotFound, rec.Code)
}

// CONTRACT-AUTH-ML-012: expose opt-in returns magic_link_url (full open URL) for local shell.
func TestContract_MagicLinkHTTP_ExposeOpenURLOptIn(t *testing.T) {
	pool, _ := queries(t)
	mailer := service.NewMemoryMagicLinkMailer()
	handler := httpapi.NewServer(pool,
		httpapi.WithMagicLinkMailer(mailer),
		httpapi.WithExposeMagicLinkTokens(true),
	).Handler()

	body, err := json.Marshal(map[string]any{tokens.JSONFieldEmail: tokens.ProductEmail("expose-url")})
	require.NoError(t, err)
	req := withAppHost(httptest.NewRequest(http.MethodPost, tokens.PathV1AuthMagicLink, bytes.NewReader(body)))
	req.Header.Set(tokens.HeaderContentType, tokens.MIMEApplicationJSON)
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)
	require.Equal(t, http.StatusAccepted, rec.Code)

	var out map[string]any
	require.NoError(t, json.NewDecoder(rec.Body).Decode(&out))
	require.Equal(t, tokens.MagicLinkOpenURL(mailer.LastToken), out[tokens.JSONFieldMagicLinkURL])
}
