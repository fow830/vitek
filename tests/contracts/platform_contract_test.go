package contracts_test

import (
	"context"
	"strconv"
	"testing"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/stretchr/testify/require"

	"vitek/internal/domain"
	"vitek/internal/repository"
	"vitek/internal/service"
	"vitek/internal/tokens"
)

// CONTRACT-PLATFORM-001: Magic Link challenges for USER and ADMIN; consume once; no password column.
func TestContract_AuthIsMagicLinkOnly(t *testing.T) {
	require.Equal(t, "MAGIC_LINK", tokens.AuthMethodMagicLink)

	ctx := context.Background()
	pool, q := queries(t)

	var passwordCol bool
	err := pool.QueryRow(ctx, `
		SELECT EXISTS (
			SELECT 1 FROM information_schema.columns
			WHERE table_schema = 'public'
			  AND table_name = 'users'
			  AND column_name = $1
		)`, tokens.TableUsersPasswordColumn).Scan(&passwordCol)
	require.NoError(t, err)
	require.False(t, passwordCol, "users must not have a password column")

	expires := pgtype.Timestamptz{Time: time.Now().UTC().Add(15 * time.Minute), Valid: true}
	chal, err := q.CreateMagicLinkChallenge(ctx, repository.CreateMagicLinkChallengeParams{
		Email:     "admin@vitek.io",
		TokenHash: "hash-admin-1",
		RoleHint:  repository.UserRoleADMIN,
		ExpiresAt: expires,
	})
	require.NoError(t, err)
	require.Equal(t, repository.UserRoleADMIN, chal.RoleHint)
	require.False(t, chal.ConsumedAt.Valid)

	consumed, err := q.ConsumeMagicLinkChallenge(ctx, "hash-admin-1")
	require.NoError(t, err)
	require.True(t, consumed.ConsumedAt.Valid)

	_, err = q.ConsumeMagicLinkChallenge(ctx, "hash-admin-1")
	require.ErrorIs(t, err, pgx.ErrNoRows)

	chalUser, err := q.CreateMagicLinkChallenge(ctx, repository.CreateMagicLinkChallengeParams{
		Email:     "user@vitek.io",
		TokenHash: "hash-user-1",
		RoleHint:  repository.UserRoleUSER,
		ExpiresAt: expires,
	})
	require.NoError(t, err)
	require.Equal(t, repository.UserRoleUSER, chalUser.RoleHint)
}

// CONTRACT-PLATFORM-002: Users have USER/ADMIN roles; both can exist.
func TestContract_UserRolesUserAndAdmin(t *testing.T) {
	ctx := context.Background()
	_, q := queries(t)

	user, err := q.CreateUserWithRole(ctx, repository.CreateUserWithRoleParams{
		Email: "role-user@vitek.io",
		Role:  repository.UserRoleUSER,
	})
	require.NoError(t, err)
	require.Equal(t, repository.UserRoleUSER, user.Role)

	admin, err := q.CreateUserWithRole(ctx, repository.CreateUserWithRoleParams{
		Email: "role-admin@vitek.io",
		Role:  repository.UserRoleADMIN,
	})
	require.NoError(t, err)
	require.Equal(t, repository.UserRoleADMIN, admin.Role)
}

// CONTRACT-PLATFORM-003: Vitek is multi-service; only listing_search is shipped; titles match tokens.
func TestContract_MultiServiceCatalog_ShippedListingSearchOnly(t *testing.T) {
	require.Equal(t, []string{tokens.ServiceCodeListingSearch}, tokens.ShippedServiceCodes)
	require.Contains(t, tokens.ReservedServiceCodes, tokens.ServiceCodeListingWarmup)

	ctx := context.Background()
	_, q := queries(t)

	shipped, err := q.ListShippedProductServices(ctx)
	require.NoError(t, err)
	require.Len(t, shipped, 1)
	require.Equal(t, tokens.ServiceCodeListingSearch, shipped[0].Code)
	require.Equal(t, tokens.ServiceTitleListingSearch, shipped[0].Title)
	require.True(t, shipped[0].Shipped)

	warmup, err := q.GetProductService(ctx, tokens.ServiceCodeListingWarmup)
	require.NoError(t, err)
	require.Equal(t, tokens.ServiceTitleListingWarmup, warmup.Title)
	require.False(t, warmup.Shipped, "listing_warmup must stay unshipped until explicitly built")
}

// CONTRACT-PLATFORM-004: New users get listing_search; tasks require entitlement.
func TestContract_UserServiceEntitlements(t *testing.T) {
	ctx := context.Background()
	pool, q := queries(t)

	users := service.NewUsers(pool)
	tasks := service.NewTasks(pool)

	user, err := users.CreateUser(ctx, "entitled@vitek.io", repository.PlanTypeFREE)
	require.NoError(t, err)

	list, err := q.ListUserServices(ctx, user.ID)
	require.NoError(t, err)
	require.Len(t, list, 1)
	require.Equal(t, tokens.ServiceCodeListingSearch, list[0].ServiceCode)

	_, err = tasks.CreateTask(ctx, user.ID, "ok")
	require.NoError(t, err)

	bare, err := q.CreateUserWithRole(ctx, repository.CreateUserWithRoleParams{
		Email: "no-entitle@vitek.io",
		Role:  repository.UserRoleUSER,
	})
	require.NoError(t, err)
	_, err = q.CreateSubscription(ctx, repository.CreateSubscriptionParams{
		UserID:   bare.ID,
		PlanType: repository.PlanTypeFREE,
	})
	require.NoError(t, err)

	_, err = tasks.CreateTask(ctx, bare.ID, "blocked")
	require.ErrorIs(t, err, domain.ErrServiceNotEntitled)
}

// CONTRACT-PLATFORM-005: Many Avito accounts can be registered (admin-managed pool).
func TestContract_AvitoAccountsScale(t *testing.T) {
	ctx := context.Background()
	_, q := queries(t)

	const n = 50
	for i := 0; i < n; i++ {
		_, err := q.CreateAvitoAccount(ctx, repository.CreateAvitoAccountParams{
			Label:       "acc-" + strconv.Itoa(i),
			Status:      repository.AvitoAccountStatusDISABLED,
			ExternalRef: "ref-" + strconv.Itoa(i),
		})
		require.NoError(t, err)
	}

	count, err := q.CountAvitoAccounts(ctx)
	require.NoError(t, err)
	require.Equal(t, int64(n), count)
}

// CONTRACT-PLATFORM-006: Admin-configurable surface exists in schema (web UI later).
func TestContract_AdminManagedSchemaSurface(t *testing.T) {
	ctx := context.Background()
	pool, _ := queries(t)

	for _, table := range tokens.AdminManagedTables {
		var exists bool
		err := pool.QueryRow(ctx,
			`SELECT EXISTS (
				SELECT 1 FROM information_schema.tables
				WHERE table_schema = 'public' AND table_name = $1
			)`, table).Scan(&exists)
		require.NoError(t, err)
		require.Truef(t, exists, "admin-managed table missing: %s", table)
	}
}
