-- wave D: notification outbox + user telegram settings.

CREATE TYPE notification_channel AS ENUM ('TELEGRAM');
CREATE TYPE notification_outbox_status AS ENUM ('PENDING', 'SUBMITTED', 'DONE', 'FAILED');

CREATE TABLE user_notification_settings (
    user_id uuid PRIMARY KEY REFERENCES users (id) ON DELETE CASCADE,
    telegram_chat_id text,
    enabled boolean NOT NULL DEFAULT false,
    updated_at timestamptz NOT NULL DEFAULT now()
);

CREATE TABLE notification_outbox (
    id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id uuid NOT NULL REFERENCES users (id) ON DELETE CASCADE,
    watch_id uuid NOT NULL REFERENCES listing_filter_watches (id) ON DELETE CASCADE,
    item_id uuid NOT NULL REFERENCES items (id) ON DELETE CASCADE,
    channel notification_channel NOT NULL DEFAULT 'TELEGRAM',
    status notification_outbox_status NOT NULL DEFAULT 'PENDING',
    attempts integer NOT NULL DEFAULT 0,
    last_error text,
    created_at timestamptz NOT NULL DEFAULT now(),
    UNIQUE (watch_id, item_id, channel)
);

CREATE INDEX notification_outbox_due_idx ON notification_outbox (status, created_at)
    WHERE status IN ('PENDING', 'FAILED');
