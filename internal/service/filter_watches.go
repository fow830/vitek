package service

import (
	"context"
	"strings"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/jackc/pgx/v5/pgxpool"

	"vitek/internal/domain"
	"vitek/internal/repository"
	"vitek/internal/tokens"
)

// FilterWatches tracks active Avito filter streams for periodic polling.
type FilterWatches struct {
	pool *pgxpool.Pool
}

func NewFilterWatches(pool *pgxpool.Pool) *FilterWatches {
	return &FilterWatches{pool: pool}
}

func (s *FilterWatches) Start(ctx context.Context, userID pgtype.UUID, query string) (repository.ListingFilterWatch, error) {
	raw := strings.TrimSpace(query)
	if !tokens.IsListingFilterURL(raw) {
		return repository.ListingFilterWatch{}, domain.ErrInvalidListingURL
	}
	fetchQuery := tokens.NormalizeListingSearchURL(raw)
	filterKey := tokens.CanonicalListingSearchQuery(raw)

	tx, err := s.pool.Begin(ctx)
	if err != nil {
		return repository.ListingFilterWatch{}, err
	}
	defer func() { _ = tx.Rollback(ctx) }()

	qtx := repository.New(tx)

	ok, err := qtx.HasUserService(ctx, repository.HasUserServiceParams{
		UserID:      userID,
		ServiceCode: tokens.ServiceCodeListingSearch,
	})
	if err != nil {
		return repository.ListingFilterWatch{}, err
	}
	if !ok {
		return repository.ListingFilterWatch{}, domain.ErrServiceNotEntitled
	}

	sub, err := qtx.GetActiveSubscriptionForUpdate(ctx, userID)
	if err != nil {
		if err == pgx.ErrNoRows {
			return repository.ListingFilterWatch{}, domain.ErrNoActiveSubscription
		}
		return repository.ListingFilterWatch{}, err
	}

	existing, err := qtx.GetFilterWatchByUserFilter(ctx, repository.GetFilterWatchByUserFilterParams{
		UserID:    userID,
		FilterKey: filterKey,
	})
	needsSlot := true
	if err == nil {
		if existing.Status == repository.ListingWatchStatusACTIVE || existing.Status == repository.ListingWatchStatusPAUSED {
			needsSlot = false
		}
	} else if err != pgx.ErrNoRows {
		return repository.ListingFilterWatch{}, err
	}

	if needsSlot {
		count, err := qtx.CountUserWatches(ctx, userID)
		if err != nil {
			return repository.ListingFilterWatch{}, err
		}
		if count >= int64(sub.MaxWatches) {
			return repository.ListingFilterWatch{}, domain.ErrWatchLimitExceeded
		}
	}

	watch, err := qtx.UpsertFilterWatch(ctx, repository.UpsertFilterWatchParams{
		UserID:    userID,
		FilterKey: filterKey,
		Query:     fetchQuery,
	})
	if err != nil {
		return repository.ListingFilterWatch{}, err
	}

	if err := tx.Commit(ctx); err != nil {
		return repository.ListingFilterWatch{}, err
	}
	return watch, nil
}

