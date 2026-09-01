package contracts_test

import (
	"testing"

	"github.com/stretchr/testify/require"

	"vitek/internal/repository"
	"vitek/internal/tokens"
)

// CONTRACT-ADMIN-UI-001: app face embeds admin CRUD panels and admin-only script hooks.
func TestContract_AdminUIPanelsInAppFace(t *testing.T) {
	body := tokens.RenderAppFaceHTML()

	require.Contains(t, body, tokens.AppDOMProxiesTable)
	require.Contains(t, body, tokens.AppDOMProxiesForm)
	require.Contains(t, body, tokens.AppDOMAvitoTable)
	require.Contains(t, body, tokens.AppDOMAvitoForm)
	require.NotContains(t, body, tokens.PathV1AdminProxies)

	admin := tokens.RenderAppFaceHTMLLoggedIn(tokens.ProductEmail("admin"), "", true)
	require.Contains(t, admin, tokens.PathV1AdminProxies)
	require.Contains(t, admin, tokens.PathV1AdminAvitoAccounts)
	require.Contains(t, admin, tokens.AppCopyAdminCreate)
	require.Contains(t, admin, tokens.AppCopyAdminDelete)

	user := tokens.RenderAppFaceHTMLLoggedIn(tokens.ProductEmail("user"), "", false)
	require.NotContains(t, user, tokens.PathV1AdminProxies)
}

// CONTRACT-ADMIN-UI-002: admin UI status options mirror DB enums.
func TestContract_AdminUIStatusValuesMatchDB(t *testing.T) {
	require.ElementsMatch(t, []string{
		string(repository.ProxyStatusACTIVE),
		string(repository.ProxyStatusBANNED),
		string(repository.ProxyStatusDISABLED),
	}, tokens.ProxyStatusValues)
	require.ElementsMatch(t, []string{
		string(repository.AvitoAccountStatusACTIVE),
		string(repository.AvitoAccountStatusDISABLED),
		string(repository.AvitoAccountStatusERROR),
	}, tokens.AvitoAccountStatusValues)
}
