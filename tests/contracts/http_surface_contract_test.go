package contracts_test

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/stretchr/testify/require"

	"vitek/internal/domain"
	"vitek/internal/repository"
	"vitek/internal/service"
	"vitek/internal/tokens"
	httpapi "vitek/internal/transport/httpapi"
)

// CONTRACT-HTTP-SURFACE-001: healthz / tasks / proxies HTTP behavior + path allowlist.
func TestContract_HTTPSurface(t *testing.T) {
	require.Equal(t, tokens.HTTPPathAllowlist, []string{
		tokens.PathHealthz,
		tokens.PathV1Users,
		tokens.PathV1Tasks,
		tokens.PathV1ProxiesActive,
		tokens.PathV1AuthMagicLink,
		tokens.PathV1AuthMagicLinkConsume,
		tokens.PathV1AuthLogout,
		tokens.PathV1AdminProxies,
		tokens.HTTPPathID(tokens.PathV1AdminProxies),
		tokens.PathV1AdminAvitoAccounts,
		tokens.HTTPPathID(tokens.PathV1AdminAvitoAccounts),
		tokens.PathRoot,
		tokens.PathAppSSE,
		tokens.PathTokensCSS,
	})

	pool, q := queries(t)
	handler := httpapi.NewServer(pool).Handler()
	ctx := t.Context()

	t.Run("healthz ok", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, tokens.PathHealthz, nil)
		rec := httptest.NewRecorder()
		handler.ServeHTTP(rec, req)
		require.Equal(t, http.StatusOK, rec.Code)

		var body map[string]string
		require.NoError(t, json.NewDecoder(rec.Body).Decode(&body))
		require.Equal(t, tokens.HealthStatusOK, body[tokens.JSONFieldStatus])
		require.Equal(t, tokens.ProductName, body[tokens.JSONFieldProduct])
	})

	t.Run("invalid plan_type 400", func(t *testing.T) {
		payload, err := json.Marshal(map[string]any{
			tokens.JSONFieldEmail:    tokens.ProductEmail("plan"),
			tokens.JSONFieldPlanType: tokens.FixtureInvalidEnum,
		})
		require.NoError(t, err)
		req := withAppHost(httptest.NewRequest(http.MethodPost, tokens.PathV1Users, bytes.NewReader(payload)))
		req.Header.Set(tokens.HeaderContentType, tokens.MIMEApplicationJSON)
		rec := httptest.NewRecorder()
		handler.ServeHTTP(rec, req)
		require.Equal(t, http.StatusBadRequest, rec.Code)
	})

	t.Run("tasks limit 409 and entitlement 403", func(t *testing.T) {
		users := service.NewUsers(pool)
		user, err := users.CreateUser(ctx, tokens.ProductEmail("http-tasks"), repository.PlanTypeFREE)
		require.NoError(t, err)

		createTask := func(userID pgtype.UUID, query string) *httptest.ResponseRecorder {
			body, err := json.Marshal(map[string]any{
				tokens.JSONFieldUserID: pgUUIDString(userID),
				tokens.JSONFieldQuery:  query,
			})
			require.NoError(t, err)
			req := withAppHost(httptest.NewRequest(http.MethodPost, tokens.PathV1Tasks, bytes.NewReader(body)))
			req.Header.Set(tokens.HeaderContentType, tokens.MIMEApplicationJSON)
			rec := httptest.NewRecorder()
			handler.ServeHTTP(rec, req)
			return rec
		}

		require.Equal(t, http.StatusCreated, createTask(user.ID, "q1").Code)
		require.Equal(t, http.StatusConflict, createTask(user.ID, "q2").Code)

		bare, err := q.CreateUserWithRole(ctx, repository.CreateUserWithRoleParams{
			Email: tokens.ProductEmail("http-noent"),
			Role:  repository.UserRoleUSER,
		})
		require.NoError(t, err)
		_, err = q.CreateSubscription(ctx, repository.CreateSubscriptionParams{
			UserID:   bare.ID,
			PlanType: repository.PlanTypeFREE,
		})
		require.NoError(t, err)

		rec := createTask(bare.ID, "blocked")
		require.Equal(t, http.StatusForbidden, rec.Code)

		var errBody map[string]string
		require.NoError(t, json.NewDecoder(rec.Body).Decode(&errBody))
		require.Equal(t, domain.ErrServiceNotEntitled.Error(), errBody[tokens.JSONFieldError])
	})

	t.Run("proxies active only via HTTP", func(t *testing.T) {
		proxies := service.NewProxies(q)
		_, err := proxies.Create(ctx, "http://active.http.example:8080", repository.ProxyStatusACTIVE, "")
		require.NoError(t, err)
		_, err = proxies.Create(ctx, "http://banned.http.example:8080", repository.ProxyStatusBANNED, "")
		require.NoError(t, err)

		req := withAppHost(httptest.NewRequest(http.MethodGet, tokens.PathV1ProxiesActive, nil))
		rec := httptest.NewRecorder()
		handler.ServeHTTP(rec, req)
		require.Equal(t, http.StatusOK, rec.Code)

		var body map[string]any
		require.NoError(t, json.NewDecoder(rec.Body).Decode(&body))
		list, ok := body[tokens.JSONFieldProxies].([]any)
		require.True(t, ok)
		require.Len(t, list, 1)
		row := list[0].(map[string]any)
		require.Equal(t, "http://active.http.example:8080", row[tokens.JSONFieldEndpoint])
		require.Equal(t, string(repository.ProxyStatusACTIVE), row[tokens.JSONFieldStatus])
	})
}

func pgUUIDString(id pgtype.UUID) string {
	if !id.Valid {
		return ""
	}
	return uuid.UUID(id.Bytes).String()
}
