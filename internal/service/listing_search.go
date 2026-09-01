package service

import (
	"context"
	"errors"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"vitek/internal/repository"
)

// ListingSearchWorker claims PENDING tasks and stores similar items.
type ListingSearchWorker struct {
	pool      *pgxpool.Pool
	items     *Items
	processor ListingProcessor
}

func NewListingSearchWorker(pool *pgxpool.Pool, processor ListingProcessor) *ListingSearchWorker {
	return &ListingSearchWorker{
		pool:      pool,
		items:     NewItems(pool),
		processor: processor,
	}
}

func (w *ListingSearchWorker) ProcessOne(ctx context.Context) (bool, error) {
	tx, err := w.pool.Begin(ctx)
	if err != nil {
		return false, err
	}
	defer func() { _ = tx.Rollback(ctx) }()

	qtx := repository.New(tx)
	task, err := qtx.ClaimNextPendingTask(ctx)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return false, nil
		}
		return false, err
	}

	similar, err := w.processor.FindSimilar(ctx, task.Query)
	if err != nil {
		if _, updErr := qtx.UpdateTaskStatus(ctx, repository.UpdateTaskStatusParams{
			ID:     task.ID,
			Status: repository.TaskStatusFAILED,
		}); updErr != nil {
			return false, updErr
		}
		if err := tx.Commit(ctx); err != nil {
			return false, err
		}
		return true, nil
	}

	for rank, hit := range similar {
		item, recErr := w.items.UpsertTx(ctx, qtx, hit.AvitoID, hit.Title)
		if recErr != nil {
			return false, recErr
		}
		if err := qtx.InsertTaskItem(ctx, repository.InsertTaskItemParams{
			TaskID: task.ID,
			ItemID: item.ID,
			Rank:   int32(rank),
		}); err != nil {
			return false, err
		}
	}

	if _, err := qtx.UpdateTaskStatus(ctx, repository.UpdateTaskStatusParams{
		ID:     task.ID,
		Status: repository.TaskStatusCOMPLETED,
	}); err != nil {
		return false, err
	}
	if err := tx.Commit(ctx); err != nil {
		return false, err
	}
	return true, nil
}
