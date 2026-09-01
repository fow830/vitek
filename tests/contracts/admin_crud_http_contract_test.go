package contracts_test

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/require"

	"vitek/internal/repository"
	"vitek/internal/service"
	"vitek/internal/tokens"
	httpapi "vitek/internal/transport/httpapi"
)

// CONTRACT-ADMIN-CRUD-001: admin can CRUD proxies; non-admin forbidden; anon unauthorized.
func TestContract_AdminProxiesCRUD(t *testing.T) {
	pool, _ := queries(t)
	mailer := service.NewMemoryMagicLinkMailer()
	handler := httpapi.NewServer(pool, httpapi.WithMagicLinkMailer(mailer)).Handler()
	ctx := t.Context()
	q := repository.New(pool)

	_, err := q.CreateUserWithRole(ctx, repository.CreateUserWithRoleParams{
		Email: tokens.ProductEmail("crud-admin"),
		Role:  repository.UserRoleADMIN,
	})
	require.NoError(t, err)
	_, err = q.CreateUserWithRole(ctx, repository.CreateUserWithRoleParams{
		Email: tokens.ProductEmail("crud-user"),
		Role:  repository.UserRoleUSER,
	})
	require.NoError(t, err)

	adminCookie := loginCookie(t, handler, mailer, tokens.ProductEmail("crud-admin"))
	userCookie := loginCookie(t, handler, mailer, tokens.ProductEmail("crud-user"))

	// anon
	req := withAppHost(httptest.NewRequest(http.MethodGet, tokens.PathV1AdminProxies, nil))
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)
	require.Equal(t, http.StatusUnauthorized, rec.Code)

	// user forbidden
	req = withAppHost(httptest.NewRequest(http.MethodGet, tokens.PathV1AdminProxies, nil))
	req.Header.Set(tokens.HeaderCookie, tokens.CookieSessionName+"="+userCookie)
	rec = httptest.NewRecorder()
	handler.ServeHTTP(rec, req)
	require.Equal(t, http.StatusForbidden, rec.Code)

	// admin create
	payload, err := json.Marshal(map[string]any{
		tokens.JSONFieldEndpoint: tokens.FixtureAdminProxyEndpoint,
		tokens.JSONFieldStatus:   string(repository.ProxyStatusACTIVE),
		tokens.JSONFieldLabel:    "msk",
	})
	require.NoError(t, err)
	req = withAppHost(httptest.NewRequest(http.MethodPost, tokens.PathV1AdminProxies, bytes.NewReader(payload)))
	req.Header.Set(tokens.HeaderContentType, tokens.MIMEApplicationJSON)
	req.Header.Set(tokens.HeaderCookie, tokens.CookieSessionName+"="+adminCookie)
	rec = httptest.NewRecorder()
	handler.ServeHTTP(rec, req)
	require.Equal(t, http.StatusCreated, rec.Code)

	var created map[string]any
	require.NoError(t, json.NewDecoder(rec.Body).Decode(&created))
	id := created[tokens.JSONFieldID].(string)

	// admin list
	req = withAppHost(httptest.NewRequest(http.MethodGet, tokens.PathV1AdminProxies, nil))
	req.Header.Set(tokens.HeaderCookie, tokens.CookieSessionName+"="+adminCookie)
	rec = httptest.NewRecorder()
	handler.ServeHTTP(rec, req)
	require.Equal(t, http.StatusOK, rec.Code)
	var list map[string]any
	require.NoError(t, json.NewDecoder(rec.Body).Decode(&list))
	require.Len(t, list[tokens.JSONFieldProxies].([]any), 1)

	// admin patch
	patch, err := json.Marshal(map[string]any{
		tokens.JSONFieldEndpoint: tokens.FixtureAdminProxyEndpoint,
		tokens.JSONFieldStatus:   string(repository.ProxyStatusDISABLED),
		tokens.JSONFieldLabel:    "msk-off",
	})
	require.NoError(t, err)
	req = withAppHost(httptest.NewRequest(http.MethodPatch, tokens.HTTPResourceID(tokens.PathV1AdminProxies, id), bytes.NewReader(patch)))
	req.Header.Set(tokens.HeaderContentType, tokens.MIMEApplicationJSON)
	req.Header.Set(tokens.HeaderCookie, tokens.CookieSessionName+"="+adminCookie)
	rec = httptest.NewRecorder()
	handler.ServeHTTP(rec, req)
	require.Equal(t, http.StatusOK, rec.Code)
	var patched map[string]any
	require.NoError(t, json.NewDecoder(rec.Body).Decode(&patched))
	require.Equal(t, string(repository.ProxyStatusDISABLED), patched[tokens.JSONFieldStatus])
	require.Equal(t, "msk-off", patched[tokens.JSONFieldLabel])

	bad, err := json.Marshal(map[string]any{
		tokens.JSONFieldEndpoint: tokens.FixtureAdminProxyEndpoint,
		tokens.JSONFieldStatus:   tokens.FixtureInvalidEnum,
		tokens.JSONFieldLabel:    "x",
	})
	require.NoError(t, err)
	req = withAppHost(httptest.NewRequest(http.MethodPost, tokens.PathV1AdminProxies, bytes.NewReader(bad)))
	req.Header.Set(tokens.HeaderContentType, tokens.MIMEApplicationJSON)
	req.Header.Set(tokens.HeaderCookie, tokens.CookieSessionName+"="+adminCookie)
	rec = httptest.NewRecorder()
	handler.ServeHTTP(rec, req)
	require.Equal(t, http.StatusBadRequest, rec.Code)
}

