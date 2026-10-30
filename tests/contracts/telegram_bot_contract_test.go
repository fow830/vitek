package contracts_test

import (
	"context"
	"strconv"
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	"vitek/internal/repository"
	"vitek/internal/service"
	"vitek/internal/tokens"
)

type contractTGSender struct {
	Sent []string
}

func (s *contractTGSender) Send(_ context.Context, chatID, text string) error {
	s.Sent = append(s.Sent, chatID+":"+text)
	return nil
}

// CONTRACT-TG-001: Telegram filter URL starts listing_search watch and binds chat notifications.
func TestContract_TelegramBot_StartsWatchFromFilterURL(t *testing.T) {
	ctx := context.Background()
	pool, q := queries(t)

	sender := &contractTGSender{}
	bot := service.NewTelegramBot(pool, sender)

	chatID := "424242"
	reply, err := bot.HandleText(ctx, chatID, tokens.FixtureFilterListingURL)
	require.NoError(t, err)
	require.Contains(t, reply.Text, tokens.TGCopyWatchStarted)

	settings, err := q.GetNotificationSettingsByTelegramChatID(ctx, &chatID)
	require.NoError(t, err)
	require.True(t, settings.Enabled)
	require.NotNil(t, settings.TelegramChatID)
	require.Equal(t, chatID, *settings.TelegramChatID)

	watches, err := q.ListFilterWatchesByUser(ctx, settings.UserID)
	require.NoError(t, err)
	require.NotEmpty(t, watches)
	require.Equal(t, repository.ListingWatchStatusACTIVE, watches[0].Status)
}

// CONTRACT-TG-002: /start binds chat and returns help; invalid URL is rejected in plain language.
func TestContract_TelegramBot_StartAndInvalidURL(t *testing.T) {
	ctx := context.Background()
	pool, _ := queries(t)
	sender := &contractTGSender{}
	bot := service.NewTelegramBot(pool, sender)

	chatID := "777001"
	reply, err := bot.HandleText(ctx, chatID, tokens.TGCommandStart)
	require.NoError(t, err)
	require.Contains(t, reply.Text, tokens.TGCopyHelp)
	require.Contains(t, reply.Text, tokens.TGBtnResume)
	require.Contains(t, reply.Text, tokens.TGBtnStop)
	require.Contains(t, reply.Text, tokens.TGBtnNewLink)

	bad, err := bot.HandleText(ctx, chatID, "https://example.com/not-avito")
	require.NoError(t, err)
	require.Contains(t, bad.Text, tokens.TGCopyNeedFilterURL)
}

// CONTRACT-TG-005: /stop disables active watches; a new filter URL replaces the previous one.
func TestContract_TelegramBot_StopAndReplaceWatch(t *testing.T) {
	ctx := context.Background()
	pool, q := queries(t)
	sender := &contractTGSender{}
	bot := service.NewTelegramBot(pool, sender)

	chatID := "888002"
	_, err := bot.HandleText(ctx, chatID, tokens.FixtureFilterListingURL)
	require.NoError(t, err)

	settings, err := q.GetNotificationSettingsByTelegramChatID(ctx, &chatID)
	require.NoError(t, err)
	watches, err := q.ListFilterWatchesByUser(ctx, settings.UserID)
	require.NoError(t, err)
	require.Equal(t, repository.ListingWatchStatusACTIVE, watches[0].Status)

	stopReply, err := bot.HandleText(ctx, chatID, tokens.TGBtnStop)
	require.NoError(t, err)
	require.Contains(t, stopReply.Text, tokens.TGCopyWatchStopped)
	watches, err = q.ListAllFilterWatchesByUser(ctx, settings.UserID)
	require.NoError(t, err)
	require.Equal(t, repository.ListingWatchStatusDISABLED, watches[0].Status)

	_, err = bot.HandleText(ctx, chatID, tokens.FixtureFilterListingURL)
	require.NoError(t, err)
	_, err = bot.HandleText(ctx, chatID, tokens.FixtureFilterListingURL2)
	require.NoError(t, err)
	watches, err = q.ListFilterWatchesByUser(ctx, settings.UserID)
	require.NoError(t, err)
	active := 0
	for _, w := range watches {
		if w.Status == repository.ListingWatchStatusACTIVE {
			active++
			require.Contains(t, w.Query, "moskva")
		}
	}
	require.Equal(t, 1, active)
}

