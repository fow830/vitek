package service

import (
	"context"
	"errors"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/jackc/pgx/v5/pgxpool"

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

// Tasks enforces plan limits when creating search tasks (transactional, FOR UPDATE).
type Tasks struct {
	pool *pgxpool.Pool
}

func NewTasks(pool *pgxpool.Pool) *Tasks {
	return &Tasks{pool: pool}
}

func (s *Tasks) CreateTask(ctx context.Context, userID pgtype.UUID, query string) (repository.Task, error) {
	tx, err := s.pool.Begin(ctx)
	if err != nil {
		return repository.Task{}, err
	}
	defer func() { _ = tx.Rollback(ctx) }()

	qtx := repository.New(tx)

	sub, err := qtx.GetActiveSubscriptionForUpdate(ctx, userID)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return repository.Task{}, domain.ErrNoActiveSubscription
		}
		return repository.Task{}, err
	}

	count, err := qtx.CountUserTasks(ctx, userID)
	if err != nil {
		return repository.Task{}, err
	}
	if count >= int64(sub.MaxTasks) {
		return repository.Task{}, domain.ErrSubscriptionLimitExceeded
	}

	task, err := qtx.CreateTask(ctx, repository.CreateTaskParams{
		UserID: userID,
		Query:  query,
	})
	if err != nil {
		return repository.Task{}, err
	}

	if err := tx.Commit(ctx); err != nil {
		return repository.Task{}, err
	}
	return task, nil
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
