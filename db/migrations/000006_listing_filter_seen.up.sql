-- listing_search: per-user filter baseline (seen avito_ids; first run seeds, no task_items).

CREATE TABLE listing_filter_seen (
    user_id uuid NOT NULL REFERENCES users (id) ON DELETE CASCADE,
    filter_key text NOT NULL,
    avito_id text NOT NULL,
    created_at timestamptz NOT NULL DEFAULT now(),
    PRIMARY KEY (user_id, filter_key, avito_id)
);

CREATE INDEX listing_filter_seen_user_filter_idx ON listing_filter_seen (user_id, filter_key);