// CONTRACT-TG-006: stop then start buttons control ACTIVE ↔ DISABLED parsing.
func TestContract_TelegramBot_StopAndStart(t *testing.T) {
	ctx := context.Background()
	pool, q := queries(t)
	sender := &contractTGSender{}
	bot := service.NewTelegramBot(pool, sender)

	chatID := "999003"
	_, err := bot.HandleText(ctx, chatID, tokens.FixtureFilterListingURL)
	require.NoError(t, err)

	settings, err := q.GetNotificationSettingsByTelegramChatID(ctx, &chatID)
	require.NoError(t, err)

	stopReply, err := bot.HandleText(ctx, chatID, tokens.TGBtnStop)
	require.NoError(t, err)
	require.Contains(t, stopReply.Text, tokens.TGCopyWatchStopped)
	watches, err := q.ListAllFilterWatchesByUser(ctx, settings.UserID)
	require.NoError(t, err)
	require.Equal(t, repository.ListingWatchStatusDISABLED, watches[0].Status)

	startReply, err := bot.HandleText(ctx, chatID, tokens.TGBtnResume)
	require.NoError(t, err)
	require.Contains(t, startReply.Text, tokens.TGCopyWatchResumed)
	watches, err = q.ListFilterWatchesByUser(ctx, settings.UserID)
	require.NoError(t, err)
	require.Equal(t, repository.ListingWatchStatusACTIVE, watches[0].Status)

	newLink, err := bot.HandleText(ctx, chatID, tokens.TGBtnNewLink)
	require.NoError(t, err)
	require.Contains(t, newLink.Text, tokens.TGCopyAwaitNewLink)
}

// CONTRACT-TG-003: outbox delivery text includes human title + Avito item page URL.
func TestContract_NotifyOutboxIncludesItemURL(t *testing.T) {
	ctx := context.Background()
	pool, q := queries(t)

	sender := &contractTGSender{}
	notify := service.NewNotifications(pool, sender)

	user, err := service.NewUsers(pool).CreateUser(ctx, tokens.ProductEmail("tg-notify"), repository.PlanTypePRO)
	require.NoError(t, err)
	_, err = notify.UpsertSettings(ctx, user.ID, "555", true)
	require.NoError(t, err)

	watch, err := service.NewFilterWatches(pool).Start(ctx, user.ID, tokens.FixtureFilterListingURL)
	require.NoError(t, err)
	item, err := q.UpsertItem(ctx, repository.UpsertItemParams{
		AvitoID: tokens.FixtureAvitoItemID1,
		Title:   tokens.FixtureAvitoItemTitle1,
	})
	require.NoError(t, err)
	require.NoError(t, notify.EnqueueWatchHit(ctx, q, user.ID, watch.ID, item.ID))

	n, err := notify.ProcessOutbox(ctx, tokens.ListingSearchWatchDueLimit)
	require.NoError(t, err)
	require.Equal(t, 1, n)
	require.Len(t, sender.Sent, 1)
	require.Contains(t, sender.Sent[0], tokens.FixtureAvitoItemTitle1)
	require.Contains(t, sender.Sent[0], tokens.AvitoItemPageURL(tokens.FixtureAvitoItemID1))
}

// countingListingProcessor counts FindSimilar calls (shared stream SoT).
type countingListingProcessor struct {
	inner service.ListingProcessor
	calls int
}

func (p *countingListingProcessor) FindSimilar(ctx context.Context, listingURL string) ([]service.SimilarListing, error) {
	p.calls++
	return p.inner.FindSimilar(ctx, listingURL)
}

