-- name: UpsertNotificationSettings :one
INSERT INTO user_notification_settings (user_id, telegram_chat_id, enabled, updated_at)
VALUES ($1, $2, $3, now())
ON CONFLICT (user_id) DO UPDATE
SET telegram_chat_id = EXCLUDED.telegram_chat_id,
    enabled = EXCLUDED.enabled,
    updated_at = now()
RETURNING user_id, telegram_chat_id, enabled, updated_at;

-- name: GetNotificationSettings :one
SELECT user_id, telegram_chat_id, enabled, updated_at
FROM user_notification_settings
WHERE user_id = $1;

-- name: EnqueueNotification :exec
INSERT INTO notification_outbox (user_id, watch_id, item_id, channel, status)
VALUES ($1, $2, $3, 'TELEGRAM', 'PENDING')
ON CONFLICT (watch_id, item_id, channel) DO NOTHING;

-- name: ClaimPendingNotifications :many
SELECT
    o.id,
    o.user_id,
    o.watch_id,
    o.item_id,
    o.channel,
    o.status,
    o.attempts,
    o.last_error,
    o.created_at,
    i.title AS item_title,
    i.avito_id AS item_avito_id,
    s.telegram_chat_id,
    COALESCE(s.enabled, false) AS notify_enabled
FROM notification_outbox o
JOIN items i ON i.id = o.item_id
LEFT JOIN user_notification_settings s ON s.user_id = o.user_id
WHERE o.status IN ('PENDING', 'FAILED')
ORDER BY o.created_at ASC
LIMIT $1
FOR UPDATE OF o SKIP LOCKED;

-- name: MarkNotificationSubmitted :exec
UPDATE notification_outbox
SET status = 'SUBMITTED',
    attempts = attempts + 1
WHERE id = $1;

-- name: MarkNotificationDone :exec
UPDATE notification_outbox
SET status = 'DONE',
    last_error = NULL
WHERE id = $1;

-- name: MarkNotificationFailed :exec
UPDATE notification_outbox
SET status = 'FAILED',
    last_error = $2,
    attempts = attempts + 1
WHERE id = $1;
