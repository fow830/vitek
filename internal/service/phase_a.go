package service

import (
	"context"
	"errors"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgtype"

	"vitek/internal/domain"
	"vitek/internal/repository"
)

// Users creates users with an active subscription plan.
type Users struct {
	q *repository.Queries
}

func NewUsers(q *repository.Queries) *Users {
	return &Users{q: q}
}

func (s *Users) CreateUser(ctx context.Context, email string, plan repository.PlanType) (repository.User, error) {
	user, err := s.q.CreateUser(ctx, email)
	if err != nil {
		return repository.User{}, err
	}
	_, err = s.q.CreateSubscription(ctx, repository.CreateSubscriptionParams{
		UserID:   user.ID,
		PlanType: plan,
	})
	if err != nil {
		return repository.User{}, err
	}
	return user, nil
}

// Tasks enforces plan limits when creating search tasks.
type Tasks struct {
	q *repository.Queries
}

func NewTasks(q *repository.Queries) *Tasks {
	return &Tasks{q: q}
}

func (s *Tasks) CreateTask(ctx context.Context, userID pgtype.UUID, query string) (repository.Task, error) {
	sub, err := s.q.GetActiveSubscriptionByUser(ctx, userID)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return repository.Task{}, domain.ErrNoActiveSubscription
		}
		return repository.Task{}, err
	}

	maxTasks, err := s.q.GetPlanMaxTasks(ctx, sub.PlanType)
	if err != nil {
		return repository.Task{}, err
	}

	count, err := s.q.CountUserTasks(ctx, userID)
	if err != nil {
		return repository.Task{}, err
	}
	if count >= int64(maxTasks) {
		return repository.Task{}, domain.ErrSubscriptionLimitExceeded
	}

	return s.q.CreateTask(ctx, repository.CreateTaskParams{
		UserID: userID,
		Query:  query,
	})
}

// Proxies exposes only ACTIVE proxies to callers.
type Proxies struct {
	q *repository.Queries
}

func NewProxies(q *repository.Queries) *Proxies {
	return &Proxies{q: q}
}

func (s *Proxies) Create(ctx context.Context, endpoint string, status repository.ProxyStatus) (repository.Proxy, error) {
	return s.q.CreateProxy(ctx, repository.CreateProxyParams{
		Endpoint: endpoint,
		Status:   status,
	})
}

func (s *Proxies) ListActive(ctx context.Context) ([]repository.Proxy, error) {
	return s.q.ListActiveProxies(ctx)
}

// Items deduplicates Avito listings by avito_id.
type Items struct {
	q *repository.Queries
}

func NewItems(q *repository.Queries) *Items {
	return &Items{q: q}
}

func (s *Items) Record(ctx context.Context, avitoID, title string) (repository.Item, error) {
	item, err := s.q.InsertItem(ctx, repository.InsertItemParams{
		AvitoID: avitoID,
		Title:   title,
	})
	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == "23505" {
			return repository.Item{}, domain.ErrDuplicateAvitoID
		}
		return repository.Item{}, err
	}
	return item, nil
}