// CONTRACT-TG-007: many Telegram users on the same filter share one Avito fetch stream.
func TestContract_TelegramBot_SharedFilterStream(t *testing.T) {
	ctx := context.Background()
	pool, q := queries(t)
	mock := startAvitoMock(t)
	t.Cleanup(mock.Close)

	sender := &contractTGSender{}
	bot := service.NewTelegramBot(pool, sender)
	_, err := bot.HandleText(ctx, "100001", tokens.FixtureFilterListingURL)
	require.NoError(t, err)
	_, err = bot.HandleText(ctx, "100002", tokens.FixtureFilterListingURL)
	require.NoError(t, err)

	s1, err := q.GetNotificationSettingsByTelegramChatID(ctx, strPtr("100001"))
	require.NoError(t, err)
	s2, err := q.GetNotificationSettingsByTelegramChatID(ctx, strPtr("100002"))
	require.NoError(t, err)
	require.NotEqual(t, s1.UserID, s2.UserID)

	_, err = service.NewProxies(pool).Create(ctx, mock.URL, repository.ProxyStatusACTIVE, "shared-proxy")
	require.NoError(t, err)
	_, err = service.NewAvitoAccounts(pool).CreateWithSecret(ctx, "shared-acc", repository.AvitoAccountStatusACTIVE, "a@t", "x")
	require.NoError(t, err)

	counter := &countingListingProcessor{inner: service.NewAvitoListingProcessor(pool, service.NewAvitoClient(service.WithAvitoHTTPBase(mock.URL)))}
	worker := service.NewListingSearchWorker(pool, counter)
	n, err := worker.ProcessWatchPolls(ctx)
	require.NoError(t, err)
	require.Equal(t, 2, n)
	require.Equal(t, 1, counter.calls)

	w1, err := q.ListFilterWatchesByUser(ctx, s1.UserID)
	require.NoError(t, err)
	w2, err := q.ListFilterWatchesByUser(ctx, s2.UserID)
	require.NoError(t, err)
	require.Equal(t, w1[0].FilterKey, w2[0].FilterKey)
}

func strPtr(s string) *string { return &s }

// CONTRACT-TG-004: bot identity + env token + control keyboard live in tokens SoT.
func TestContract_TelegramBot_Tokens(t *testing.T) {
	require.Equal(t, "@vitekapp_bot", tokens.TelegramBotUsername)
	require.Equal(t, "TELEGRAM_BOT_TOKEN", tokens.EnvTelegramBotToken)
	require.Contains(t, tokens.EnvExampleKeys, tokens.EnvTelegramBotToken)
	require.Equal(t, tokens.TelegramUserEmail("42"), "tg-42@"+tokens.ProductDomain)
	require.Equal(t, "https://api.telegram.org", tokens.TelegramAPIBase)
	require.Equal(t, "sendMessage", tokens.TelegramAPISendMessage)
	require.Equal(t, "getUpdates", tokens.TelegramAPIGetUpdates)
	require.Equal(t, 10*time.Second, tokens.TelegramPollClientSlack)
	require.Equal(t, "telegram token missing", tokens.ErrMsgTelegramTokenMissing)
	require.Equal(t, "nil telegram api", tokens.ErrMsgTelegramAPINil)
	require.Equal(t, "empty chat id", tokens.ErrMsgTelegramEmptyChatID)
	kb := tokens.TelegramControlKeyboard()
	require.Contains(t, kb, "keyboard")
	require.Equal(t, tokens.TGBtnResume, "Запустить")
	require.Equal(t, tokens.TGBtnStop, "Остановить")
	require.Equal(t, tokens.TGBtnNewLink, "Новая ссылка")
	rows, ok := kb["keyboard"].([][]map[string]string)
	require.True(t, ok)
	require.Len(t, rows, 2)
	require.Equal(t, tokens.TGBtnResume, rows[0][0]["text"])
	require.Equal(t, tokens.TGBtnStop, rows[0][1]["text"])
	require.Equal(t, tokens.TGBtnNewLink, rows[1][0]["text"])
	_ = strconv.Itoa(0)
}
