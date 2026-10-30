package service

import (
	"context"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/jackc/pgx/v5/pgxpool"

	"vitek/internal/domain"
	"vitek/internal/repository"
	"vitek/internal/tokens"
)

// TelegramBotReply is a bot answer shown with the control keyboard.
type TelegramBotReply struct {
	Text string
}

// TelegramBot turns chat messages into listing_search watches + notification binding.
type TelegramBot struct {
	pool      *pgxpool.Pool
	messenger NotificationSender
	watches   *FilterWatches
	users     *Users
	notify    *Notifications
}

func NewTelegramBot(pool *pgxpool.Pool, messenger NotificationSender) *TelegramBot {
	return &TelegramBot{
		pool:      pool,
		messenger: messenger,
		watches:   NewFilterWatches(pool),
		users:     NewUsers(pool),
		notify:    NewNotifications(pool, messenger),
	}
}

// HandleText processes one inbound Telegram text message and returns the reply body.
func (b *TelegramBot) HandleText(ctx context.Context, chatID, text string) (TelegramBotReply, error) {
	chatID = strings.TrimSpace(chatID)
	text = strings.TrimSpace(text)
	if chatID == "" {
		return TelegramBotReply{}, fmt.Errorf("%s", tokens.ErrMsgTelegramEmptyChatID)
	}

	user, err := b.ensureUser(ctx, chatID)
	if err != nil {
		return TelegramBotReply{}, err
	}
	if _, err := b.notify.UpsertSettings(ctx, user.ID, chatID, true); err != nil {
		return TelegramBotReply{}, err
	}

	cmd := strings.Split(text, "@")[0]
	switch cmd {
	case tokens.TGCommandStart, "":
		return TelegramBotReply{Text: tokens.TGCopyHelp}, nil
	case tokens.TGCommandStop, tokens.TGBtnStop:
		return b.stopWatches(ctx, user.ID)
	case tokens.TGCommandPause, tokens.TGBtnPause:
		return b.pauseWatches(ctx, user.ID)
	case tokens.TGCommandResume, tokens.TGBtnResume:
		return b.resumeWatches(ctx, user.ID)
	case tokens.TGBtnNewLink:
		return TelegramBotReply{Text: tokens.TGCopyAwaitNewLink}, nil
	}

	if !tokens.IsListingFilterURL(text) {
		return TelegramBotReply{Text: tokens.TGCopyNeedFilterURL}, nil
	}

	if _, err := b.disableActiveWatches(ctx, user.ID); err != nil {
		return TelegramBotReply{Text: tokens.TGCopyWatchFailed + " " + err.Error()}, nil
	}
	if _, err := b.watches.Start(ctx, user.ID, text); err != nil {
		if err == domain.ErrInvalidListingURL {
			return TelegramBotReply{Text: tokens.TGCopyNeedFilterURL}, nil
		}
		return TelegramBotReply{Text: tokens.TGCopyWatchFailed + " " + err.Error()}, nil
	}
	return TelegramBotReply{Text: tokens.TGCopyWatchStarted}, nil
}

func (b *TelegramBot) stopWatches(ctx context.Context, userID pgtype.UUID) (TelegramBotReply, error) {
	n, err := b.disableActiveWatches(ctx, userID)
	if err != nil {
		return TelegramBotReply{}, err
	}
	if n == 0 {
		return TelegramBotReply{Text: tokens.TGCopyWatchNone}, nil
	}
	return TelegramBotReply{Text: tokens.TGCopyWatchStopped}, nil
}

func (b *TelegramBot) pauseWatches(ctx context.Context, userID pgtype.UUID) (TelegramBotReply, error) {
	list, err := b.watches.ListAllForUser(ctx, userID)
	if err != nil {
		return TelegramBotReply{}, err
	}
	n := 0
	for _, w := range list {
		if w.Status != repository.ListingWatchStatusACTIVE {
			continue
		}
		if _, err := b.watches.UpdateStatus(ctx, userID, w.ID, tokens.ListingWatchStatusPaused); err != nil {
			return TelegramBotReply{}, err
		}
		n++
	}
	if n == 0 {
		return TelegramBotReply{Text: tokens.TGCopyWatchNone}, nil
	}
	return TelegramBotReply{Text: tokens.TGCopyWatchPaused}, nil
}

