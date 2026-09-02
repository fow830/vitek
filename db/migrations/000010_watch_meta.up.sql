-- wave C: watch meta status from same fetch channel.

CREATE TYPE listing_watch_meta_status AS ENUM ('PENDING', 'READY', 'FAILED');

ALTER TABLE listing_filter_watches
    ADD COLUMN meta_status listing_watch_meta_status NOT NULL DEFAULT 'PENDING',
    ADD COLUMN meta_json jsonb NOT NULL DEFAULT '{}'::jsonb;
