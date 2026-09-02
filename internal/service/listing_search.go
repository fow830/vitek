package service

import (
	"context"
	"encoding/json"
	"errors"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/jackc/pgx/v5/pgxpool"

	"vitek/internal/repository"
	"vitek/internal/tokens"
)

// ListingSearchWorker claims PENDING tasks and stores similar items.
type ListingSearchWorker struct {
	pool      *pgxpool.Pool
	items     *Items
	processor ListingProcessor
	notify    *Notifications
}

func NewListingSearchWorker(pool *pgxpool.Pool, processor ListingProcessor) *ListingSearchWorker {
	return &ListingSearchWorker{
		pool:      pool,
		items:     NewItems(pool),
		processor: processor,
		notify:    NewNotifications(pool, nil),
	}
}

func NewListingSearchWorkerWithNotify(pool *pgxpool.Pool, processor ListingProcessor, notify *Notifications) *ListingSearchWorker {
	w := NewListingSearchWorker(pool, processor)
	if notify != nil {
		w.notify = notify
	}
	return w
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

func (w *ListingSearchWorker) ProcessWatchPolls(ctx context.Context) (int, error) {
	watches, err := repository.New(w.pool).ListDueFilterWatches(ctx)
	if err != nil {
		return 0, err
	}

	n := 0
	for _, watch := range watches {
		if err := w.pollWatch(ctx, watch); err != nil {
			return n, err
		}
		n++
	}
	return n, nil
}

func (w *ListingSearchWorker) pollWatch(ctx context.Context, watch repository.ListingFilterWatch) error {
	similar, findErr := w.processor.FindSimilar(ctx, watch.Query)

	tx, err := w.pool.Begin(ctx)
	if err != nil {
		return err
	}
	defer func() { _ = tx.Rollback(ctx) }()
	qtx := repository.New(tx)

	if findErr != nil {
		msg := findErr.Error()
		if _, err := qtx.RecordFilterWatchPollFailure(ctx, repository.RecordFilterWatchPollFailureParams{
			ID:          watch.ID,
			LastError:   &msg,
			MaxFailures: int32(tokens.WatchAutoPauseAfterFails),
		}); err != nil {
			return err
		}
		return tx.Commit(ctx)
	}

	filtered := make([]SimilarListing, 0, len(similar))
	for _, hit := range similar {
		if tokens.ListingTitleAllowedForFilter(watch.Query, hit.Title) {
			filtered = append(filtered, hit)
		}
	}
	if len(filtered) > tokens.ListingSearchSERPMaxItems {
		filtered = filtered[:tokens.ListingSearchSERPMaxItems]
	}

	if err := w.applyFilterHits(ctx, qtx, watch.UserID, watch.FilterKey, filtered, func(item repository.Item) error {
		if err := qtx.InsertWatchHit(ctx, repository.InsertWatchHitParams{
			WatchID: watch.ID,
			ItemID:  item.ID,
		}); err != nil {
			return err
		}
		return w.notify.EnqueueWatchHit(ctx, qtx, watch.UserID, watch.ID, item.ID)
	}); err != nil {
		return err
	}
	if err := qtx.TouchFilterWatchPolled(ctx, watch.ID); err != nil {
		return err
	}
	meta := tokens.ParseListingFilterMeta(watch.Query)
	if mp, ok := w.processor.(FilterMetaProvider); ok {
		meta = mp.LastFilterMeta()
	}
	metaBytes, err := json.Marshal(meta)
	if err != nil {
		return err
	}
	if err := qtx.UpdateFilterWatchMeta(ctx, repository.UpdateFilterWatchMetaParams{
		ID:         watch.ID,
		MetaStatus: repository.ListingWatchMetaStatusREADY,
		MetaJson:   metaBytes,
	}); err != nil {
		return err
	}
	return tx.Commit(ctx)
}

func (w *ListingSearchWorker) applyFilterHits(
	ctx context.Context,
	qtx *repository.Queries,
	userID pgtype.UUID,
	filterKey string,
	similar []SimilarListing,
	onNew func(repository.Item) error,
) error {
	seen, err := qtx.ListFilterSeenAvitoIDs(ctx, repository.ListFilterSeenAvitoIDsParams{
		UserID:    userID,
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
	for _, hit := range similar {
		item, recErr := w.items.UpsertTx(ctx, qtx, hit.AvitoID, hit.Title)
		if recErr != nil {
			return recErr
		}
		if err := qtx.InsertFilterSeen(ctx, repository.InsertFilterSeenParams{
			UserID:    userID,
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
		if onNew != nil {
			if err := onNew(item); err != nil {
				return err
			}
		}
	}
	return nil
}

func (w *ListingSearchWorker) completeFilterTask(ctx context.Context, qtx *repository.Queries, task repository.Task, similar []SimilarListing) error {
	filterKey := tokens.CanonicalListingSearchQuery(task.Query)
	rank := 0
	if err := w.applyFilterHits(ctx, qtx, task.UserID, filterKey, similar, func(item repository.Item) error {
		err := qtx.InsertTaskItem(ctx, repository.InsertTaskItemParams{
			TaskID: task.ID,
			ItemID: item.ID,
			Rank:   int32(rank),
		})
		rank++
		return err
	}); err != nil {
		return err
	}
	_, err := qtx.UpdateTaskStatus(ctx, repository.UpdateTaskStatusParams{
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