func (b *TelegramBot) resumeWatches(ctx context.Context, userID pgtype.UUID) (TelegramBotReply, error) {
	list, err := b.watches.ListAllForUser(ctx, userID)
	if err != nil {
		return TelegramBotReply{}, err
	}
	n := 0
	for _, w := range list {
		if w.Status != repository.ListingWatchStatusPAUSED {
			continue
		}
		if _, err := b.watches.UpdateStatus(ctx, userID, w.ID, tokens.ListingWatchStatusActive); err != nil {
			return TelegramBotReply{}, err
		}
		n++
	}
	if n > 0 {
		return TelegramBotReply{Text: tokens.TGCopyWatchResumed}, nil
	}
	// No paused watch: revive the newest disabled one if present.
	var newest *repository.ListingFilterWatch
	for i := range list {
		w := list[i]
		if w.Status != repository.ListingWatchStatusDISABLED {
			continue
		}
		if newest == nil || w.CreatedAt.Time.After(newest.CreatedAt.Time) {
			cp := w
			newest = &cp
		}
	}
	if newest == nil {
		return TelegramBotReply{Text: tokens.TGCopyNeedURLToStart}, nil
	}
	if _, err := b.watches.UpdateStatus(ctx, userID, newest.ID, tokens.ListingWatchStatusActive); err != nil {
		return TelegramBotReply{}, err
	}
	return TelegramBotReply{Text: tokens.TGCopyWatchResumed}, nil
}

func (b *TelegramBot) disableActiveWatches(ctx context.Context, userID pgtype.UUID) (int, error) {
	list, err := b.watches.ListAllForUser(ctx, userID)
	if err != nil {
		return 0, err
	}
	n := 0
	for _, w := range list {
		if w.Status != repository.ListingWatchStatusACTIVE && w.Status != repository.ListingWatchStatusPAUSED {
			continue
		}
		if err := b.watches.Disable(ctx, userID, w.ID); err != nil {
			return n, err
		}
		n++
	}
	return n, nil
}

func (b *TelegramBot) ensureUser(ctx context.Context, chatID string) (repository.User, error) {
	q := repository.New(b.pool)
	settings, err := q.GetNotificationSettingsByTelegramChatID(ctx, &chatID)
	if err == nil {
		user, err := q.GetUserByID(ctx, settings.UserID)
		if err == nil {
			return user, nil
		}
		if err != pgx.ErrNoRows {
			return repository.User{}, err
		}
	} else if err != pgx.ErrNoRows {
		return repository.User{}, err
	}

	email := tokens.TelegramUserEmail(chatID)
	if u, err := q.GetUserByEmail(ctx, email); err == nil {
		return u, nil
	} else if err != pgx.ErrNoRows {
		return repository.User{}, err
	}
	return b.users.CreateUser(ctx, email, repository.PlanTypePRO)
}

// RunLongPoll blocks until ctx is done, forwarding updates to HandleText.
func (b *TelegramBot) RunLongPoll(ctx context.Context, api *TelegramAPIClient) error {
	if api == nil {
		return fmt.Errorf("%s", tokens.ErrMsgTelegramAPINil)
	}
	offset := 0
	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
		}
		updates, err := api.GetUpdates(ctx, offset, tokens.TelegramPollTimeoutSec)
		if err != nil {
			if ctx.Err() != nil {
				return ctx.Err()
			}
			select {
			case <-ctx.Done():
				return ctx.Err()
			case <-time.After(3 * time.Second):
			}
			continue
		}
		for _, u := range updates {
			if u.UpdateID >= offset {
				offset = u.UpdateID + 1
			}
			if u.Message == nil || strings.TrimSpace(u.Message.Text) == "" {
				continue
			}
			chatID := strconv.FormatInt(u.Message.Chat.ID, 10)
			reply, err := b.HandleText(ctx, chatID, u.Message.Text)
			if err != nil {
				_ = api.SendWithKeyboard(ctx, chatID, tokens.TGCopyWatchFailed)
				continue
			}
			_ = api.SendWithKeyboard(ctx, chatID, reply.Text)
		}
	}
}
