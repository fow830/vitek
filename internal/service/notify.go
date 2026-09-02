package service

import (
	"context"
	"fmt"
	"strings"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/jackc/pgx/v5/pgxpool"

	"vitek/internal/repository"
	"vitek/internal/tokens"
)

// NotificationSender delivers outbox messages (Telegram adapter; stub in contracts).
type NotificationSender interface {
	Send(ctx context.Context, chatID, text string) error
}

// LogNotificationSender records sends for tests.
type LogNotificationSender struct {
	Sent []string
}

func (s *LogNotificationSender) Send(_ context.Context, chatID, text string) error {
	s.Sent = append(s.Sent, chatID+":"+text)
	return nil
}

// Notifications manages settings + outbox delivery.
type Notifications struct {
	pool   *pgxpool.Pool
	sender NotificationSender
}

func NewNotifications(pool *pgxpool.Pool, sender NotificationSender) *Notifications {
	if sender == nil {
		sender = &LogNotificationSender{}
	}
	return &Notifications{pool: pool, sender: sender}
}

func (s *Notifications) UpsertSettings(ctx context.Context, userID pgtype.UUID, chatID string, enabled bool) (repository.UserNotificationSetting, error) {
	chatID = strings.TrimSpace(chatID)
	var chatPtr *string
	if chatID != "" {
		chatPtr = &chatID
	}
	return repository.New(s.pool).UpsertNotificationSettings(ctx, repository.UpsertNotificationSettingsParams{
		UserID:         userID,
		TelegramChatID: chatPtr,
		Enabled:        enabled,
	})
}

func (s *Notifications) GetSettings(ctx context.Context, userID pgtype.UUID) (repository.UserNotificationSetting, error) {
	row, err := repository.New(s.pool).GetNotificationSettings(ctx, userID)
	if err != nil {
		if err == pgx.ErrNoRows {
			return repository.UserNotificationSetting{UserID: userID, Enabled: false}, nil
		}
		return repository.UserNotificationSetting{}, err
	}
	return row, nil
}

func (s *Notifications) EnqueueWatchHit(ctx context.Context, qtx *repository.Queries, userID, watchID, itemID pgtype.UUID) error {
	return qtx.EnqueueNotification(ctx, repository.EnqueueNotificationParams{
		UserID:  userID,
		WatchID: watchID,
		ItemID:  itemID,
	})
}

func (s *Notifications) ProcessOutbox(ctx context.Context, limit int32) (int, error) {
	tx, err := s.pool.Begin(ctx)
	if err != nil {
		return 0, err
	}
	defer func() { _ = tx.Rollback(ctx) }()
	qtx := repository.New(tx)
	rows, err := qtx.ClaimPendingNotifications(ctx, limit)
	if err != nil {
		return 0, err
	}
	n := 0
	for _, row := range rows {
		if !row.NotifyEnabled || row.TelegramChatID == nil || strings.TrimSpace(*row.TelegramChatID) == "" {
			if err := qtx.MarkNotificationDone(ctx, row.ID); err != nil {
				return n, err
			}
			n++
			continue
		}
		if err := qtx.MarkNotificationSubmitted(ctx, row.ID); err != nil {
			return n, err
		}
		text := fmt.Sprintf(tokens.NotifyWatchHitFormat, row.ItemTitle, row.ItemAvitoID)
		if err := s.sender.Send(ctx, *row.TelegramChatID, text); err != nil {
			msg := err.Error()
			_ = qtx.MarkNotificationFailed(ctx, repository.MarkNotificationFailedParams{
				ID:        row.ID,
				LastError: &msg,
			})
			continue
		}
		if err := qtx.MarkNotificationDone(ctx, row.ID); err != nil {
			return n, err
		}
		n++
	}
	if err := tx.Commit(ctx); err != nil {
		return n, err
	}
	return n, nil
}

func NotificationSettingsJSON(row repository.UserNotificationSetting) map[string]any {
	out := map[string]any{
		tokens.JSONFieldUserID:  UUIDString(row.UserID),
		tokens.JSONFieldEnabled: row.Enabled,
	}
	if row.TelegramChatID != nil {
		out[tokens.JSONFieldTelegramChatID] = *row.TelegramChatID
	} else {
		out[tokens.JSONFieldTelegramChatID] = ""
	}
	return out
}
