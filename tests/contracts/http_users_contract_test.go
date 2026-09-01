package contracts_test

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/require"

	"vitek/internal/repository"
	"vitek/internal/tokens"
	httpapi "vitek/internal/transport/httpapi"
)

// CONTRACT-HTTP-USERS-001: POST /v1/users returns 400 on malformed email and 409 on duplicate.
func TestContract_CreateUserHTTPValidationAndConflict(t *testing.T) {
	pool, _ := queries(t)
	handler := httpapi.NewServer(pool).Handler()

	t.Run("malformed email returns 400", func(t *testing.T) {
		body, err := json.Marshal(map[string]any{
			tokens.JSONFieldEmail:    "not-an-email",
			tokens.JSONFieldPlanType: repository.PlanTypeFREE,
		})
		require.NoError(t, err)

		req := httptest.NewRequest(http.MethodPost, tokens.PathV1Users, bytes.NewReader(body))
		req.Header.Set(tokens.HeaderContentType, tokens.MIMEApplicationJSON)
		rec := httptest.NewRecorder()
		handler.ServeHTTP(rec, req)

		require.Equal(t, http.StatusBadRequest, rec.Code)
	})

	t.Run("duplicate email returns 409", func(t *testing.T) {
		payload, err := json.Marshal(map[string]any{
			tokens.JSONFieldEmail:    tokens.ProductEmail("dup"),
			tokens.JSONFieldPlanType: repository.PlanTypeFREE,
		})
		require.NoError(t, err)

		req1 := httptest.NewRequest(http.MethodPost, tokens.PathV1Users, bytes.NewReader(payload))
		req1.Header.Set(tokens.HeaderContentType, tokens.MIMEApplicationJSON)
		rec1 := httptest.NewRecorder()
		handler.ServeHTTP(rec1, req1)
		require.Equal(t, http.StatusCreated, rec1.Code)

		req2 := httptest.NewRequest(http.MethodPost, tokens.PathV1Users, bytes.NewReader(payload))
		req2.Header.Set(tokens.HeaderContentType, tokens.MIMEApplicationJSON)
		rec2 := httptest.NewRecorder()
		handler.ServeHTTP(rec2, req2)
		require.Equal(t, http.StatusConflict, rec2.Code)
	})
}
