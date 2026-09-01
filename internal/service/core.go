package service

import (
	"context"
	"errors"
	"strings"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/jackc/pgx/v5/pgxpool"

	"vitek/internal/domain"
	"vitek/internal/repository"
	"vitek/internal/tokens"
)

// Users creates users with an active subscription and default shipped-service entitlement.
type Users struct {
	pool *pgxpool.Pool
}

func NewUsers(pool *pgxpool.Pool) *Users {
	return &Users{pool: pool}
}

func (s *Users) CreateUser(ctx context.Context, email string, plan repository.PlanType) (repository.User, error) {
	tx, err := s.pool.Begin(ctx)
	if err != nil {
		return repository.User{}, err
	}
	defer func() { _ = tx.Rollback(ctx) }()

	qtx := repository.New(tx)

	user, err := qtx.CreateUser(ctx, email)
	if err != nil {
		return repository.User{}, err
	}
	_, err = qtx.CreateSubscription(ctx, repository.CreateSubscriptionParams{
		UserID:   user.ID,
		PlanType: plan,
	})
	if err != nil {
		return repository.User{}, err
	}
	for _, code := range tokens.ShippedServiceCodes() {
		_, err = qtx.GrantUserService(ctx, repository.GrantUserServiceParams{
			UserID:      user.ID,
			ServiceCode: code,
		})
		if err != nil {
			return repository.User{}, err
		}
	}

	if err := tx.Commit(ctx); err != nil {
		return repository.User{}, err
	}
	return user, nil
}

// Tasks enforces entitlements + plan limits when creating search tasks (transactional, FOR UPDATE).
type Tasks struct {
	pool *pgxpool.Pool
}

func NewTasks(pool *pgxpool.Pool) *Tasks {
	return &Tasks{pool: pool}
}

func (s *Tasks) CreateTask(ctx context.Context, userID pgtype.UUID, query string) (repository.Task, error) {
	query = strings.TrimSpace(query)
	if !tokens.ValidListingURL(query) {
		return repository.Task{}, domain.ErrInvalidListingURL
	}

	tx, err := s.pool.Begin(ctx)
	if err != nil {
		return repository.Task{}, err
	}
	defer func() { _ = tx.Rollback(ctx) }()

	qtx := repository.New(tx)

	ok, err := qtx.HasUserService(ctx, repository.HasUserServiceParams{
		UserID:      userID,
		ServiceCode: tokens.ServiceCodeListingSearch,
	})
	if err != nil {
		return repository.Task{}, err
	}
	if !ok {
		return repository.Task{}, domain.ErrServiceNotEntitled
	}

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

func NewProxies(pool *pgxpool.Pool) *Proxies {
	return &Proxies{q: repository.New(pool)}
}

func (s *Proxies) Create(ctx context.Context, endpoint string, status repository.ProxyStatus, label string) (repository.Proxy, error) {
	return s.q.CreateProxy(ctx, repository.CreateProxyParams{
		Endpoint: endpoint,
		Status:   status,
		Label:    label,
	})
}

func (s *Proxies) ListActive(ctx context.Context) ([]repository.Proxy, error) {
	return s.q.ListActiveProxies(ctx)
}

func (s *Proxies) ListAll(ctx context.Context) ([]repository.Proxy, error) {
	return s.q.ListAllProxies(ctx)
}

func (s *Proxies) Update(ctx context.Context, id pgtype.UUID, endpoint string, status repository.ProxyStatus, label string) (repository.Proxy, error) {
	return s.q.UpdateProxy(ctx, repository.UpdateProxyParams{
		ID:       id,
		Endpoint: endpoint,
		Status:   status,
		Label:    label,
	})
}

// Items deduplicates Avito listings by avito_id.
type Items struct {
	pool *pgxpool.Pool
}

func NewItems(pool *pgxpool.Pool) *Items {
	return &Items{pool: pool}
}

func (s *Items) Record(ctx context.Context, avitoID, title string) (repository.Item, error) {
	item, err := repository.New(s.pool).InsertItem(ctx, repository.InsertItemParams{
		AvitoID: avitoID,
		Title:   title,
	})
	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == tokens.PGCodeUniqueViolation {
			return repository.Item{}, domain.ErrDuplicateAvitoID
		}
		return repository.Item{}, err
	}
	return item, nil
}

// UpsertTx inserts an item or returns the existing row on avito_id conflict (same transaction).
func (s *Items) UpsertTx(ctx context.Context, q *repository.Queries, avitoID, title string) (repository.Item, error) {
	item, err := q.InsertItem(ctx, repository.InsertItemParams{
		AvitoID: avitoID,
		Title:   title,
	})
	if err == nil {
		return item, nil
	}
	var pgErr *pgconn.PgError
	if errors.As(err, &pgErr) && pgErr.Code == tokens.PGCodeUniqueViolation {
		return q.GetItemByAvitoID(ctx, avitoID)
	}
	return repository.Item{}, err
}
