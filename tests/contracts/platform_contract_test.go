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

// CONTRACT-PLATFORM-001: Magic Link only — no password; consume-once; expired rejected.
func TestContract_AuthIsMagicLinkOnly(t *testing.T) {
	require.Equal(t, []string{tokens.AuthMethodMagicLink}, tokens.AuthMethodsAllowlist)

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

	expires := pgtype.Timestamptz{Time: time.Now().UTC().Add(tokens.MagicLinkTTL), Valid: true}
	_, err = q.CreateMagicLinkChallenge(ctx, repository.CreateMagicLinkChallengeParams{
		Email:     tokens.ProductEmail("admin"),
		TokenHash: "hash-admin-1",
		RoleHint:  repository.UserRoleADMIN,
		ExpiresAt: expires,
	})
	require.NoError(t, err)

	consumed, err := q.ConsumeMagicLinkChallenge(ctx, "hash-admin-1")
	require.NoError(t, err)
	require.True(t, consumed.ConsumedAt.Valid)
	require.Equal(t, repository.UserRoleADMIN, consumed.RoleHint)

	_, err = q.ConsumeMagicLinkChallenge(ctx, "hash-admin-1")
	require.ErrorIs(t, err, pgx.ErrNoRows)

	past := pgtype.Timestamptz{Time: time.Now().UTC().Add(-time.Minute), Valid: true}
	_, err = q.CreateMagicLinkChallenge(ctx, repository.CreateMagicLinkChallengeParams{
		Email:     tokens.ProductEmail("expired"),
		TokenHash: "hash-expired-1",
		RoleHint:  repository.UserRoleUSER,
		ExpiresAt: past,
	})
	require.NoError(t, err)
	_, err = q.ConsumeMagicLinkChallenge(ctx, "hash-expired-1")
	require.ErrorIs(t, err, pgx.ErrNoRows, "expired magic link must not consume")

	_, err = q.CreateMagicLinkChallenge(ctx, repository.CreateMagicLinkChallengeParams{
		Email:     tokens.ProductEmail("user"),
		TokenHash: "hash-user-1",
		RoleHint:  repository.UserRoleUSER,
		ExpiresAt: expires,
	})
	require.NoError(t, err)
}

// CONTRACT-PLATFORM-002: Users have USER/ADMIN roles; both can exist.
func TestContract_UserRolesUserAndAdmin(t *testing.T) {
	ctx := context.Background()
	_, q := queries(t)

	user, err := q.CreateUserWithRole(ctx, repository.CreateUserWithRoleParams{
		Email: tokens.ProductEmail("role-user"),
		Role:  repository.UserRoleUSER,
	})
	require.NoError(t, err)
	require.Equal(t, repository.UserRoleUSER, user.Role)

	admin, err := q.CreateUserWithRole(ctx, repository.CreateUserWithRoleParams{
		Email: tokens.ProductEmail("role-admin"),
		Role:  repository.UserRoleADMIN,
	})
	require.NoError(t, err)
	require.Equal(t, repository.UserRoleADMIN, admin.Role)
}

// CONTRACT-PLATFORM-003: product_services rows match tokens.ProductServiceCatalog exactly.
func TestContract_MultiServiceCatalog_ExactTokenMatch(t *testing.T) {
	ctx := context.Background()
	_, q := queries(t)

	rows, err := q.ListProductServices(ctx)
	require.NoError(t, err)
	require.Len(t, rows, len(tokens.ProductServiceCatalog))

	byCode := make(map[string]repository.ProductService, len(rows))
	for _, r := range rows {
		byCode[r.Code] = r
	}
	for _, spec := range tokens.ProductServiceCatalog {
		got, ok := byCode[spec.Code]
		require.Truef(t, ok, "missing service %s", spec.Code)
		require.Equal(t, spec.Title, got.Title)
		require.Equal(t, spec.Shipped, got.Shipped)
	}

	require.Equal(t, []string{tokens.ServiceCodeListingSearch}, tokens.ShippedServiceCodes())
	require.Equal(t, []string{tokens.ServiceCodeListingWarmup}, tokens.ReservedServiceCodes())
}

