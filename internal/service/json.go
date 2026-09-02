package service

import (
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgtype"

	"vitek/internal/repository"
	"vitek/internal/tokens"
)

func UUIDString(id pgtype.UUID) string {
	if !id.Valid {
		return ""
	}
	return uuid.UUID(id.Bytes).String()
}

func TaskJSON(task Task) map[string]any {
	return map[string]any{
		tokens.JSONFieldID:     UUIDString(task.ID),
		tokens.JSONFieldUserID: UUIDString(task.UserID),
		tokens.JSONFieldQuery:  task.Query,
		tokens.JSONFieldStatus: string(task.Status),
		tokens.JSONFieldKind:   tokens.ListingSearchKindTask,
	}
}

func TaskResultJSON(it ListTaskItemsRow) map[string]any {
	return map[string]any{
		tokens.JSONFieldID:      UUIDString(it.ID),
		tokens.JSONFieldAvitoID: it.AvitoID,
		tokens.JSONFieldTitle:   it.Title,
		tokens.JSONFieldRank:    it.Rank,
	}
}

func WatchJSON(watch repository.ListingFilterWatch) map[string]any {
	meta := tokens.ParseListingFilterMeta(watch.Query)
	return map[string]any{
		tokens.JSONFieldID:         UUIDString(watch.ID),
		tokens.JSONFieldUserID:     UUIDString(watch.UserID),
		tokens.JSONFieldQuery:      watch.Query,
		tokens.JSONFieldStatus:     string(watch.Status),
		tokens.JSONFieldKind:       tokens.ListingSearchKindWatch,
		tokens.JSONFieldRegion:     meta.Region,
		tokens.JSONFieldCategories: meta.Categories,
		tokens.JSONFieldLabel:      meta.Label,
		tokens.JSONFieldParams:     meta.Params,
		tokens.JSONFieldExtras:     meta.Extras,
	}
}

func WatchHitJSON(it repository.ListWatchHitsRow) map[string]any {
	return map[string]any{
		tokens.JSONFieldID:      UUIDString(it.ID),
		tokens.JSONFieldAvitoID: it.AvitoID,
		tokens.JSONFieldTitle:   it.Title,
		tokens.JSONFieldFoundAt: it.FoundAt.Time,
	}
}