// CONTRACT-ADMIN-CRUD-002: admin CRUD avito accounts.
func TestContract_AdminAvitoAccountsCRUD(t *testing.T) {
	pool, _ := queries(t)
	mailer := service.NewMemoryMagicLinkMailer()
	handler := httpapi.NewServer(pool, httpapi.WithMagicLinkMailer(mailer)).Handler()
	ctx := t.Context()
	q := repository.New(pool)

	_, err := q.CreateUserWithRole(ctx, repository.CreateUserWithRoleParams{
		Email: tokens.ProductEmail("avito-admin"),
		Role:  repository.UserRoleADMIN,
	})
	require.NoError(t, err)
	cookie := loginCookie(t, handler, mailer, tokens.ProductEmail("avito-admin"))

	payload, err := json.Marshal(map[string]any{
		tokens.JSONFieldLabel:       "pool-1",
		tokens.JSONFieldStatus:      string(repository.AvitoAccountStatusACTIVE),
		tokens.JSONFieldExternalRef: "ref-1",
	})
	require.NoError(t, err)
	req := withAppHost(httptest.NewRequest(http.MethodPost, tokens.PathV1AdminAvitoAccounts, bytes.NewReader(payload)))
	req.Header.Set(tokens.HeaderContentType, tokens.MIMEApplicationJSON)
	req.Header.Set(tokens.HeaderCookie, tokens.CookieSessionName+"="+cookie)
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)
	require.Equal(t, http.StatusCreated, rec.Code)

	req = withAppHost(httptest.NewRequest(http.MethodGet, tokens.PathV1AdminAvitoAccounts, nil))
	req.Header.Set(tokens.HeaderCookie, tokens.CookieSessionName+"="+cookie)
	rec = httptest.NewRecorder()
	handler.ServeHTTP(rec, req)
	require.Equal(t, http.StatusOK, rec.Code)
	var list map[string]any
	require.NoError(t, json.NewDecoder(rec.Body).Decode(&list))
	accounts := list[tokens.JSONFieldAccounts].([]any)
	require.Len(t, accounts, 1)
	id := accounts[0].(map[string]any)[tokens.JSONFieldID].(string)

	patch, err := json.Marshal(map[string]any{
		tokens.JSONFieldLabel:       "pool-1b",
		tokens.JSONFieldStatus:      string(repository.AvitoAccountStatusDISABLED),
		tokens.JSONFieldExternalRef: "ref-1b",
	})
	require.NoError(t, err)
	req = withAppHost(httptest.NewRequest(http.MethodPatch, tokens.HTTPResourceID(tokens.PathV1AdminAvitoAccounts, id), bytes.NewReader(patch)))
	req.Header.Set(tokens.HeaderContentType, tokens.MIMEApplicationJSON)
	req.Header.Set(tokens.HeaderCookie, tokens.CookieSessionName+"="+cookie)
	rec = httptest.NewRecorder()
	handler.ServeHTTP(rec, req)
	require.Equal(t, http.StatusOK, rec.Code)
	var patched map[string]any
	require.NoError(t, json.NewDecoder(rec.Body).Decode(&patched))
	require.Equal(t, string(repository.AvitoAccountStatusDISABLED), patched[tokens.JSONFieldStatus])
	require.Equal(t, "pool-1b", patched[tokens.JSONFieldLabel])

	bad, err := json.Marshal(map[string]any{
		tokens.JSONFieldLabel:       "x",
		tokens.JSONFieldStatus:      tokens.FixtureInvalidEnum,
		tokens.JSONFieldExternalRef: "r",
	})
	require.NoError(t, err)
	req = withAppHost(httptest.NewRequest(http.MethodPost, tokens.PathV1AdminAvitoAccounts, bytes.NewReader(bad)))
	req.Header.Set(tokens.HeaderContentType, tokens.MIMEApplicationJSON)
	req.Header.Set(tokens.HeaderCookie, tokens.CookieSessionName+"="+cookie)
	rec = httptest.NewRecorder()
	handler.ServeHTTP(rec, req)
	require.Equal(t, http.StatusBadRequest, rec.Code)
}

func loginCookie(t *testing.T, handler http.Handler, mailer *service.MemoryMagicLinkMailer, email string) string {
	t.Helper()
	body, err := json.Marshal(map[string]any{tokens.JSONFieldEmail: email})
	require.NoError(t, err)
	req := withAppHost(httptest.NewRequest(http.MethodPost, tokens.PathV1AuthMagicLink, bytes.NewReader(body)))
	req.Header.Set(tokens.HeaderContentType, tokens.MIMEApplicationJSON)
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)
	require.Equal(t, http.StatusAccepted, rec.Code)
	require.NotEmpty(t, mailer.LastToken)

	cbody, err := json.Marshal(map[string]any{tokens.JSONFieldToken: mailer.LastToken})
	require.NoError(t, err)
	creq := withAppHost(httptest.NewRequest(http.MethodPost, tokens.PathV1AuthMagicLinkConsume, bytes.NewReader(cbody)))
	creq.Header.Set(tokens.HeaderContentType, tokens.MIMEApplicationJSON)
	crec := httptest.NewRecorder()
	handler.ServeHTTP(crec, creq)
	require.Equal(t, http.StatusOK, crec.Code)

	for _, c := range crec.Result().Cookies() {
		if c.Name == tokens.CookieSessionName {
			return c.Value
		}
	}
	t.Fatal("session cookie missing")
	return ""
}
