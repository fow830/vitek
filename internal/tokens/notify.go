package tokens

import "time"

// Notification / outbox / telegram tokens (wave D — Telegram bot contracted).
const (
	NotificationChannelTelegram = "TELEGRAM"

	NotificationOutboxPending   = "PENDING"
	NotificationOutboxSubmitted = "SUBMITTED"
	NotificationOutboxDone      = "DONE"
	NotificationOutboxFailed    = "FAILED"

	JSONFieldTelegramChatID = "telegram_chat_id"
	JSONFieldEnabled        = "enabled"

	ErrMsgNotificationsFailed = "notifications failed"

	NotifyWatchHitFormat = "%s\n%s"

	TelegramBotUsername    = "@vitekapp_bot"
	TelegramAPIBase        = "https://api.telegram.org"
	TelegramAPISendMessage = "sendMessage"
	TelegramAPIGetUpdates  = "getUpdates"
	TelegramPollTimeoutSec = 25
	TelegramPollClientSlack = 10 * time.Second
	TelegramHTTPTimeout    = 35 * time.Second

	TelegramJSONChatID      = "chat_id"
	TelegramJSONText        = "text"
	TelegramJSONReplyMarkup = "reply_markup"
	TelegramJSONOK          = "ok"
	TelegramJSONAllowedUpdates = "allowed_updates"
	TelegramJSONOffset      = "offset"
	TelegramJSONTimeout     = "timeout"

	ErrMsgTelegramTokenMissing = "telegram token missing"
	ErrMsgTelegramAPINil       = "nil telegram api"
	ErrMsgTelegramEmptyChatID  = "empty chat id"
	ErrMsgTelegramSendFailed   = "telegram sendMessage failed"
	ErrMsgTelegramUpdatesFailed = "telegram getUpdates failed"
	ErrMsgTelegramUpdatesNotOK  = "telegram getUpdates not ok"

	TGCommandStart  = "/start"
	TGCommandStop   = "/stop"
	TGCommandPause  = "/pause"
	TGCommandResume = "/resume"

	TGBtnResume  = "Запустить"
	TGBtnStop    = "Остановить"
	TGBtnNewLink = "Новая ссылка"
	TGBtnPause   = "Пауза" // legacy slash/compat; not on keyboard

	TGCopyHelp = "Управление мониторингом Avito:\n" +
		"• Запустить — включить парсинг\n" +
		"• Остановить — выключить парсинг\n" +
		"• Новая ссылка — заменить фильтр"

	TGCopyWatchStarted   = "Мониторинг включён. Новые объявления будут приходить в этот чат."
	TGCopyWatchStopped   = "Парсинг остановлен. Нажми «Запустить» или пришли новую ссылку."
	TGCopyWatchPaused    = "Парсинг на паузе. Нажми «Запустить», чтобы продолжить."
	TGCopyWatchResumed   = "Парсинг снова запущен."
	TGCopyWatchNone      = "Активного мониторинга нет. Пришли ссылку на фильтр Avito."
	TGCopyNeedURLToStart = "Нечего запускать. Пришли ссылку на фильтр Avito."
	TGCopyAwaitNewLink   = "Пришли ссылку на фильтр Avito (страница поиска)."
	TGCopyNeedFilterURL  = "Нужна ссылка на фильтр Avito (страница поиска), не карточку объявления."
	TGCopyWatchFailed    = "Не удалось выполнить действие."
	LogWorkerTelegramBot = "telegram_bot: %v"

	// AllowedGoModNotifyFragments may appear in go.mod after CONTRACT-DAY0-002b.
	GoModRedisClient    = "github.com/redis/go-redis"
	GoModTelegramBotAPI = "github.com/go-telegram-bot-api/telegram-bot-api"
)

// TelegramUserEmail builds a stable platform email for a Telegram chat.
func TelegramUserEmail(chatID string) string {
	return "tg-" + chatID + "@" + ProductDomain
}

// TelegramControlKeyboard is the persistent reply keyboard for watch controls.
func TelegramControlKeyboard() map[string]any {
	return map[string]any{
		"keyboard": [][]map[string]string{
			{{"text": TGBtnResume}, {"text": TGBtnStop}},
			{{"text": TGBtnNewLink}},
		},
		"resize_keyboard":   true,
		"one_time_keyboard": false,
	}
}

var SchemaNotificationChannels = []string{NotificationChannelTelegram}

var SchemaNotificationOutboxStatuses = []string{
	NotificationOutboxPending,
	NotificationOutboxSubmitted,
	NotificationOutboxDone,
	NotificationOutboxFailed,
}
