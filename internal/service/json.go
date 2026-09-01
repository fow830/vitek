package service

import (
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgtype"

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
