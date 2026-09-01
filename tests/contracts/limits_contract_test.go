package contracts_test

import (
	"context"
	"testing"

	"github.com/stretchr/testify/require"

	"vitek/internal/domain"
	"vitek/internal/repository"
	"vitek/internal/service"
)

// CONTRACT-LIMITS-001: FREE plan cannot exceed plan_limits.max_tasks (1).
// Context: SaaS abuse guard — free tier is single-task only.
func TestContract_SubscriptionTaskLimits(t *testing.T) {
	ctx := context.Background()
	_, q := queries(t)

	users := service.NewUsers(q)
	tasks := service.NewTasks(q)

	t.Run("FREE plan strictly enforces single task limit", func(t *testing.T) {
		user, err := users.CreateUser(ctx, "dev@vitek.io", repository.PlanTypeFREE)
		require.NoError(t, err)

		_, err = tasks.CreateTask(ctx, user.ID, "Avito Query 1")
		require.NoError(t, err)

		_, err = tasks.CreateTask(ctx, user.ID, "Avito Query 2")
		require.ErrorIs(t, err, domain.ErrSubscriptionLimitExceeded)
	})
}
