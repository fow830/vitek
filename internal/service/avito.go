package service

import (
	"context"

	"github.com/jackc/pgx/v5/pgtype"
	"github.com/jackc/pgx/v5/pgxpool"

	"vitek/internal/repository"
)

type AvitoAccounts struct {
	q *repository.Queries
}

func NewAvitoAccounts(pool *pgxpool.Pool) *AvitoAccounts {
	return &AvitoAccounts{q: repository.New(pool)}
}

func (s *AvitoAccounts) Create(ctx context.Context, label string, status repository.AvitoAccountStatus, ref string) (repository.AvitoAccount, error) {
	return s.q.CreateAvitoAccount(ctx, repository.CreateAvitoAccountParams{
		Label:       label,
		Status:      status,
		ExternalRef: ref,
	})
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

func (s *AvitoAccounts) Count(ctx context.Context) (int64, error) {
	return s.q.CountAvitoAccounts(ctx)
}
