-- listing_search: active filter watches polled by worker.

CREATE TYPE listing_watch_status AS ENUM ('ACTIVE', 'PAUSED', 'DISABLED');

CREATE TABLE listing_filter_watches (
    id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id uuid NOT NULL REFERENCES users (id) ON DELETE CASCADE,
    filter_key text NOT NULL,
    query text NOT NULL,
    status listing_watch_status NOT NULL DEFAULT 'ACTIVE',
    last_polled_at timestamptz,
    created_at timestamptz NOT NULL DEFAULT now(),
    UNIQUE (user_id, filter_key)
);

CREATE INDEX listing_filter_watches_due_idx ON listing_filter_watches (status, last_polled_at);

CREATE TABLE listing_watch_hits (
    watch_id uuid NOT NULL REFERENCES listing_filter_watches (id) ON DELETE CASCADE,
    item_id uuid NOT NULL REFERENCES items (id) ON DELETE CASCADE,
    found_at timestamptz NOT NULL DEFAULT now(),
    PRIMARY KEY (watch_id, item_id)
);

CREATE INDEX listing_watch_hits_watch_id_idx ON listing_watch_hits (watch_id, found_at DESC);
