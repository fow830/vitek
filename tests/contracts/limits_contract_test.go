package contracts_test

import (
	"context"
	"errors"
	"fmt"
	"sync"
	"testing"

	"github.com/stretchr/testify/require"

	"vitek/internal/domain"
	"vitek/internal/repository"
	"vitek/internal/service"
	"vitek/internal/tokens"
)

// CONTRACT-LIMITS-001: FREE plan cannot exceed plan_limits.max_tasks (1).
// Context: SaaS abuse guard — free tier is single-task only.
func TestContract_SubscriptionTaskLimits(t *testing.T) {
	ctx := context.Background()
	pool, _ := queries(t)

	users := service.NewUsers(pool)
	tasks := service.NewTasks(pool)

	t.Run("FREE plan strictly enforces single task limit", func(t *testing.T) {
		user, err := users.CreateUser(ctx, tokens.ProductEmail("dev"), repository.PlanTypeFREE)
		require.NoError(t, err)

		_, err = tasks.CreateTask(ctx, user.ID, "Avito Query 1")
		require.NoError(t, err)

		_, err = tasks.CreateTask(ctx, user.ID, "Avito Query 2")
		require.ErrorIs(t, err, domain.ErrSubscriptionLimitExceeded)
	})
}

// CONTRACT-LIMITS-002: Parallel creates cannot bypass FREE max_tasks via TOCTOU.
func TestContract_SubscriptionTaskLimits_Concurrent(t *testing.T) {
	ctx := context.Background()
	pool, _ := queries(t)

	users := service.NewUsers(pool)
	tasks := service.NewTasks(pool)

	user, err := users.CreateUser(ctx, tokens.ProductEmail("concurrent"), repository.PlanTypeFREE)
	require.NoError(t, err)

	const workers = 10
	var wg sync.WaitGroup
	errs := make(chan error, workers)

	for i := 0; i < workers; i++ {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()
			_, createErr := tasks.CreateTask(ctx, user.ID, fmt.Sprintf("Avito Query %d", id))
			errs <- createErr
		}(i)
	}

	wg.Wait()
	close(errs)

	var successCount, limitErrCount int
	for err := range errs {
		switch {
		case err == nil:
			successCount++
		case errors.Is(err, domain.ErrSubscriptionLimitExceeded):
			limitErrCount++
		default:
			require.NoError(t, err)
		}
	}

	require.Equal(t, 1, successCount, "exactly one task must succeed on FREE")
	require.Equal(t, workers-1, limitErrCount, "all other workers must hit limit error")
}