func (s *FilterWatches) UpdateStatus(ctx context.Context, userID, watchID pgtype.UUID, statusRaw string) (repository.ListingFilterWatch, error) {
	var status repository.ListingWatchStatus
	switch statusRaw {
	case tokens.ListingWatchStatusActive:
		status = repository.ListingWatchStatusACTIVE
	case tokens.ListingWatchStatusPaused:
		status = repository.ListingWatchStatusPAUSED
	case tokens.ListingWatchStatusDisabled:
		status = repository.ListingWatchStatusDISABLED
	default:
		return repository.ListingFilterWatch{}, domain.ErrInvalidWatchStatus
	}

	tx, err := s.pool.Begin(ctx)
	if err != nil {
		return repository.ListingFilterWatch{}, err
	}
	defer func() { _ = tx.Rollback(ctx) }()
	qtx := repository.New(tx)

	cur, err := qtx.GetFilterWatchForUser(ctx, repository.GetFilterWatchForUserParams{ID: watchID, UserID: userID})
	if err != nil {
		if err == pgx.ErrNoRows {
			return repository.ListingFilterWatch{}, domain.ErrWatchNotFound
		}
		return repository.ListingFilterWatch{}, err
	}

	if status == repository.ListingWatchStatusACTIVE &&
		cur.Status == repository.ListingWatchStatusDISABLED {
		sub, err := qtx.GetActiveSubscriptionForUpdate(ctx, userID)
		if err != nil {
			if err == pgx.ErrNoRows {
				return repository.ListingFilterWatch{}, domain.ErrNoActiveSubscription
			}
			return repository.ListingFilterWatch{}, err
		}
		count, err := qtx.CountUserWatches(ctx, userID)
		if err != nil {
			return repository.ListingFilterWatch{}, err
		}
		if count >= int64(sub.MaxWatches) {
			return repository.ListingFilterWatch{}, domain.ErrWatchLimitExceeded
		}
	}

	watch, err := qtx.UpdateFilterWatchStatus(ctx, repository.UpdateFilterWatchStatusParams{
		ID:     watchID,
		Status: status,
		UserID: userID,
	})
	if err != nil {
		if err == pgx.ErrNoRows {
			return repository.ListingFilterWatch{}, domain.ErrWatchNotFound
		}
		return repository.ListingFilterWatch{}, err
	}
	if err := tx.Commit(ctx); err != nil {
		return repository.ListingFilterWatch{}, err
	}
	return watch, nil
}

func (s *FilterWatches) Disable(ctx context.Context, userID, watchID pgtype.UUID) error {
	_, err := s.UpdateStatus(ctx, userID, watchID, tokens.ListingWatchStatusDisabled)
	return err
}

func (s *FilterWatches) ResetBaseline(ctx context.Context, userID, watchID pgtype.UUID) error {
	tx, err := s.pool.Begin(ctx)
	if err != nil {
		return err
	}
	defer func() { _ = tx.Rollback(ctx) }()
	qtx := repository.New(tx)

	watch, err := qtx.GetFilterWatchForUser(ctx, repository.GetFilterWatchForUserParams{ID: watchID, UserID: userID})
	if err != nil {
		if err == pgx.ErrNoRows {
			return domain.ErrWatchNotFound
		}
		return err
	}
	if err := qtx.DeleteFilterSeenForUserFilter(ctx, repository.DeleteFilterSeenForUserFilterParams{
		UserID:    userID,
		FilterKey: watch.FilterKey,
	}); err != nil {
		return err
	}
	if err := qtx.ClearWatchHits(ctx, watch.ID); err != nil {
		return err
	}
	if err := qtx.ResetFilterWatchBaseline(ctx, watch.ID); err != nil {
		return err
	}
	return tx.Commit(ctx)
}

func (s *FilterWatches) GetForUser(ctx context.Context, userID, watchID pgtype.UUID) (repository.ListingFilterWatch, error) {
	watch, err := repository.New(s.pool).GetFilterWatchForUser(ctx, repository.GetFilterWatchForUserParams{
		ID:     watchID,
		UserID: userID,
	})
	if err != nil {
		if err == pgx.ErrNoRows {
			return repository.ListingFilterWatch{}, domain.ErrWatchNotFound
		}
		return repository.ListingFilterWatch{}, err
	}
	return watch, nil
}

func (s *FilterWatches) ListResultsForUser(ctx context.Context, userID, watchID pgtype.UUID) ([]repository.ListWatchHitsRow, error) {
	if _, err := s.GetForUser(ctx, userID, watchID); err != nil {
		return nil, err
	}
	return repository.New(s.pool).ListWatchHits(ctx, watchID)
}

func (s *FilterWatches) ListForUser(ctx context.Context, userID pgtype.UUID) ([]repository.ListingFilterWatch, error) {
	return repository.New(s.pool).ListFilterWatchesByUser(ctx, userID)
}
