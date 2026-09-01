package service

import (
	"context"
	"errors"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"vitek/internal/repository"
	"vitek/internal/tokens"
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

	if tokens.IsListingFilterURL(task.Query) {
		if err := w.completeFilterTask(ctx, qtx, task, similar); err != nil {
			return false, err
		}
	} else if err := w.completeItemTask(ctx, qtx, task, similar); err != nil {
		return false, err
	}

	if err := tx.Commit(ctx); err != nil {
		return false, err
	}
	return true, nil
}

func (w *ListingSearchWorker) completeFilterTask(ctx context.Context, qtx *repository.Queries, task repository.Task, similar []SimilarListing) error {
	filterKey := tokens.CanonicalListingSearchQuery(task.Query)
	seen, err := qtx.ListFilterSeenAvitoIDs(ctx, repository.ListFilterSeenAvitoIDsParams{
		UserID:    task.UserID,
		FilterKey: filterKey,
	})
	if err != nil {
		return err
	}
	seenSet := make(map[string]struct{}, len(seen))
	for _, avitoID := range seen {
		seenSet[avitoID] = struct{}{}
	}

	baseline := len(seenSet) == 0
	rank := 0
	for _, hit := range similar {
		item, recErr := w.items.UpsertTx(ctx, qtx, hit.AvitoID, hit.Title)
		if recErr != nil {
			return recErr
		}
		if err := qtx.InsertFilterSeen(ctx, repository.InsertFilterSeenParams{
			UserID:    task.UserID,
			FilterKey: filterKey,
			AvitoID:   hit.AvitoID,
		}); err != nil {
			return err
		}
		if baseline {
			continue
		}
		if _, ok := seenSet[hit.AvitoID]; ok {
			continue
		}
		if err := qtx.InsertTaskItem(ctx, repository.InsertTaskItemParams{
			TaskID: task.ID,
			ItemID: item.ID,
			Rank:   int32(rank),
		}); err != nil {
			return err
		}
		rank++
	}

	_, err = qtx.UpdateTaskStatus(ctx, repository.UpdateTaskStatusParams{
		ID:     task.ID,
		Status: repository.TaskStatusCOMPLETED,
	})
	return err
}

func (w *ListingSearchWorker) completeItemTask(ctx context.Context, qtx *repository.Queries, task repository.Task, similar []SimilarListing) error {
	for rank, hit := range similar {
		item, recErr := w.items.UpsertTx(ctx, qtx, hit.AvitoID, hit.Title)
		if recErr != nil {
			return recErr
		}
		if err := qtx.InsertTaskItem(ctx, repository.InsertTaskItemParams{
			TaskID: task.ID,
			ItemID: item.ID,
			Rank:   int32(rank),
		}); err != nil {
			return err
		}
	}
	_, err := qtx.UpdateTaskStatus(ctx, repository.UpdateTaskStatusParams{
		ID:     task.ID,
		Status: repository.TaskStatusCOMPLETED,
	})
	return err
}
