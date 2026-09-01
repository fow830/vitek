package service

import (
	"context"
	"errors"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/jackc/pgx/v5/pgxpool"

	"vitek/internal/domain"
	"vitek/internal/repository"
)

type AvitoAccounts struct {
	pool *pgxpool.Pool
	q    *repository.Queries
}

func NewAvitoAccounts(pool *pgxpool.Pool) *AvitoAccounts {
	return &AvitoAccounts{pool: pool, q: repository.New(pool)}
}

func (s *AvitoAccounts) Create(ctx context.Context, label string, status repository.AvitoAccountStatus, ref string) (repository.AvitoAccount, error) {
	return s.q.CreateAvitoAccount(ctx, repository.CreateAvitoAccountParams{
		Label:       label,
		Status:      status,
		ExternalRef: ref,
	})
}

func (s *AvitoAccounts) CreateWithSecret(ctx context.Context, label string, status repository.AvitoAccountStatus, login, password string) (repository.AvitoAccount, error) {
	tx, err := s.pool.Begin(ctx)
	if err != nil {
		return repository.AvitoAccount{}, err
	}
	defer func() { _ = tx.Rollback(ctx) }()
	qtx := repository.New(tx)

	acc, err := qtx.CreateAvitoAccount(ctx, repository.CreateAvitoAccountParams{
		Label:       label,
		Status:      status,
		ExternalRef: login,
	})
	if err != nil {
		return repository.AvitoAccount{}, err
	}
	if err := qtx.UpsertAvitoAccountSecret(ctx, repository.UpsertAvitoAccountSecretParams{
		AccountID: acc.ID,
		Password:  password,
	}); err != nil {
		return repository.AvitoAccount{}, err
	}
	if err := tx.Commit(ctx); err != nil {
		return repository.AvitoAccount{}, err
	}
	return acc, nil
}

func (s *AvitoAccounts) UpsertSecret(ctx context.Context, accountID pgtype.UUID, password string) error {
	return s.q.UpsertAvitoAccountSecret(ctx, repository.UpsertAvitoAccountSecretParams{
		AccountID: accountID,
		Password:  password,
	})
}

func (s *AvitoAccounts) PickActive(ctx context.Context) (repository.PickActiveAvitoAccountRow, error) {
	row, err := s.q.PickActiveAvitoAccount(ctx)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return repository.PickActiveAvitoAccountRow{}, domain.ErrListingSearchNoAccount
		}
		return repository.PickActiveAvitoAccountRow{}, err
	}
	return row, nil
}

func (s *AvitoAccounts) List(ctx context.Context) ([]repository.AvitoAccount, error) {
	return s.q.ListAvitoAccounts(ctx)
}

func (s *AvitoAccounts) Update(ctx context.Context, id pgtype.UUID, label string, status repository.AvitoAccountStatus, ref string) (repository.AvitoAccount, error) {
	return s.q.UpdateAvitoAccount(ctx, repository.UpdateAvitoAccountParams{
		ID:          id,
		Label:       label,
		Status:      status,
		ExternalRef: ref,
	})
}

func (s *AvitoAccounts) Delete(ctx context.Context, id pgtype.UUID) error {
	_, err := s.q.DeleteAvitoAccount(ctx, id)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return domain.ErrResourceNotFound
		}
		return err
	}
	return nil
}

func (s *AvitoAccounts) Count(ctx context.Context) (int64, error) {
	return s.q.CountAvitoAccounts(ctx)
}
