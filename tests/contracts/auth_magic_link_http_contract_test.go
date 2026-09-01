package contracts_test

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/stretchr/testify/require"

	"vitek/internal/repository"
	"vitek/internal/service"
	"vitek/internal/tokens"
	httpapi "vitek/internal/transport/httpapi"
)

// CONTRACT-AUTH-ML-001: request Magic Link stores challenge; consume issues session cookie.
func TestContract_MagicLinkHTTP_RequestAndConsume(t *testing.T) {
	pool, q := queries(t)
	mailer := service.NewMemoryMagicLinkMailer()
	handler := httpapi.NewServer(pool, httpapi.WithMagicLinkMailer(mailer)).Handler()
	ctx := t.Context()

	_, err := q.CreateUserWithRole(ctx, repository.CreateUserWithRoleParams{
		Email: tokens.ProductEmail("admin-ml"),
		Role:  repository.UserRoleADMIN,
	})
	require.NoError(t, err)

	body, err := json.Marshal(map[string]any{tokens.JSONFieldEmail: tokens.ProductEmail("admin-ml")})
	require.NoError(t, err)
	req := withAppHost(httptest.NewRequest(http.MethodPost, tokens.PathV1AuthMagicLink, bytes.NewReader(body)))
	req.Header.Set(tokens.HeaderContentType, tokens.MIMEApplicationJSON)
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)
	require.Equal(t, http.StatusAccepted, rec.Code)
	require.NotEmpty(t, mailer.LastToken)
	require.Equal(t, tokens.ProductEmail("admin-ml"), mailer.LastEmail)

	consumeBody, err := json.Marshal(map[string]any{tokens.JSONFieldToken: mailer.LastToken})
	require.NoError(t, err)
	creq := withAppHost(httptest.NewRequest(http.MethodPost, tokens.PathV1AuthMagicLinkConsume, bytes.NewReader(consumeBody)))
	creq.Header.Set(tokens.HeaderContentType, tokens.MIMEApplicationJSON)
	crec := httptest.NewRecorder()
	handler.ServeHTTP(crec, creq)
	require.Equal(t, http.StatusOK, crec.Code)
	require.True(t, strings.Contains(crec.Header().Get(tokens.HeaderSetCookie), tokens.CookieSessionName+"="))

	var out map[string]any
	require.NoError(t, json.NewDecoder(crec.Body).Decode(&out))
	require.Equal(t, tokens.ProductEmail("admin-ml"), out[tokens.JSONFieldEmail])
	require.Equal(t, string(repository.UserRoleADMIN), out[tokens.JSONFieldRole])

	// reuse token fails
	creq2 := withAppHost(httptest.NewRequest(http.MethodPost, tokens.PathV1AuthMagicLinkConsume, bytes.NewReader(consumeBody)))
	creq2.Header.Set(tokens.HeaderContentType, tokens.MIMEApplicationJSON)
	crec2 := httptest.NewRecorder()
	handler.ServeHTTP(crec2, creq2)
	require.Equal(t, http.StatusUnauthorized, crec2.Code)
}

// CONTRACT-AUTH-ML-002: expired challenge cannot be consumed.
func TestContract_MagicLinkHTTP_ExpiredRejected(t *testing.T) {
	pool, _ := queries(t)
	mailer := service.NewMemoryMagicLinkMailer()
	auth := service.NewAuth(pool, mailer)
	ctx := t.Context()

	_, err := repository.New(pool).CreateUserWithRole(ctx, repository.CreateUserWithRoleParams{
		Email: tokens.ProductEmail("exp"),
		Role:  repository.UserRoleUSER,
	})
	require.NoError(t, err)

	raw, err := auth.RequestMagicLink(ctx, tokens.ProductEmail("exp"), -time.Minute)
	require.NoError(t, err)
	require.NotEmpty(t, raw)

	_, _, err = auth.ConsumeMagicLink(ctx, raw)
	require.ErrorIs(t, err, service.ErrInvalidMagicLink)
}