// CONTRACT-PLATFORM-004: New users get all shipped services; tasks require entitlement; no-sub forbidden.
func TestContract_UserServiceEntitlements(t *testing.T) {
	ctx := context.Background()
	pool, q := queries(t)

	users := service.NewUsers(pool)
	tasks := service.NewTasks(pool)

	user, err := users.CreateUser(ctx, tokens.ProductEmail("entitled"), repository.PlanTypeFREE)
	require.NoError(t, err)

	list, err := q.ListUserServices(ctx, user.ID)
	require.NoError(t, err)
	require.Len(t, list, len(tokens.ShippedServiceCodes()))
	require.Equal(t, tokens.ServiceCodeListingSearch, list[0].ServiceCode)

	_, err = tasks.CreateTask(ctx, user.ID, "ok")
	require.NoError(t, err)

	bare, err := q.CreateUserWithRole(ctx, repository.CreateUserWithRoleParams{
		Email: tokens.ProductEmail("no-entitle"),
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

	nosub, err := q.CreateUserWithRole(ctx, repository.CreateUserWithRoleParams{
		Email: tokens.ProductEmail("no-sub"),
		Role:  repository.UserRoleUSER,
	})
	require.NoError(t, err)
	_, err = q.GrantUserService(ctx, repository.GrantUserServiceParams{
		UserID:      nosub.ID,
		ServiceCode: tokens.ServiceCodeListingSearch,
	})
	require.NoError(t, err)
	_, err = tasks.CreateTask(ctx, nosub.ID, "nosub")
	require.ErrorIs(t, err, domain.ErrNoActiveSubscription)
}

// CONTRACT-PLATFORM-005: Avito accounts pool supports many rows; required columns exist.
func TestContract_AvitoAccountsScale(t *testing.T) {
	ctx := context.Background()
	pool, q := queries(t)

	for _, col := range []string{"id", "label", "status", "external_ref", "created_at", "updated_at"} {
		var exists bool
		err := pool.QueryRow(ctx, `
			SELECT EXISTS (
				SELECT 1 FROM information_schema.columns
				WHERE table_schema = 'public' AND table_name = 'avito_accounts' AND column_name = $1
			)`, col).Scan(&exists)
		require.NoError(t, err)
		require.Truef(t, exists, "avito_accounts.%s missing", col)
	}

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

// CONTRACT-PLATFORM-006: Admin-configurable tables exist with Magic Link challenge columns.
func TestContract_AdminManagedSchemaSurface(t *testing.T) {
	ctx := context.Background()
	pool, _ := queries(t)

	for _, table := range tokens.SchemaPlatformTables {
		var exists bool
		err := pool.QueryRow(ctx,
			`SELECT EXISTS (
				SELECT 1 FROM information_schema.tables
				WHERE table_schema = 'public' AND table_name = $1
			)`, table).Scan(&exists)
		require.NoError(t, err)
		require.Truef(t, exists, "admin-managed table missing: %s", table)
	}

	for _, col := range []string{"email", "token_hash", "role_hint", "expires_at", "consumed_at"} {
		var exists bool
		err := pool.QueryRow(ctx, `
			SELECT EXISTS (
				SELECT 1 FROM information_schema.columns
				WHERE table_schema = 'public'
				  AND table_name = 'magic_link_challenges'
				  AND column_name = $1
			)`, col).Scan(&exists)
		require.NoError(t, err)
		require.Truef(t, exists, "magic_link_challenges.%s missing", col)
	}
}

// CONTRACT-PLATFORM-007: plan_limits match tokens for FREE/PRO/ULTRA.
func TestContract_PlanLimitsMatchTokens(t *testing.T) {
	ctx := context.Background()
	_, q := queries(t)

	rows, err := q.ListPlanLimits(ctx)
	require.NoError(t, err)
	require.Len(t, rows, 3)

	want := map[repository.PlanType]int32{
		repository.PlanTypeFREE:  tokens.PlanMaxTasksFREE,
		repository.PlanTypePRO:   tokens.PlanMaxTasksPRO,
		repository.PlanTypeULTRA: tokens.PlanMaxTasksULTRA,
	}
	for _, r := range rows {
		max, ok := want[r.PlanType]
		require.Truef(t, ok, "unexpected plan %s", r.PlanType)
		require.Equal(t, max, r.MaxTasks)
	}
}
