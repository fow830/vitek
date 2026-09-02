package service

import (
	"context"
	"errors"
	"fmt"
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
	if tokens.IsListingFilterURL(query) {
		query = tokens.CanonicalListingSearchQuery(query)
	} else {
		query = tokens.NormalizeListingSearchURL(query)
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

func (s *Tasks) GetForUser(ctx context.Context, userID, taskID pgtype.UUID) (repository.Task, error) {
	q := repository.New(s.pool)
	task, err := q.GetTask(ctx, taskID)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return repository.Task{}, domain.ErrTaskNotFound
		}
		return repository.Task{}, err
	}
	if task.UserID != userID {
		return repository.Task{}, domain.ErrForbidden
	}
	return task, nil
}

func (s *Tasks) ListForUser(ctx context.Context, userID pgtype.UUID) ([]repository.Task, error) {
	return repository.New(s.pool).ListTasksByUser(ctx, repository.ListTasksByUserParams{
		UserID: userID,
		Limit:  tokens.ListingSearchTaskListLimit,
	})
}

func (s *Tasks) ListResultsForUser(ctx context.Context, userID, taskID pgtype.UUID) ([]repository.ListTaskItemsRow, error) {
	if _, err := s.GetForUser(ctx, userID, taskID); err != nil {
		return nil, err
	}
	return repository.New(s.pool).ListTaskItems(ctx, taskID)
}
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

func (s *Proxies) RecordHealthOK(ctx context.Context, id pgtype.UUID) error {
	return s.q.RecordProxyHealthOK(ctx, id)
}

func (s *Proxies) RecordHealthFail(ctx context.Context, id pgtype.UUID, errMsg string) (repository.Proxy, error) {
	errMsg = strings.TrimSpace(errMsg)
	if errMsg == "" {
		errMsg = tokens.ErrMsgProxyProbeFailed
	}
	return s.q.RecordProxyHealthFail(ctx, repository.RecordProxyHealthFailParams{
		LastErr:   &errMsg,
		DeadAfter: int32(tokens.ProxyDeadAfterFails),
		ID:        id,
	})
}

// ProxyProbeFunc probes one proxy against a target URL.
type ProxyProbeFunc func(ctx context.Context, proxyEndpoint, targetURL string) error

// ProbeActive runs health probes for all active (non-DEAD) proxies and pauses bindings on DEAD.
func ProbeActive(ctx context.Context, proxies *Proxies, bindings *Bindings, probe ProxyProbeFunc) (ok, fail int, err error) {
	if probe == nil {
		probe = DefaultHTTPProxyProbe
	}
	list, err := proxies.ListActive(ctx)
	if err != nil {
		return 0, 0, err
	}
	target := tokens.ProxyProbeURL()
	for _, p := range list {
		if err := probe(ctx, p.Endpoint, target); err != nil {
			updated, recErr := proxies.RecordHealthFail(ctx, p.ID, err.Error())
			if recErr != nil {
				return ok, fail, recErr
			}
			if updated.Health == repository.ProxyHealthStatusDEAD && bindings != nil {
				if err := bindings.PauseForProxy(ctx, p.ID); err != nil {
					return ok, fail, err
				}
			}
			fail++
			continue
		}
		if err := proxies.RecordHealthOK(ctx, p.ID); err != nil {
			return ok, fail, err
		}
		ok++
	}
	return ok, fail, nil
}

// AssertProxyPoolReady enforces fetchable proxy pool rules for production boots.
func AssertProxyPoolReady(ctx context.Context, proxies *Proxies, appEnv string) error {
	if appEnv != tokens.AppEnvProduction {
		return nil
	}
	list, err := proxies.ListActive(ctx)
	if err != nil {
		return err
	}
	if len(list) < tokens.ProxyPoolMinActive {
		return fmt.Errorf("%s", tokens.ErrMsgProxyPoolTooSmall)
	}
	nonBridge := 0
	for _, p := range list {
		if !tokens.IsDockerBridgeProxyEndpoint(p.Endpoint) {
			nonBridge++
		}
	}
	if nonBridge == 0 {
		return fmt.Errorf("%s", tokens.ErrMsgProxyPoolDockerBridgeOnly)
	}
	return nil
}

func (s *Proxies) Update(ctx context.Context, id pgtype.UUID, endpoint string, status repository.ProxyStatus, label string) (repository.Proxy, error) {
	return s.q.UpdateProxy(ctx, repository.UpdateProxyParams{
		ID:       id,
		Endpoint: endpoint,
		Status:   status,
		Label:    label,
	})
}

func (s *Proxies) Delete(ctx context.Context, id pgtype.UUID) error {
	_, err := s.q.DeleteProxy(ctx, id)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return domain.ErrResourceNotFound
		}
		return err
	}
	return nil
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
	return q.UpsertItem(ctx, repository.UpsertItemParams{
		AvitoID: avitoID,
		Title:   title,
	})
}
