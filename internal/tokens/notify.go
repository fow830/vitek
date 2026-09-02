package tokens

// Notification / outbox / telegram tokens (wave D — delivery stub until bot contracted).
const (
	NotificationChannelTelegram = "TELEGRAM"

	NotificationOutboxPending   = "PENDING"
	NotificationOutboxSubmitted = "SUBMITTED"
	NotificationOutboxDone      = "DONE"
	NotificationOutboxFailed    = "FAILED"

	JSONFieldTelegramChatID = "telegram_chat_id"
	JSONFieldEnabled        = "enabled"

	ErrMsgNotificationsFailed = "notifications failed"

	NotifyWatchHitFormat = "%s: %s"

	// AllowedGoModNotifyFragments may appear in go.mod after CONTRACT-DAY0-002b.
	GoModRedisClient    = "github.com/redis/go-redis"
	GoModTelegramBotAPI = "github.com/go-telegram-bot-api/telegram-bot-api"
)

var SchemaNotificationChannels = []string{NotificationChannelTelegram}

var SchemaNotificationOutboxStatuses = []string{
	NotificationOutboxPending,
	NotificationOutboxSubmitted,
	NotificationOutboxDone,
	NotificationOutboxFailed,
}