// CONTRACT-AUTH-ML-003: unknown email still accepted (no enumeration); mailer gets token.
func TestContract_MagicLinkHTTP_UnknownEmailAccepted(t *testing.T) {
	pool, _ := queries(t)
	mailer := service.NewMemoryMagicLinkMailer()
	handler := httpapi.NewServer(pool, httpapi.WithMagicLinkMailer(mailer)).Handler()

	body, err := json.Marshal(map[string]any{tokens.JSONFieldEmail: tokens.ProductEmail("newuser")})
	require.NoError(t, err)
	req := withAppHost(httptest.NewRequest(http.MethodPost, tokens.PathV1AuthMagicLink, bytes.NewReader(body)))
	req.Header.Set(tokens.HeaderContentType, tokens.MIMEApplicationJSON)
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)
	require.Equal(t, http.StatusAccepted, rec.Code)
	require.Equal(t, tokens.ProductEmail("newuser"), mailer.LastEmail)
	require.NotEmpty(t, mailer.LastToken)

	consumeBody, err := json.Marshal(map[string]any{tokens.JSONFieldToken: mailer.LastToken})
	require.NoError(t, err)
	creq := withAppHost(httptest.NewRequest(http.MethodPost, tokens.PathV1AuthMagicLinkConsume, bytes.NewReader(consumeBody)))
	creq.Header.Set(tokens.HeaderContentType, tokens.MIMEApplicationJSON)
	crec := httptest.NewRecorder()
	handler.ServeHTTP(crec, creq)
	require.Equal(t, http.StatusOK, crec.Code)

	var out map[string]any
	require.NoError(t, json.NewDecoder(crec.Body).Decode(&out))
	require.Equal(t, tokens.ProductEmail("newuser"), out[tokens.JSONFieldEmail])
	require.Equal(t, string(repository.UserRoleUSER), out[tokens.JSONFieldRole])

	setCookie := crec.Header().Get(tokens.HeaderSetCookie)
	require.Contains(t, setCookie, tokens.CookieSessionName+"=")
	require.Contains(t, setCookie, tokens.CookieAttrHttpOnly)
	require.Contains(t, setCookie, tokens.CookieAttrPath+tokens.CookiePath)
	require.Contains(t, setCookie, tokens.CookieAttrSameSite+tokens.CookieSameSite)

	// New Magic Link signup must get FREE subscription (can create a task).
	userID := out[tokens.JSONFieldID].(string)
	uid, err := uuid.Parse(userID)
	require.NoError(t, err)
	sub, err := repository.New(pool).GetActiveSubscription(t.Context(), pgtype.UUID{Bytes: uid, Valid: true})
	require.NoError(t, err)
	require.Equal(t, repository.PlanTypeFREE, sub.PlanType)

	taskBody, err := json.Marshal(map[string]any{
		tokens.JSONFieldUserID: userID,
		tokens.JSONFieldQuery:  "ml-signup-task",
	})
	require.NoError(t, err)
	treq := withAppHost(httptest.NewRequest(http.MethodPost, tokens.PathV1Tasks, bytes.NewReader(taskBody)))
	treq.Header.Set(tokens.HeaderContentType, tokens.MIMEApplicationJSON)
	trec := httptest.NewRecorder()
	handler.ServeHTTP(trec, treq)
	require.Equal(t, http.StatusCreated, trec.Code)
}

// CONTRACT-AUTH-ML-004: logout revokes session; cookie no longer authorizes app SSE.
func TestContract_MagicLinkHTTP_LogoutRevokesSession(t *testing.T) {
	pool, q := queries(t)
	mailer := service.NewMemoryMagicLinkMailer()
	handler := httpapi.NewServer(pool, httpapi.WithMagicLinkMailer(mailer)).Handler()
	ctx := t.Context()

	_, err := q.CreateUserWithRole(ctx, repository.CreateUserWithRoleParams{
		Email: tokens.ProductEmail("logout-admin"),
		Role:  repository.UserRoleADMIN,
	})
	require.NoError(t, err)

	body, err := json.Marshal(map[string]any{tokens.JSONFieldEmail: tokens.ProductEmail("logout-admin")})
	require.NoError(t, err)
	req := withAppHost(httptest.NewRequest(http.MethodPost, tokens.PathV1AuthMagicLink, bytes.NewReader(body)))
	req.Header.Set(tokens.HeaderContentType, tokens.MIMEApplicationJSON)
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)
	require.Equal(t, http.StatusAccepted, rec.Code)

	cbody, err := json.Marshal(map[string]any{tokens.JSONFieldToken: mailer.LastToken})
	require.NoError(t, err)
	creq := withAppHost(httptest.NewRequest(http.MethodPost, tokens.PathV1AuthMagicLinkConsume, bytes.NewReader(cbody)))
	creq.Header.Set(tokens.HeaderContentType, tokens.MIMEApplicationJSON)
	crec := httptest.NewRecorder()
	handler.ServeHTTP(crec, creq)
	require.Equal(t, http.StatusOK, crec.Code)

	var cookie string
	for _, c := range crec.Result().Cookies() {
		if c.Name == tokens.CookieSessionName {
			cookie = c.Value
		}
	}
	require.NotEmpty(t, cookie)

	lreq := withAppHost(httptest.NewRequest(http.MethodPost, tokens.PathV1AuthLogout, nil))
	lreq.Header.Set(tokens.HeaderCookie, tokens.CookieSessionName+"="+cookie)
	lrec := httptest.NewRecorder()
	handler.ServeHTTP(lrec, lreq)
	require.Equal(t, http.StatusOK, lrec.Code)
	require.Contains(t, lrec.Header().Get(tokens.HeaderSetCookie), tokens.CookieSessionName+"=")

	sreq := appHostRequest(http.MethodGet, tokens.PathAppSSE)
	sreq.Header.Set(tokens.HeaderCookie, tokens.CookieSessionName+"="+cookie)
	srec := httptest.NewRecorder()
	handler.ServeHTTP(srec, sreq)
	require.Equal(t, http.StatusUnauthorized, srec.Code)
}

