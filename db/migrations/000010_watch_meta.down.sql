ALTER TABLE listing_filter_watches
    DROP COLUMN IF EXISTS meta_json,
    DROP COLUMN IF EXISTS meta_status;

DROP TYPE IF EXISTS listing_watch_meta_status;