// CONTRACT-AUTH-ML-005: Secure cookie flag appears only when WithSecureCookies(true).
func TestContract_MagicLinkHTTP_SecureCookieFlag(t *testing.T) {
	pool, q := queries(t)
	mailer := service.NewMemoryMagicLinkMailer()
	ctx := t.Context()
	_, err := q.CreateUserWithRole(ctx, repository.CreateUserWithRoleParams{
		Email: tokens.ProductEmail("secure"),
		Role:  repository.UserRoleUSER,
	})
	require.NoError(t, err)

	consume := func(handler http.Handler) string {
		t.Helper()
		body, err := json.Marshal(map[string]any{tokens.JSONFieldEmail: tokens.ProductEmail("secure")})
		require.NoError(t, err)
		req := withAppHost(httptest.NewRequest(http.MethodPost, tokens.PathV1AuthMagicLink, bytes.NewReader(body)))
		req.Header.Set(tokens.HeaderContentType, tokens.MIMEApplicationJSON)
		rec := httptest.NewRecorder()
		handler.ServeHTTP(rec, req)
		require.Equal(t, http.StatusAccepted, rec.Code)

		cbody, err := json.Marshal(map[string]any{tokens.JSONFieldToken: mailer.LastToken})
		require.NoError(t, err)
		creq := withAppHost(httptest.NewRequest(http.MethodPost, tokens.PathV1AuthMagicLinkConsume, bytes.NewReader(cbody)))
		creq.Header.Set(tokens.HeaderContentType, tokens.MIMEApplicationJSON)
		crec := httptest.NewRecorder()
		handler.ServeHTTP(crec, creq)
		require.Equal(t, http.StatusOK, crec.Code)
		return crec.Header().Get(tokens.HeaderSetCookie)
	}

	insecure := consume(httpapi.NewServer(pool, httpapi.WithMagicLinkMailer(mailer)).Handler())
	require.NotContains(t, insecure, tokens.CookieAttrSecure)

	secure := consume(httpapi.NewServer(pool,
		httpapi.WithMagicLinkMailer(mailer),
		httpapi.WithSecureCookies(true),
	).Handler())
	require.Contains(t, secure, tokens.CookieAttrSecure)
}

// CONTRACT-AUTH-ML-006: default server must not leak raw magic token in HTTP JSON.
func TestContract_MagicLinkHTTP_DefaultDoesNotExposeToken(t *testing.T) {
	pool, _ := queries(t)
	mailer := service.NewMemoryMagicLinkMailer()
	handler := httpapi.NewServer(pool, httpapi.WithMagicLinkMailer(mailer)).Handler()

	body, err := json.Marshal(map[string]any{tokens.JSONFieldEmail: tokens.ProductEmail("noexpose")})
	require.NoError(t, err)
	req := withAppHost(httptest.NewRequest(http.MethodPost, tokens.PathV1AuthMagicLink, bytes.NewReader(body)))
	req.Header.Set(tokens.HeaderContentType, tokens.MIMEApplicationJSON)
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)
	require.Equal(t, http.StatusAccepted, rec.Code)
	require.NotEmpty(t, mailer.LastToken)

	var out map[string]any
	require.NoError(t, json.NewDecoder(rec.Body).Decode(&out))
	_, hasToken := out[tokens.JSONFieldToken]
	require.False(t, hasToken, "production-default must not put raw token in JSON")
}

// CONTRACT-AUTH-ML-007: WithExposeMagicLinkTokens(true) returns token for local app shell.
func TestContract_MagicLinkHTTP_ExposeTokenOptIn(t *testing.T) {
	pool, _ := queries(t)
	mailer := service.NewMemoryMagicLinkMailer()
	handler := httpapi.NewServer(pool,
		httpapi.WithMagicLinkMailer(mailer),
		httpapi.WithExposeMagicLinkTokens(true),
	).Handler()

	body, err := json.Marshal(map[string]any{tokens.JSONFieldEmail: tokens.ProductEmail("expose")})
	require.NoError(t, err)
	req := withAppHost(httptest.NewRequest(http.MethodPost, tokens.PathV1AuthMagicLink, bytes.NewReader(body)))
	req.Header.Set(tokens.HeaderContentType, tokens.MIMEApplicationJSON)
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)
	require.Equal(t, http.StatusAccepted, rec.Code)

	var out map[string]any
	require.NoError(t, json.NewDecoder(rec.Body).Decode(&out))
	require.Equal(t, mailer.LastToken, out[tokens.JSONFieldToken])
}
